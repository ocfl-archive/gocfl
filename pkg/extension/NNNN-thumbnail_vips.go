//go:build vips && !imagick && cgo

package extension

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"emperror.dev/errors"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
)

func (thumb *Thumbnail) StreamObject(object ocfl.Object, reader io.Reader, stateFiles []string, dest string) error {
	if len(stateFiles) == 0 {
		return errors.Errorf("no state files for object '%s'", object.GetID())
	}
	if !slices.Contains([]string{"png", "jpeg"}, strings.ToLower(thumb.ThumbnailConfig.Ext)) {
		thumb.logger.Info().Any(
			thumb.errorFactory.LogError(
				ErrorThumbnailExtension,
				fmt.Sprintf("unsupported target image format '%s'", thumb.ThumbnailConfig.Ext),
				nil,
			),
		).Msg("")
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
		thumb.logger.Info().Any(
			thumb.errorFactory.LogError(
				LogThumbnailExtension,
				fmt.Sprintf("thumbnail for '%s' already created", stateFiles[0]),
				nil,
			),
		).Msg("")
		return nil
	}
	img, err := vips.NewImageFromReader(reader)
	if err != nil {
		thumb.logger.Info().Any(
			errorTopic,
			thumb.errorFactory.NewError(
				ErrorThumbnailExtension,
				fmt.Sprintf("cannot decode image '%s'", stateFiles[0]),
				err,
			),
		).Msg("")
		return nil
	}
	defer img.Close()

	width := img.Width()
	height := img.Height()
	if width == 0 || height == 0 {
		thumb.logger.Info().Any(
			thumb.errorFactory.LogError(
				ErrorThumbnailExtension,
				fmt.Sprintf("image '%s' has zero size", stateFiles[0]),
				nil,
			),
		).Msg("")
		return nil
	}
	thumb.logger.Info().Any(
		thumb.errorFactory.LogError(
			LogThumbnailExtension,
			fmt.Sprintf("image '%s' format: %s, size: %d x %d", stateFiles[0], img.Format().FileExt(), width, height),
			nil,
		),
	).Msg("")
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
		thumb.logger.Info().Any(
			errorTopic,
			thumb.errorFactory.NewError(
				ErrorThumbnailExtension,
				fmt.Sprintf("cannot resize image '%s'", stateFiles[0]),
				err,
			),
		).Msg("")
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
		thumb.logger.Info().Any(
			thumb.errorFactory.LogError(
				ErrorThumbnailExtension,
				fmt.Sprintf("unsupported target image format '%s'", thumb.ThumbnailConfig.Ext),
				nil,
			),
		).Msg("")
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
	thumb.logger.Info().Any(
		thumb.errorFactory.LogError(
			LogThumbnailExtension,
			fmt.Sprintf("thumbnail stored: %s", targetFile),
			nil,
		),
	).Msg("")
	ml := &ThumbnailResult{
		Filename:    targetFile,
		Ext:         thumb.ThumbnailConfig.Ext,
		Error:       "",
		ID:          "internal vips",
		ThumbDigest: digest,
	}
	thumb.streamInfo[head][infoName] = ml
	return nil
}
