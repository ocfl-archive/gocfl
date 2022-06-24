package main

import (
	"fmt"
	"github.com/goph/emperror"
	lm "github.com/je4/utils/v2/pkg/logger"
	flag "github.com/spf13/pflag"
	"gitlab.switch.ch/ub-unibas/gocfl/v2/pkg/checksum"
	"gitlab.switch.ch/ub-unibas/gocfl/v2/pkg/ocfl"
	"gitlab.switch.ch/ub-unibas/gocfl/v2/pkg/storagelayout"
	"gitlab.switch.ch/ub-unibas/gocfl/v2/pkg/zipfs"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//const LOGFORMAT = `%{time:2006-01-02T15:04:05.000} %{module}::%{shortfunc} [%{shortfile}] > %{level:.5s} - %{message}`
const LOGFORMAT = `%{time:2006-01-02T15:04:05.000} %{shortpkg}::%{longfunc} [%{shortfile}] > %{level:.5s} - %{message}`

func main() {
	var err error

	var zipfile = flag.String("file", "", "ocfl zip filename")
	var logfile = flag.String("logfile", "", "name of logfile")
	var loglevel = flag.String("loglevel", "DEBUG", "CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG")

	flag.Parse()

	logger, lf := lm.CreateLogger("ocfl", *logfile, nil, *loglevel, LOGFORMAT)
	defer lf.Close()

	var zipSize int64
	var zipReader *os.File
	var zipWriter *os.File

	tempFile := fmt.Sprintf("%s.tmp", *zipfile)
	if zipWriter, err = os.Create(tempFile); err != nil {
		err = emperror.ExposeStackTrace(emperror.Wrapf(err, "cannot create zip file %s", tempFile))
		stack, ok := emperror.StackTrace(err)
		if ok {
			log.Print(stack)
		}
		panic(err)
	}
	defer func() {
		zipWriter.Close()
		if err := os.Rename(fmt.Sprintf("%s.tmp", *zipfile), *zipfile); err != nil {
			log.Print(err)
		}
	}()

	stat, err := os.Stat(*zipfile)
	if err != nil {
		log.Print(emperror.Wrapf(err, "%s does not exist. creating new file", *zipfile))
	} else {
		zipSize = stat.Size()
		if zipReader, err = os.Open(*zipfile); err != nil {
			err = emperror.ExposeStackTrace(emperror.Wrapf(err, "cannot open zip file %s", *zipfile))
			stack, ok := emperror.StackTrace(err)
			if ok {
				log.Print(stack)
			}
			panic(err)
		}
		defer func() {
			zipReader.Close()
			if err := os.Rename(*zipfile, fmt.Sprintf("%s.%s", *zipfile, time.Now().Format("20060201_150405"))); err != nil {
				panic(err)
			}
		}()

	}

	zfs, err := zipfs.NewFSIO(zipReader, zipSize, zipWriter, logger)
	if err != nil {
		err = emperror.ExposeStackTrace(emperror.Wrap(err, "cannot create zipfs"))
		stack, ok := emperror.StackTrace(err)
		if ok {
			log.Print(stack)
		}
		panic(err)
	}
	defer zfs.Close()
	defaultStorageLayout, err := storagelayout.NewDefaultStorageLayout()
	if err != nil {
		panic(err)
	}

	storageRoot, err := ocfl.NewOCFLStorageRoot(zfs, defaultStorageLayout, logger)
	if err != nil {
		panic(err)
	}

	/*
		o, err := ocfl.NewOCFLObject(zfs, "", filepath.Base(*zipfile), logger)
		if err != nil {
			err = emperror.ExposeStackTrace(emperror.Wrap(err, "cannot create zipfs"))
			stack, ok := emperror.StackTrace(err)
			if ok {
				log.Print(stack)
			}
			panic(err)
		}
		defer o.Close()
	*/

	o, err := storageRoot.OpenObject("test042")
	if err != nil {
		err = emperror.ExposeStackTrace(emperror.Wrapf(err, "cannot open object %s", "test042"))
		stack, ok := emperror.StackTrace(err)
		if ok {
			log.Print(stack)
		}
		panic(err)
	}
	defer o.Close()

	if err := o.StartUpdate("test 42", "Jürgen Enge", "juergen.enge@unibas.ch"); err != nil {
		err = emperror.ExposeStackTrace(emperror.Wrap(err, "cannot add file"))
		stack, ok := emperror.StackTrace(err)
		if ok {
			log.Print(stack)
		}
		panic(err)
	}

	testdir := "C:/temp/bangbang/datatables"

	if err := filepath.Walk(testdir, func(path string, info fs.FileInfo, err error) error {
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
			err = emperror.ExposeStackTrace(emperror.Wrap(err, "cannot add file"))
			stack, ok := emperror.StackTrace(err)
			if ok {
				log.Print(stack)
			}
			panic(err)
		}
		if _, err := file.Seek(0, 0); err != nil {
			panic(err)
		}
		if err := o.AddFile(strings.Trim(strings.TrimPrefix(filepath.ToSlash(path), testdir), "/"), file, checksum); err != nil {
			err = emperror.ExposeStackTrace(emperror.Wrap(err, "cannot add file"))
			stack, ok := emperror.StackTrace(err)
			if ok {
				log.Print(stack)
			}
			panic(err)
		}
		return nil
	}); err != nil {
		panic(err)
	}

	o2, err := storageRoot.OpenObject("test041")
	if err != nil {
		err = emperror.ExposeStackTrace(emperror.Wrapf(err, "cannot open object %s", "test042"))
		stack, ok := emperror.StackTrace(err)
		if ok {
			log.Print(stack)
		}
		panic(err)
	}
	defer o2.Close()

	if err := o2.StartUpdate("test 41", "Jürgen Enge", "juergen.enge@unibas.ch"); err != nil {
		err = emperror.ExposeStackTrace(emperror.Wrap(err, "cannot add file"))
		stack, ok := emperror.StackTrace(err)
		if ok {
			log.Print(stack)
		}
		panic(err)
	}

	testdir2 := "C:/temp/bangbang/bootstrap"

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
			err = emperror.ExposeStackTrace(emperror.Wrap(err, "cannot add file"))
			stack, ok := emperror.StackTrace(err)
			if ok {
				log.Print(stack)
			}
			panic(err)
		}
		if _, err := file.Seek(0, 0); err != nil {
			panic(err)
		}
		if err := o2.AddFile(strings.Trim("x"+strings.TrimPrefix(filepath.ToSlash(path), testdir2), "/"), file, checksum); err != nil {
			err = emperror.ExposeStackTrace(emperror.Wrap(err, "cannot add file"))
			stack, ok := emperror.StackTrace(err)
			if ok {
				log.Print(stack)
			}
			panic(err)
		}
		return nil
	}); err != nil {
		panic(err)
	}
}
