//go:build vips && !imagick && cgo

package extension

import (
	"emperror.dev/errors"
	"fmt"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"io"
	"slices"
	"strings"
)

func (thumb *Thumbnail) StreamObject(object ocfl.Object, reader io.Reader, stateFiles []string, dest string) error {
	if len(stateFiles) == 0 {
		return errors.Errorf("no state files for object '%s'", object.GetID())
	}
	if !slices.Contains([]string{"png", "jpeg"}, strings.ToLower(thumb.ThumbnailConfig.Ext)) {
		thumb.logger.Info().Msgf("unsupported target image format '%s'", thumb.ThumbnailConfig.Ext)
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
		thumb.logger.Info().Msgf("thumbnail for '%s' already created", stateFiles[0])
		return nil
	}
	//ext := filepath.Ext(stateFiles[0])

	img, err := vips.NewImageFromReader(reader)
	if err != nil {
		thumb.logger.Info().Msgf("cannot decode image '%s': %v", stateFiles[0], err)
		return nil
	}
	defer img.Close()

	width := img.Width()
	height := img.Height()
	if width == 0 || height == 0 {
		thumb.logger.Info().Msgf("image '%s' has zero size", stateFiles[0])
		return nil
	}

	thumb.logger.Info().Msgf("image '%s' format: %s, size: %d x %d", stateFiles[0], img.Format().FileExt(), width, height)

	rectAspect := float64(width) / float64(height)
	thumbAspect := float64(thumb.Width) / float64(thumb.Height)
	newHeight := int(thumb.Height)
	newWidth := int(thumb.Width)
	_ = newWidth
	if rectAspect > thumbAspect {
		newHeight = (height * int(thumb.Width)) / width
	} else {
		newWidth = (width * int(thumb.Height)) / height
	}
	scale := float64(newHeight) / float64(height)
	if err := img.Resize(scale, vips.KernelLanczos3); err != nil {
		thumb.logger.Info().Msgf("cannot resize image '%s': %v", stateFiles[0], err)
		return nil
	}
	var imgBytes []byte
	var meta *vips.ImageMetadata
	switch thumb.ThumbnailConfig.Ext {
	case "png":
		imgBytes, meta, err = img.ExportPng(vips.NewPngExportParams())
	case "jpeg":
		imgBytes, meta, err = img.ExportJpeg(vips.NewJpegExportParams())
	default:
		thumb.logger.Info().Msgf("unsupported target image format '%s'", thumb.ThumbnailConfig.Ext)
		return nil
	}
	_ = meta

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

	thumb.logger.Info().Msgf("thumbnail stored: %s", targetFile)
	ml := &ThumbnailResult{
		//SourceDigest: cs,
		Filename:    targetFile,
		Ext:         thumb.ThumbnailConfig.Ext,
		Error:       "",
		ID:          "internal vips",
		ThumbDigest: digest,
	}
	thumb.streamInfo[head][infoName] = ml
	return nil
}
