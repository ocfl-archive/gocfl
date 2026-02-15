package util

import "io"

type StreamFS interface {
	Create(name string) (io.WriteCloser, error)
}
