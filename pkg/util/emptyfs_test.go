package util

import (
	"errors"
	"io/fs"
	"os"
	"testing"
)

func TestEmptyFS_Open(t *testing.T) {
	fsys := EmptyFS{}

	t.Run("valid path", func(t *testing.T) {
		_, err := fsys.Open("anyfile.txt")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if pathErr, ok := errors.AsType[*fs.PathError](err); ok {
			if !errors.Is(pathErr.Err, os.ErrNotExist) {
				t.Errorf("expected os.ErrNotExist, got %v", pathErr.Err)
			}
		} else {
			t.Fatalf("expected fs.PathError, got %T", err)
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		_, err := fsys.Open("/absolute/path")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if pathErr, ok := errors.AsType[*fs.PathError](err); ok {
			if !errors.Is(pathErr.Err, fs.ErrInvalid) {
				t.Errorf("expected fs.ErrInvalid, got %v", pathErr.Err)
			}
		} else {
			t.Fatalf("expected fs.PathError, got %T", err)
		}
	})
}
