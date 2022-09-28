package main

import (
	"emperror.dev/errors"
	"fmt"
	lm "github.com/je4/utils/v2/pkg/logger"
	flag "github.com/spf13/pflag"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension/storageroot"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"go.ub.unibas.ch/gocfl/v2/pkg/zipfs"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// const LOGFORMAT = `%{time:2006-01-02T15:04:05.000} %{module}::%{shortfunc} [%{shortfile}] > %{level:.5s} - %{message}`
const LOGFORMAT = `%{time:2006-01-02T15:04:05.000} %{shortpkg}::%{longfunc} [%{shortfile}] > %{level:.5s} - %{message}`
const VERSION = "1.0"

func main() {
	var panicking = true

	var err error

	var zipfile = flag.String("file", "", "ocfl zip filename")
	var srcdir = flag.String("source", "", "source folder")
	var logfile = flag.String("logfile", "", "name of logfile")
	var loglevel = flag.String("loglevel", "DEBUG", "CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG")

	flag.Parse()

	logger, lf := lm.CreateLogger("ocfl", *logfile, nil, *loglevel, LOGFORMAT)
	defer lf.Close()

	if *srcdir == "" {
		logger.Errorf("invalid source dir: %s", *srcdir)
		return
	}

	fi, err := os.Stat(*srcdir)
	if err != nil {
		logger.Errorf("cannot stat source dir %s: %v", *srcdir, err)
		return
	}
	if !fi.IsDir() {
		logger.Errorf("source dir %s is not a directory", *srcdir)
		return
	}

	var zipSize int64
	var zipReader *os.File
	var zipWriter *os.File

	tempFile := fmt.Sprintf("%s.tmp", *zipfile)
	if zipWriter, err = os.Create(tempFile); err != nil {
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		panic(err)
	}
	defer func() {
		zipWriter.Close()
		// only if no panic has happened
		if panicking {
			if err := os.Remove(fmt.Sprintf("%s.tmp", *zipfile)); err != nil {
				logger.Error(err)
			}
		} else {
			if err := os.Rename(fmt.Sprintf("%s.tmp", *zipfile), *zipfile); err != nil {
				logger.Error(err)
			}
		}
	}()

	stat, err := os.Stat(*zipfile)
	if err != nil {
		log.Print(errors.Wrapf(err, "%s does not exist. creating new file", *zipfile))
	} else {
		zipSize = stat.Size()
		if zipReader, err = os.Open(*zipfile); err != nil {
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			panic(err)
		}
		defer func() {
			zipReader.Close()
			// only if no panic has happened
			if !panicking {
				if err := os.Rename(*zipfile, fmt.Sprintf("%s.%s", *zipfile, time.Now().Format("20060201_150405"))); err != nil {
					logger.Error(err)
				}
			}
		}()

	}

	zfs, err := zipfs.NewFSIO(zipReader, zipSize, zipWriter, logger)
	if err != nil {
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		panic(err)
	}
	defer func() {
		if !panicking {
			zfs.Close()
		}

	}()
	defaultStorageLayout, err := storageroot.NewDefaultStorageLayout()
	if err != nil {
		panic(err)
	}

	storageRoot, err := ocfl.NewStorageRoot(zfs, VERSION, defaultStorageLayout, logger)
	if err != nil {
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		panic(err)
	}

	// TEST042
	o, err := storageRoot.OpenObject("test042")
	if err != nil {
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		panic(err)
	}

	if err := o.StartUpdate("test 42", "Jürgen Enge", "juergen.enge@unibas.ch"); err != nil {
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		panic(err)
	}
	defer func() {
		if !panicking {
			o.Close()
		}
	}()

	if err := filepath.Walk(*srcdir, func(path string, info fs.FileInfo, err error) error {
		// directory not interesting
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		checksum, err := checksum.Checksum(file, checksum.DigestSHA512)
		if err != nil {
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			panic(err)
		}
		if _, err := file.Seek(0, 0); err != nil {
			panic(err)
		}
		if err := o.AddFile(strings.Trim(strings.TrimPrefix(filepath.ToSlash(path), *srcdir), "/"), file, checksum); err != nil {
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			panic(err)
		}
		return nil
	}); err != nil {
		panic(err)
	}

	// TEST041
	o2, err := storageRoot.OpenObject("test041")
	if err != nil {
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		panic(err)
	}
	defer func() {
		if !panicking {
			o2.Close()
		}
	}()

	if err := o2.StartUpdate("test 41", "Jürgen Enge", "juergen.enge@unibas.ch"); err != nil {
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		panic(err)
	}

	testdir2 := "C:/temp/bangbang/img"

	if err := filepath.Walk(testdir2, func(path string, info fs.FileInfo, err error) error {
		// directory not interesting
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		checksum, err := checksum.Checksum(file, checksum.DigestSHA512)
		if err != nil {
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			panic(err)
		}
		if _, err := file.Seek(0, 0); err != nil {
			panic(err)
		}
		if err := o2.AddFile(strings.Trim("x"+strings.TrimPrefix(filepath.ToSlash(path), testdir2), "/"), file, checksum); err != nil {
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			panic(err)
		}
		return nil
	}); err != nil {
		panic(err)
	}

	panicking = false
}
