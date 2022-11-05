package main

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/op/go-logging"
	flag "github.com/spf13/pflag"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"go.ub.unibas.ch/gocfl/v2/pkg/osfs"
	"go.ub.unibas.ch/gocfl/v2/pkg/zipfs"
	"log"
	"os"
	"strings"
)

// const LOGFORMAT = `%{time:2006-01-02T15:04:05.000} %{module}::%{shortfunc} [%{shortfile}] > %{level:.5s} - %{message}`
const LOGFORMAT = `%{time:2006-01-02T15:04:05.000} %{shortpkg}::%{longfunc} [%{shortfile}] > %{level:.5s} - %{message}`
const VERSION = "1.0"

func showStatus(ctx context.Context) error {
	status, err := ocfl.GetValidationStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot get status of validation")
	}
	status.Compact()
	for _, err := range status.Errors {
		fmt.Println(err)
		//logger.Infof("ERROR: %v", err)
	}
	for _, err := range status.Warnings {
		fmt.Println(err)
		//logger.Infof("WARN:  %v", err)
	}
	fmt.Println("\n")
	return nil
}

func checkObject(dest ocfl.OCFLFS, extensionFactory *ocfl.ExtensionFactory, logger *logging.Logger) error {
	ctx := ocfl.NewContextValidation(context.TODO())
	defer showStatus(ctx)
	object, err := ocfl.NewObject(ctx, dest, "", "", logger)
	if err != nil {
		return errors.Wrap(err, "cannot load object")
	}
	if err := object.Check(); err != nil {
		return errors.Wrapf(err, "check of %s failed", object.GetID())
	}
	return nil
}

func check(dest ocfl.OCFLFS, extensionFactory *ocfl.ExtensionFactory, logger *logging.Logger) error {
	defaultStorageLayout, err := ocfl.NewDefaultStorageRootExtension()
	if err != nil {
		panic(err)
	}

	storageRoot, err := ocfl.NewStorageRoot(ocfl.NewContextValidation(context.TODO()), dest, VERSION, defaultStorageLayout, extensionFactory, logger)
	if err != nil {
		return errors.Wrap(err, "cannot create new storageroot")
	}
	if err := storageRoot.Check(); err != nil {
		return errors.Wrap(err, "ocfl not valid")
	}
	return nil
}

func ingest(dest ocfl.OCFLFS, srcdir string, extensionFactory *ocfl.ExtensionFactory, logger *logging.Logger) error {

	if srcdir == "" {
		return errors.Errorf("invalid source dir: %s", srcdir)
	}

	fi, err := os.Stat(srcdir)
	if err != nil {
		return errors.Wrapf(err, "cannot stat source dir %s", srcdir)
	}
	if !fi.IsDir() {
		return errors.Errorf("source dir %s is not a directory", srcdir)
	}

	defaultStorageLayout, err := ocfl.NewDefaultStorageRootExtension()
	if err != nil {
		panic(err)
	}

	storageRoot, err := ocfl.NewStorageRoot(ocfl.NewContextValidation(context.TODO()), dest, VERSION, defaultStorageLayout, extensionFactory, logger)
	if err != nil {
		return errors.Wrap(err, "cannot create new storageroot")
	}

	// TEST042
	o, err := storageRoot.OpenObject("test042")
	if err != nil {
		return errors.Wrapf(err, "cannot create object %s", "test042")
	}

	if err := o.StartUpdate("test 42", "Jürgen Enge", "juergen.enge@unibas.ch"); err != nil {
		return errors.Wrapf(err, "cannot start update for object %s", "test 42")
	}

	if err := o.AddFolder(os.DirFS(srcdir)); err != nil {
		panic(err)
	}

	if err := o.Close(); err != nil {
		return errors.Wrapf(err, "cannot close object %s", "test042")
	}

	return nil
}

var target = flag.String("target", "", "ocfl zip or folder")
var checkFlag = flag.Bool("check", false, "only check file")
var checkObjectFlag = flag.Bool("checkobject", false, "only check object structure file")
var srcDir = flag.String("source", "", "source folder")
var logfile = flag.String("logfile", "", "name of logfile")
var loglevel = flag.String("loglevel", "DEBUG", "CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG")

func main() {

	var err error

	//	var version = flag.String("version", "", "ocfl version")

	flag.Parse()

	logger, lf := lm.CreateLogger("ocfl", *logfile, nil, *loglevel, LOGFORMAT)
	defer lf.Close()

	extensionFactory, err := ocfl.NewExtensionFactory()
	if err != nil {
		logger.Errorf("cannot instantiate extension factory: %v", err)
		return
	}

	extensionFactory.AddCreator(ocfl.StorageLayoutDirectCleanName, func(config []byte) (ocfl.Extension, error) {
		return ocfl.NewStorageLayoutDirectClean(config)
	})

	extensionFactory.AddCreator(ocfl.StorageLayoutFlatDirectName, func(config []byte) (ocfl.Extension, error) {
		return ocfl.NewStorageLayoutFlatDirect(config)
	})

	var ocfs ocfl.OCFLFS

	var zipSize int64
	var zipReader *os.File
	var zipWriter *os.File

	var zipFile string
	var objectPath string
	if strings.HasSuffix(strings.ToLower(*target), ".zip") {
		zipFile = *target
	} else {
		if pos := strings.Index(*target, ".zip/"); pos != -1 {
			zipFile = (*target)[0 : pos+4]
			objectPath = (*target)[pos+4:]
		}
	}
	if zipFile != "" {
		stat, err := os.Stat(zipFile)
		if err != nil {
			log.Print(errors.Wrapf(err, "%s does not exist. creating new file", zipFile))
		} else {
			zipSize = stat.Size()
			if zipReader, err = os.Open(zipFile); err != nil {
				logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
				panic(err)
			}

		}

		if *srcDir != "" {
			tempFile := fmt.Sprintf("%s.tmp", *target)
			if zipWriter, err = os.Create(tempFile); err != nil {
				logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
				panic(err)
			}
		}
		ocfs, err = zipfs.NewFSIO(zipReader, zipSize, zipWriter, logger)
		if err != nil {
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			panic(err)
		}
	} else {
		ocfs, err = osfs.NewFSIO(*target, logger)
		if err != nil {
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			panic(err)
		}
	}

	// do stuff here...
	switch {
	case *srcDir != "":
		if err := ingest(ocfs, *srcDir, extensionFactory, logger); err != nil {
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			panic(err)
		}
	case *checkObjectFlag:
		objfs := ocfs.SubFS(objectPath)
		if err := checkObject(objfs, extensionFactory, logger); err != nil {
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			panic(err)
		}
	case *checkFlag:
		if err := check(ocfs, extensionFactory, logger); err != nil {
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			panic(err)
		}
	}

	if err := ocfs.Close(); err != nil {
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		panic(err)
	}
	if zipWriter != nil {
		if err := zipWriter.Close(); err != nil {
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			panic(err)
		}
	}
	if zipReader != nil {
		if err := zipReader.Close(); err != nil {
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			panic(err)
		}
		if zipWriter != nil {
			if err := os.Rename(fmt.Sprintf("%s.tmp", *target), *target); err != nil {
				logger.Error(err)
				panic(err)
			}
		}
	}
}
