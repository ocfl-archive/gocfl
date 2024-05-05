//go:build !imagick

package extension

import (
	"bytes"
	"emperror.dev/errors"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/je4/gocfl/v2/pkg/subsystem/thumbnail"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"slices"
	"strings"
)

func NewThumbnail(config *ThumbnailConfig, mig *thumbnail.Thumbnail, logger zLogger.ZWrapper) (*Thumbnail, error) {
	sl := &Thumbnail{
		ThumbnailConfig: config,
		logger:          logger,
		thumbnail:       mig,
		buffer:          map[string]*bytes.Buffer{},
		counter:         map[string]int64{},
		streamInfo:      map[string]map[string]*ThumbnailResult{},
	}
	//	sl.writer = brotli.NewWriter(sl.buffer)
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	if mig != nil {
		sl.sourceFS = mig.SourceFS
	}
	return sl, nil
}

type Thumbnail struct {
	*ThumbnailConfig
	logger      zLogger.ZWrapper
	fsys        fs.FS
	lastHead    string
	thumbnail   *thumbnail.Thumbnail
	buffer      map[string]*bytes.Buffer
	writer      *brotli.Writer
	sourceFS    fs.FS
	currentHead string
	done        bool
	counter     map[string]int64
	streamInfo  map[string]map[string]*ThumbnailResult
}

func (thumb *Thumbnail) Terminate() error {
	return nil
}

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

	img, format, err := image.Decode(reader)
	if err != nil {
		thumb.logger.Infof("cannot decode image '%s': %v", stateFiles[0], err)
		return nil
	}
	rect := img.Bounds()
	thumb.logger.Infof("image format: %s, size: %d x %d", format, rect.Dx(), rect.Dy())

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

	fsys := object.GetFS()
	if fsys == nil {
		return errors.New("no filesystem set")
	}
	pr, pw := io.Pipe()
	done := make(chan bool, 1)
	go func() {
		defer pw.Close()
		var err error = nil
		switch thumb.ThumbnailConfig.Ext {
		case "png":
			err = png.Encode(pw, newImg)
		case "jpeg":
			err = jpeg.Encode(pw, newImg, nil)
		default:
			err = errors.Errorf("unsupported image format '%s'", thumb.ThumbnailConfig.Ext)
		}
		if err != nil {
			_ = pw.CloseWithError(errors.Wrap(err, "cannot encode image"))
		}
		done <- true
	}()

	thumb.counter[head]++
	targetFile, digest, err := thumb.storeThumbnail(object, head, pr)
	if err != nil {
		return errors.Wrap(err, "cannot store thumbnail")
	}
	<-done

	_ = digest
	thumb.logger.Infof("thumbnail stored: %s", targetFile)
	ml := &ThumbnailResult{
		//SourceDigest: cs,
		Filename:    targetFile,
		Ext:         thumb.ThumbnailConfig.Ext,
		Error:       "",
		ID:          "internal",
		ThumbDigest: digest,
	}
	thumb.streamInfo[head][infoName] = ml
	return nil
}
