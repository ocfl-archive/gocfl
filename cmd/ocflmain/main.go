package main

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/op/go-logging"
	flag "github.com/spf13/pflag"
	"go.ub.unibas.ch/gocfl/v2/data/defaultextensions"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension"
	"go.ub.unibas.ch/gocfl/v2/pkg/genericfs"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"go.ub.unibas.ch/gocfl/v2/pkg/osfs"
	"go.ub.unibas.ch/gocfl/v2/pkg/zipfs"
	"io/fs"
	"log"
	"os"
	"strings"
)

// const LOGFORMAT = `%{time:2006-01-02T15:04:05.000} %{module}::%{shortfunc} [%{shortfile}] > %{level:.5s} - %{message}`
const LOGFORMAT = `%{time:2006-01-02T15:04:05.000} %{shortpkg}::%{longfunc} [%{shortfile}] > %{level:.5s} - %{message}`
const VERSION = "1.1"

func showStatus(ctx context.Context) error {
	status, err := ocfl.GetValidationStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot get status of validation")
	}
	status.Compact()
	context := ""
	for _, err := range status.Errors {
		if err.Context != context {
			fmt.Printf("\n[%s]\n", err.Context)
			context = err.Context
		}
		fmt.Printf("   Validation Error #%s - %s [%s]\n", err.Code, err.Description, err.Description2)
		//logger.Infof("ERROR: %v", err)
	}
	for _, err := range status.Warnings {
		if err.Context != context {
			fmt.Printf("\n[%s]\n", err.Context)
			context = err.Context
		}
		fmt.Printf("   Validation Warning #%s - %s [%s]\n", err.Code, err.Description, err.Description2)
		//logger.Infof("WARN:  %v", err)
	}
	fmt.Println("\n")
	return nil
}

/*
func checkObject(dest ocfl.OCFLFS, extensionFactory *ocfl.ExtensionFactory, logger *logging.Logger) error {
	ctx := ocfl.NewContextValidation(context.TODO())
	defer showStatus(ctx)
	object, err := ocfl.newObject(ctx, dest, "", "", logger)
	if err != nil {
		return errors.Wrap(err, "cannot load object")
	}
	if err := object.Check(); err != nil {
		return errors.Wrapf(err, "check of %s failed", object.GetID())
	}
	return nil
}
*/

func check(dest ocfl.OCFLFS, extensionFactory *ocfl.ExtensionFactory, logger *logging.Logger) error {
	ctx := ocfl.NewContextValidation(context.TODO())
	defer showStatus(ctx)
	storageRoot, err := ocfl.LoadStorageRoot(ctx, dest, extensionFactory, logger)

	if err != nil {
		return errors.Wrap(err, "cannot create new storageroot")
	}
	if err := storageRoot.Check(); err != nil {
		return errors.Wrap(err, "ocfl not valid")
	}
	return nil
}

func ingest(dest ocfl.OCFLFS, srcdir string, extensionFactory *ocfl.ExtensionFactory, storageRootExtensions, objectExtensions []ocfl.Extension, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, logger *logging.Logger) error {

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

	ctx := ocfl.NewContextValidation(context.TODO())
	var storageRoot ocfl.StorageRoot
	if dest.HasContent() {
		storageRoot, err = ocfl.LoadStorageRoot(ctx, dest, extensionFactory, logger)
		if err != nil {
			return errors.Wrap(err, "cannot load new storageroot")
		}
	} else {
		storageRoot, err = ocfl.CreateStorageRoot(ctx, dest, VERSION, extensionFactory, storageRootExtensions, digest, logger)
		if err != nil {
			return errors.Wrap(err, "cannot create new storageroot")
		}
	}

	defer showStatus(ctx)

	// TEST042
	id := "https://hdl.handle.net/20394823094823/test042"

	var o ocfl.Object
	exists, err := storageRoot.ObjectExists(id)
	if err != nil {
		return errors.Wrapf(err, "cannot check for existence of %s", id)
	}
	if exists {
		o, err = storageRoot.LoadObjectByID(id)
		if err != nil {
			return errors.Wrapf(err, "cannot load object %s", id)
		}
	} else {
		o, err = storageRoot.CreateObject(id, VERSION, digest, fixity, objectExtensions)
		if err != nil {
			return errors.Wrapf(err, "cannot create object %s", id)
		}
	}

	if err := o.StartUpdate("test 42", "Jürgen Enge", "juergen.enge@unibas.ch"); err != nil {
		return errors.Wrapf(err, "cannot start update for object %s", "test 42")
	}

	if err := o.AddFolder(os.DirFS(srcdir), false); err != nil {
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
var defaultExtensionFolder = flag.String("extensions", "", "folder with default extension configs. needs subfolder object and storageroot")
var sha256 = flag.Bool("sha256", false, "use sha256 as main hash algorithm")
var sha512 = flag.Bool("sha512", false, "use sha512 as main hash algorithm")
var fixity = flag.String("fixity", "", "comma separated list of fixity digest algorithms")

type x interface {
	print()
}

type y struct {
	s string
}

func (_y y) print() {
	fmt.Println(_y.s)
}

func main() {

	var err error

	//	var version = flag.String("version", "", "ocfl version")

	flag.Parse()

	if *sha256 && *sha512 {
		log.Println("please do not use -sha256 AND -sha512 at the same time")
		return
	}

	if *checkFlag && (*sha256 || *sha512) {
		log.Println("ignoring hash flags for check")
	}

	logger, lf := lm.CreateLogger("ocfl", *logfile, nil, *loglevel, LOGFORMAT)
	defer lf.Close()

	extensionFactory, err := ocfl.NewExtensionFactory(logger)
	if err != nil {
		logger.Errorf("cannot instantiate extension factory: %v", err)
		return
	}

	extensionFactory.AddCreator(extension.DirectCleanName, func(fs ocfl.OCFLFS) (ocfl.Extension, error) {
		return extension.NewDirectCleanFS(fs)
	})

	extensionFactory.AddCreator(extension.PathDirectName, func(fs ocfl.OCFLFS) (ocfl.Extension, error) {
		return extension.NewPathDirectFS(fs)
	})

	extensionFactory.AddCreator(extension.StorageLayoutFlatDirectName, func(fs ocfl.OCFLFS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutFlatDirectFS(fs)
	})

	extensionFactory.AddCreator(extension.StorageLayoutHashAndIdNTupleName, func(fs ocfl.OCFLFS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutHashAndIdNTupleFS(fs)
	})

	extensionFactory.AddCreator(extension.StorageLayoutHashedNTupleName, func(fs ocfl.OCFLFS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutHashedNTupleFS(fs)
	})

	extensionFactory.AddCreator(extension.StorageLayoutPairTreeName, func(fs ocfl.OCFLFS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutPairTreeFS(fs)
	})

	//
	// load default extensions
	//
	var dExtDirFS fs.FS
	if *defaultExtensionFolder == "" {
		dExtDirFS = defaultextensions.DefaultExtensionFS
	} else {
		dExtDirFS = os.DirFS(*defaultExtensionFolder)
	}
	ofs, err := genericfs.NewGenericFS(dExtDirFS, ".", logger)
	if err != nil {
		logger.Panicf("cannot create generic fs for %v", dExtDirFS)
	}
	subFS, err := ofs.SubFS("storageroot")
	if err != nil {
		logger.Panicf("cannot create subfs %s for %v", "storageroot", dExtDirFS)
	}
	storageRootExtensions, err := extensionFactory.LoadExtensions(subFS)
	if err != nil {
		logger.Panicf("cannot load extension folder %v", ofs)
	}
	subFS, err = ofs.SubFS("object")
	if err != nil {
		logger.Panicf("cannot create subfs %s for %v", "object", dExtDirFS)
	}
	objectExtensions, err := extensionFactory.LoadExtensions(subFS)
	if err != nil {
		logger.Panicf("cannot load extension folder %v", ofs)
	}

	var ocfs ocfl.OCFLFS

	var zipSize int64
	var zipReader *os.File
	var zipWriter *os.File

	var zipFile string
	//var objectPath string
	if strings.HasSuffix(strings.ToLower(*target), ".zip") {
		zipFile = *target
	} else {
		if pos := strings.Index(*target, ".zip/"); pos != -1 {
			zipFile = (*target)[0 : pos+4]
			//objectPath = (*target)[pos+4:]
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
		ocfs, err = zipfs.NewFSIO(zipReader, zipSize, zipWriter, ".", logger)
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

	var digest checksum.DigestAlgorithm
	if *sha256 {
		digest = checksum.DigestSHA256
	} else {
		digest = checksum.DigestSHA512
	}
	switch {
	case *srcDir != "":

		if err := ingest(ocfs, *srcDir, extensionFactory, storageRootExtensions, objectExtensions, digest, []checksum.DigestAlgorithm{checksum.DigestMD5, checksum.DigestBlake2b256}, logger); err != nil {
			stackTrace := ocfl.GetErrorStacktrace(err)
			logger.Errorf("%v%+v", err, stackTrace)
			panic(err)
		}
		/*
			case *checkObjectFlag:
				objfs := ocfs.SubFS(objectPath)
				if err := checkObject(objfs, extensionFactory, digest, logger); err != nil {
					logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
					panic(err)
				}
		*/
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
	}
	if zipWriter != nil {
		if err := os.Rename(fmt.Sprintf("%s.tmp", *target), *target); err != nil {
			logger.Error(err)
			panic(err)
		}
	}
}
