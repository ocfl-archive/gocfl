//go:build (!(imagick && !vips) && !(!imagick && vips)) || !cgo

package extension

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"slices"
	"strings"

	"emperror.dev/errors"
	"github.com/nfnt/resize"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/vp8"
	_ "golang.org/x/image/vp8l"
	_ "golang.org/x/image/webp"
)

func (thumb *Thumbnail) StreamObject(object object.VersionWriter, reader io.Reader, stateFiles []string, dest string) error {
	if len(stateFiles) == 0 {
		return errors.Errorf("no state files for object '%s'", object.GetID())
	}
	if !slices.Contains([]string{"png", "jpeg"}, strings.ToLower(thumb.ThumbnailConfig.Ext)) {
		thumb.logger.Info().Msgf("unsupported target image format '%s'", thumb.ThumbnailConfig.Ext)
		return nil
	}
	inventory := object.GetInventory()
	head := inventory.GetHead()
	if _, ok := thumb.counter[head.String()]; !ok {
		thumb.counter[head.String()] = 0
	}
	if _, ok := thumb.streamInfo[head.String()]; !ok {
		thumb.streamInfo[head.String()] = map[string]*ThumbnailResult{}
	}
	infoName := fmt.Sprintf("%s/content/%s", head, stateFiles[0])
	if _, ok := thumb.streamInfo[head.String()][infoName]; ok {
		thumb.logger.Info().Msgf("thumbnail for '%s' already created", stateFiles[0])
		return nil
	}
	//ext := filepath.Ext(stateFiles[0])

	img, format, err := image.Decode(reader)
	if err != nil {
		thumb.logger.Info().Msgf("cannot decode image '%s': %v", stateFiles[0], err)
		return nil
	}
	rect := img.Bounds()
	dx := rect.Dx()
	dy := rect.Dy()
	thumb.logger.Info().Msgf("image format: %s, size: %d x %d", format, dx, dy)

	if dx == 0 || dy == 0 {
		thumb.logger.Info().Msgf("image '%s' has no size", stateFiles[0])
		return nil
	}

	rectAspect := rect.Dx() / rect.Dy()
	thumbAspect := int(thumb.Width) / int(thumb.Height)
	newHeight := uint(thumb.Height)
	newWidth := uint(thumb.Width)
	if rectAspect > thumbAspect {
		newHeight = uint(rect.Dy() * int(thumb.Width) / rect.Dx())
	} else {
		newWidth = uint(rect.Dx() * int(thumb.Height) / rect.Dy())
	}

	newImg := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
	img = nil // free memory
	if newImg == nil {
		return errors.Errorf("cannot resize image '%s'", stateFiles[0])
	}
	if _, ok := thumb.streamImg[head.String()]; !ok {
		thumb.streamImg[head.String()] = map[string]image.Image{}
	}
	thumb.streamImg[head.String()][infoName] = newImg
	/*
		fsys := object.GetFS()
		if fsys == nil {
			return errors.New("no filesystem set")
		}
		pr, pw := io.Pipe()
		done := make(chan bool, 1)
		go func() {
			var err error = nil
			defer func() {
				if err == nil {
					if err := pw.Close(); err != nil {
						thumb.logger.Error().Err(err).Msg("cannot close pipe")
					}
				} else {
					if err := pw.CloseWithError(errors.Wrap(err, "cannot encode image")); err != nil {
						thumb.logger.Error().Err(err).Msg("cannot close pipe")
					}
				}
				done <- true
			}()
			switch thumb.ThumbnailConfig.Ext {
			case "png":
				err = png.Encode(pw, newImg)
			case "jpeg":
				err = jpeg.Encode(pw, newImg, nil)
			default:
				err = errors.Errorf("unsupported image format '%s'", thumb.ThumbnailConfig.Ext)
			}
		}()

		thumb.counter[head.String()]++
		targetFile, digest, err := thumb.storeThumbnail(object, head, pr)
		if err != nil {
			return errors.Wrap(err, "cannot store thumbnail")
		}
		<-done

		_ = digest
		thumb.logger.Info().Msgf("thumbnail stored: %s", targetFile)
		ml := &ThumbnailResult{
			//SourceDigest: cs,
			Filename:    targetFile,
			Ext:         thumb.ThumbnailConfig.Ext,
			Error:       "",
			ID:          "internal",
			ThumbDigest: digest,
		}
		thumb.streamInfo[head.String()][infoName] = ml

	*/
	return nil
}

func (thumb *Thumbnail) AddFileAfter(versionWriter object.VersionWriter, sourceFS fs.FS, source []string, internalPath, digest, area string, isDir bool) error {
	inventory := versionWriter.GetInventory()
	head := inventory.GetHead()
	if _, ok := thumb.counter[head.String()]; !ok {
		thumb.counter[head.String()] = 0
	}
	infoName := internalPath // fmt.Sprintf("%s/content/%s", head, dest)
	if _, ok := thumb.streamInfo[head.String()]; !ok {
		return nil
	}
	newImg, ok := thumb.streamImg[head.String()][infoName]
	if !ok {
		return nil
	}
	defer func() {
		delete(thumb.streamImg[head.String()], infoName)
		newImg = nil
	}()
	fsys := versionWriter.GetFS()
	if fsys == nil {
		return errors.New("no filesystem set")
	}
	pr, pw := io.Pipe()
	done := make(chan bool, 1)
	go func() {
		var err error = nil
		defer func() {
			if err == nil {
				if err := pw.Close(); err != nil {
					thumb.logger.Error().Err(err).Msg("cannot close pipe")
				}
			} else {
				if err := pw.CloseWithError(errors.Wrap(err, "cannot encode image")); err != nil {
					thumb.logger.Error().Err(err).Msg("cannot close pipe")
				}
			}
			done <- true
		}()
		switch thumb.ThumbnailConfig.Ext {
		case "png":
			err = png.Encode(pw, newImg)
		case "jpeg":
			err = jpeg.Encode(pw, newImg, nil)
		default:
			err = errors.Errorf("unsupported image format '%s'", thumb.ThumbnailConfig.Ext)
		}
	}()

	thumb.counter[head.String()]++
	targetFile, digest, err := thumb.storeThumbnail(versionWriter, head, pr)
	if err != nil {
		return errors.Wrap(err, "cannot store thumbnail")
	}
	thumb.logger.Info().Msgf("thumbnail stored: %s", targetFile)
	ml := &ThumbnailResult{
		//SourceDigest: cs,
		Filename:    targetFile,
		Ext:         thumb.ThumbnailConfig.Ext,
		Error:       "",
		ID:          "internal",
		ThumbDigest: digest,
	}
	thumb.streamInfo[head.String()][infoName] = ml
	return nil
}

func (thumb *Thumbnail) AddFileBefore(object object.VersionWriter, sourceFS fs.FS, source string, dest string, area string, isDir bool) error {
	return nil
}

func (thumb *Thumbnail) UpdateFileBefore(object object.VersionWriter, sourceFS fs.FS, source, dest, area string, isDir bool) error {
	return nil
}

func (thumb *Thumbnail) DeleteFileBefore(versionWriter object.VersionWriter, dest string, area string) error {
	return nil
}

func (thumb *Thumbnail) UpdateFileAfter(object object.VersionWriter, sourceFS fs.FS, source, area string, isDir bool) error {
	return nil
}

func (thumb *Thumbnail) DeleteFileAfter(object object.VersionWriter, dest string, area string) error {
	return nil
}

var _ object.ExtensionContentChange = &Thumbnail{}
