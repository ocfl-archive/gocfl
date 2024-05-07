//go:build imagick && !vips && cgo

package extension

import (
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"gopkg.in/gographics/imagick.v3/imagick"
	"io"
	"slices"
	"strings"
)

func (thumb *Thumbnail) StreamObject(object ocfl.Object, reader io.Reader, stateFiles []string, dest string) error {
	if len(stateFiles) == 0 {
		return errors.Errorf("no state files for object '%s'", object.GetID())
	}
	if !slices.Contains([]string{"png", "jpeg"}, strings.ToLower(thumb.ThumbnailConfig.Ext)) {
		thumb.logger.Infof("unsupported target image format '%s'", thumb.ThumbnailConfig.Ext)
		return nil
	}
	inventory := object.GetInventory()
	head := inventory.GetHead()
	if _, ok := thumb.counter[head]; !ok {
		thumb.counter[head] = 0
	}
	if _, ok := thumb.streamInfo[head]; !ok {
		thumb.streamInfo[head] = map[string]*ThumbnailResult{}
	}
	infoName := fmt.Sprintf("%s/content/%s", head, stateFiles[0])
	if _, ok := thumb.streamInfo[head][infoName]; ok {
		thumb.logger.Infof("thumbnail for '%s' already created", stateFiles[0])
		return nil
	}
	//ext := filepath.Ext(stateFiles[0])

	imgBytes, err := io.ReadAll(reader)
	if err != nil {
		return errors.Wrap(err, "cannot read image")
	}

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if err := mw.ReadImageBlob(imgBytes); err != nil {
		thumb.logger.Infof("cannot decode image '%s': %v", stateFiles[0], err)
		return nil
	}
	imgBytes = nil // free memory
	width, height, err := mw.GetSize()
	if err != nil {
		thumb.logger.Infof("cannot get image size of '%s': %v", stateFiles[0], err)
		return nil
	}
	if width == 0 || height == 0 {
		thumb.logger.Infof("image '%s' has zero size", stateFiles[0])
		return nil
	}
	thumb.logger.Infof("image '%s' format: %s, size: %d x %d", stateFiles[0], mw.GetFormat(), width, height)

	rectAspect := width / height
	thumbAspect := uint(thumb.Width) / uint(thumb.Height)
	newHeight := uint(thumb.Height)
	newWidth := uint(thumb.Width)
	if rectAspect > thumbAspect {
		newHeight = uint(height * uint(thumb.Width) / width)
	} else {
		newWidth = uint(width * uint(thumb.Height) / height)
	}

	if err := mw.ResizeImage(newWidth, newHeight, imagick.FILTER_LANCZOS); err != nil {
		thumb.logger.Infof("cannot resize image '%s': %v", stateFiles[0], err)
		return nil
	}

	if err := mw.SetImageFormat(thumb.ThumbnailConfig.Ext); err != nil {
		thumb.logger.Infof("cannot set image '%s' format '%s': %v", stateFiles[0], thumb.ThumbnailConfig.Ext, err)
		return nil
	}
	mw.ResetIterator()
	imgBytes = mw.GetImageBlob()

	fsys := object.GetFS()
	if fsys == nil {
		return errors.New("no filesystem set")
	}

	thumb.counter[head]++
	targetFile, digest, err := thumb.storeThumbnail(object, head, io.NopCloser(strings.NewReader(string(imgBytes))))
	if err != nil {
		return errors.Wrap(err, "cannot store thumbnail")
	}
	imgBytes = nil // free memory

	thumb.logger.Infof("thumbnail stored: %s", targetFile)
	ml := &ThumbnailResult{
		//SourceDigest: cs,
		Filename:    targetFile,
		Ext:         thumb.ThumbnailConfig.Ext,
		Error:       "",
		ID:          "internal imagick",
		ThumbDigest: digest,
	}
	thumb.streamInfo[head][infoName] = ml
	return nil
}
