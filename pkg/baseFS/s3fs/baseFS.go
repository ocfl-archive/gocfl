package s3fs

import (
	"go.ub.unibas.ch/gocfl/v2/pkg/baseFS"
	"strings"
)

type BaseFS struct {
	endpoint, accessKeyID, secretAccessKey, bucket, region string
	useSSL                                                 bool
}

func NewBaseFS(endpoint, accessKeyID, secretAccessKey, bucket, region string, useSSL bool) (baseFS.BaseFS, error) {
	bFS := &BaseFS{
		endpoint:        endpoint,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		bucket:          bucket,
		region:          region,
		useSSL:          useSSL,
	}
	return bFS, nil
}

func (b BaseFS) Valid(path string) bool {
	return strings.HasPrefix(path, "bucket:")
}

var (
	_ baseFS.BaseFS = BaseFS{}
)
