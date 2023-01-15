package indexer

import (
	"emperror.dev/errors"
	ironmaiden "github.com/je4/indexer/pkg/indexer"
	"github.com/spf13/viper"
	"strconv"
)

type Siegfried struct {
	Signature string
	MimeMap   map[string]string
}

func GetSiegfried() (*Siegfried, error) {
	var sf = &Siegfried{
		Signature: viper.GetString("Indexer.Siegfried.Signature"),
		MimeMap:   viper.GetStringMapString("Indexer.Siegfried.MimeMap"),
	}
	return sf, nil
}

type FFMPEG struct {
	FFProbe string
	WSL     bool
	Timeout string
	Online  bool
	Enabled bool
	Mime    []ironmaiden.FFMPEGMime
}

func GetFFMPEG() (*FFMPEG, error) {
	ffmpeg := &FFMPEG{
		FFProbe: viper.GetString("Indexer.FFMPEG.FFProbe"),
		WSL:     viper.GetBool("Indexer.FFMPEG.WSL"),
		Timeout: viper.GetString("Indexer.FFMPEG.Timeout"),
		Online:  viper.GetBool("Indexer.FFMPEG.Online"),
		Enabled: viper.GetBool("Indexer.FFMPEG.Enabled"),
		Mime:    []ironmaiden.FFMPEGMime{},
	}
	mimesInterface := viper.Get("Indexer.FFMPEG.Mime")
	mimeSlice, ok := mimesInterface.([]any)
	if !ok {
		return nil, errors.Errorf("Indexer.FFMPEG.Mime not a slice %v", mimesInterface)
	}
	for key, mimeInterface := range mimeSlice {
		var m = ironmaiden.FFMPEGMime{}
		mimeMap, ok := mimeInterface.(map[string]any)
		if !ok {
			return nil, errors.Errorf("Indexer.FFMPEG.Mime.%v not a map %v", key, mimeInterface)
		}
		audioInt, ok := mimeMap["audio"]
		if !ok {
			return nil, errors.Errorf("Indexer.FFMPEG.Mime.%v.%s does not exist in %v", key, "audio", mimeMap)
		}
		m.Audio, ok = audioInt.(bool)
		if !ok {
			return nil, errors.Errorf("Indexer.FFMPEG.Mime.%v.%s is not bool", key, "audio", audioInt)
		}
		videoInt, ok := mimeMap["video"]
		if !ok {
			return nil, errors.Errorf("Indexer.FFMPEG.Mime.%v.%s does not exist in %v", key, "video", mimeMap)
		}
		m.Video, ok = videoInt.(bool)
		if !ok {
			return nil, errors.Errorf("Indexer.FFMPEG.Mime.%v.%s is not bool", key, "video", videoInt)
		}
		formatInt, ok := mimeMap["format"]
		if !ok {
			return nil, errors.Errorf("Indexer.FFMPEG.Mime.%v.%s does not exist in %v", key, "format", mimeMap)
		}
		m.Format, ok = formatInt.(string)
		if !ok {
			return nil, errors.Errorf("Indexer.FFMPEG.Mime.%v.%s is not string", key, "format", formatInt)
		}
		mimeInt, ok := mimeMap["mime"]
		if !ok {
			return nil, errors.Errorf("Indexer.FFMPEG.Mime.%v.%s does not exist in %v", key, "mime", mimeMap)
		}
		m.Mime, ok = mimeInt.(string)
		if !ok {
			return nil, errors.Errorf("Indexer.FFMPEG.Mime.%v.%s is not string", key, "mime", mimeInt)
		}
		ffmpeg.Mime = append(ffmpeg.Mime, m)
	}
	return ffmpeg, nil
}

func stringMapToMimeRelevance(mimeRelevanceInterface map[string]any) (map[int]ironmaiden.MimeWeightString, error) {
	var mimeRelevance = map[int]ironmaiden.MimeWeightString{}
	for keyStr, valInterface := range mimeRelevanceInterface {
		key, err := strconv.Atoi(keyStr)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid key entry '%s' in 'MimeRelevance'", keyStr)
		}
		val, ok := valInterface.(map[string]any)
		if !ok {
			return nil, errors.Wrapf(err, "invalid value %v for key '%s' in 'MimeRelevance' map", val, keyStr)
		}
		regexpInterface, ok := val["regexp"]
		if !ok {
			return nil, errors.Wrapf(err, "invalid value %v for key '%s' in 'MimeRelevance' regexp", val, keyStr)
		}
		regexp, ok := regexpInterface.(string)
		if !ok {
			return nil, errors.Wrapf(err, "invalid value %v for key '%s' in 'MimeRelevance' regexp", val, keyStr)
		}
		weightInterface, ok := val["weight"]
		if !ok {
			return nil, errors.Wrapf(err, "invalid value %v for key '%s' in 'MimeRelevance' weight", val, keyStr)
		}
		weight, ok := weightInterface.(int64)
		if !ok {
			return nil, errors.Wrapf(err, "invalid value %v for key '%s' in 'MimeRelevance' regexp", val, keyStr)
		}
		mimeRelevance[key] = ironmaiden.MimeWeightString{
			Regexp: regexp,
			Weight: int(weight),
		}
	}
	return mimeRelevance, nil
}

func GetMimeRelevance() (map[int]ironmaiden.MimeWeightString, error) {
	mimeRelevanceInterface := viper.GetStringMap("Indexer.MimeRelevance")
	return stringMapToMimeRelevance(mimeRelevanceInterface)
}

type ImageMagick struct {
	Identify string
	Convert  string
	WSL      bool
	Timeout  string
	Online   bool
	Enabled  bool
}

func GetImageMagick() (*ImageMagick, error) {
	im := &ImageMagick{
		Identify: viper.GetString("Indexer.ImageMagick.Identify"),
		Convert:  viper.GetString("Indexer.ImageMagick.Convert"),
		WSL:      viper.GetBool("Indexer.ImageMagick.WSL"),
		Timeout:  viper.GetString("Indexer.ImageMagick.Timeout"),
		Online:   viper.GetBool("Indexer.ImageMagick.Online"),
		Enabled:  viper.GetBool("Indexer.ImageMagick.Enabled"),
	}
	return im, nil
}

type Tika struct {
	Address    string
	RegexpMime string
	Timeout    string
	Online     bool
	Enabled    bool
}

func GetTika() (*Tika, error) {
	im := &Tika{
		Address:    viper.GetString("Indexer.Tika.Address"),
		RegexpMime: viper.GetString("Indexer.Tika.RegexpMime"),
		Timeout:    viper.GetString("Indexer.Tika.Timeout"),
		Online:     viper.GetBool("Indexer.Tika.Online"),
		Enabled:    viper.GetBool("Indexer.Tika.Enabled"),
	}
	return im, nil
}
