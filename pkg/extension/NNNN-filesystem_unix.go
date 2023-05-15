//go:build !windows && !plan9

package extension

import (
	"emperror.dev/errors"
	"io/fs"
	"os"
	"runtime"
	"syscall"
	"time"
)

func (fsm *FilesystemMeta) init(fullpath string, fileInfo fs.FileInfo) error {
	fsm.OS = runtime.GOOS
	sys := fileInfo.Sys()
	if sys == nil {
		return errors.New("fileInfo.Sys() is nil")
	}
	stat_t, ok := sys.(*syscall.Stat_t)
	if !ok {
		return errors.New("fileInfo.Sys() is not *syscall.Stat_t")
	}
	fsm.ATime = time.Unix(stat_t.Atim.Sec, stat_t.Atim.Nsec)
	fsm.CTime = time.Unix(stat_t.Ctim.Sec, stat_t.Ctim.Nsec)
	fsm.MTime = time.Unix(stat_t.Mtim.Sec, stat_t.Mtim.Nsec)
	fsm.Size = uint64(stat_t.Size)
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
	unixPerms := fi.Mode() & os.ModePerm
	fsm.Attr = unixPerms.String()
	fsm.SystemStat = stat_t

	return nil
}
