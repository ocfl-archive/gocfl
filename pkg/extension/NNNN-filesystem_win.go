//go:build windows

package extension

import (
	"emperror.dev/errors"
	"github.com/gomiran/volmgmt/fileattr"
	"io/fs"
	"os"
	"runtime"
	"syscall"
	"time"
)

func (fsm *filesystemMeta) init(fullpath string, fileInfo fs.FileInfo) error {
	fsm.OS = runtime.GOOS
	sys := fileInfo.Sys()
	if sys == nil {
		return errors.New("fileInfo.Sys() is nil")
	}
	win32FileAttributeData, ok := sys.(*syscall.Win32FileAttributeData)
	if !ok {
		return errors.New("fileInfo.Sys() is not *syscall.Win32FileAttributeData")
	}
	fsm.CTime = time.Unix(0, win32FileAttributeData.CreationTime.Nanoseconds())
	fsm.MTime = time.Unix(0, win32FileAttributeData.LastWriteTime.Nanoseconds())
	fsm.ATime = time.Unix(0, win32FileAttributeData.LastAccessTime.Nanoseconds())
	fsm.Size = uint64(win32FileAttributeData.FileSizeLow)

	attr := fileattr.Value(win32FileAttributeData.FileAttributes)
	fsm.Attr = attr.String()

	fi, err := os.Lstat(fullpath)
	if err != nil {
		return errors.WithStack(err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		fsm.Symlink, err = os.Readlink(fullpath)
		if err != nil {
			return errors.Wrapf(err, "cannot read Symlink %s", fullpath)
		}
	}

	fsm.SystemStat = win32FileAttributeData

	return nil
}
