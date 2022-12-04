package cmd

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"os"
	"path/filepath"
	"strings"
)

var createCmd = &cobra.Command{
	Use:     "create [path to ocfl structure]",
	Aliases: []string{},
	Short:   "creates a new ocfl structure with initial content of one object",
	Long:    "initializes an empty ocfl structure and adds a directory subtree to it",
	Example: "gocfl create ./archive.zip /tmp/testdata --sha512 -u 'Jane Doe' -a 'mailto:user@domain' -m 'initial add' -object-id 'id:abc123'",
	Args:    cobra.ExactArgs(2),
	Run:     doCreate,
}

func initCreate() {
	createCmd.Flags().StringVarP(&flagExtensionFolder, "extensions", "e", "", "folder with extension configurations")

	createCmd.Flags().VarP(
		enumflag.New(&flagVersion, "ocfl-version", VersionIds, enumflag.EnumCaseInsensitive),
		"ocfl-version", "v", "ocfl version for new storage root")

	createCmd.Flags().StringVarP(&objectID, "object-id", "i", "", "object id to update (required)")
	createCmd.MarkFlagRequired("object-id")

	createCmd.Flags().StringVarP(&message, "message", "m", "", "message for new object version (required)")
	createCmd.MarkFlagRequired("message")

	createCmd.Flags().StringVarP(&userName, "user-name", "u", "", "user name for new object version (required)")
	createCmd.MarkFlagRequired("user-name")

	createCmd.Flags().StringVarP(&userAddress, "user-address", "a", "", "user address for new object version (required)")
	createCmd.MarkFlagRequired("user-address")

	createCmd.Flags().StringVarP(&fixity, "fixity", "f", "", "comma separated list of digest algorithms for fixity")

	createCmd.Flags().BoolVar(&digestSHA256, "sha256", false, "use sha256 as digest")
	createCmd.Flags().BoolVar(&digestSHA512, "sha512", true, "use sha512 as digest")
	createCmd.MarkFlagsMutuallyExclusive("sha256", "sha512")
}

func addObjectByPath(storageRoot ocfl.StorageRoot, fixity []checksum.DigestAlgorithm, defaultExtensions []ocfl.Extension, checkDuplicates bool, id, userName, userAddress, message, path string) (bool, error) {
	var o ocfl.Object
	exists, err := storageRoot.ObjectExists(objectID)
	if err != nil {
		return false, errors.Wrapf(err, "cannot check for existence of %s", id)
	}
	if exists {
		o, err = storageRoot.LoadObjectByID(id)
		if err != nil {
			return false, errors.Wrapf(err, "cannot load object %s", id)
		}
	} else {
		o, err = storageRoot.CreateObject(id, storageRoot.GetVersion(), storageRoot.GetDigest(), fixity, defaultExtensions)
		if err != nil {
			return false, errors.Wrapf(err, "cannot create object %s", id)
		}
	}
	if err := o.StartUpdate(message, userName, userAddress); err != nil {
		return false, errors.Wrapf(err, "cannot start update for object %s", id)
	}

	if err := o.AddFolder(os.DirFS(path), checkDuplicates); err != nil {
		return false, errors.Wrapf(err, "cannot add folder '%s' to '%s'", path, id)
	}

	if err := o.Close(); err != nil {
		return false, errors.Wrapf(err, "cannot close object '%s'", id)
	}

	return o.IsModified(), nil
}

func doCreate(cmd *cobra.Command, args []string) {
	ocflPath := filepath.ToSlash(filepath.Clean(args[0]))
	srcPath := filepath.ToSlash(filepath.Clean(args[1]))

	fmt.Printf("creating '%s'\n", ocflPath)

	logger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, LogLevelIds[persistentFlagLoglevel][0], LOGFORMAT)
	defer lf.Close()
	logger.Infof("creating '%s'", ocflPath)

	var fixityAlgs = []checksum.DigestAlgorithm{}
	for _, alg := range strings.Split(fixity, ",") {
		alg = strings.TrimSpace(strings.ToLower(alg))
		if alg == "" {
			continue
		}
		if _, err := checksum.GetHash(checksum.DigestAlgorithm(alg)); err != nil {
			logger.Errorf("unknown hash function '%s': %v", alg, err)
			return
		}
		fixityAlgs = append(fixityAlgs, checksum.DigestAlgorithm(alg))
	}
	var digest checksum.DigestAlgorithm
	if digestSHA256 {
		digest = checksum.DigestSHA256
	}
	if digestSHA512 {
		digest = checksum.DigestSHA512
	}

	if _, err := os.Stat(srcPath); err != nil {
		logger.Errorf("cannot stat '%s': %v", srcPath, err)
		return
	}

	finfo, err := os.Stat(ocflPath)
	if err != nil {
		if !(os.IsNotExist(err) && strings.HasSuffix(strings.ToLower(ocflPath), ".zip")) {
			logger.Errorf("cannot stat '%s': %v", ocflPath, err)
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
	} else {
		if strings.HasSuffix(strings.ToLower(ocflPath), ".zip") {
			logger.Errorf("path '%s' already exists", ocflPath)
			fmt.Printf("path '%s' already exists\n", ocflPath)
			return
		}
		if !finfo.IsDir() {
			logger.Errorf("'%s' is not a directory", ocflPath)
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
	}

	extensionFactory, err := ocfl.NewExtensionFactory(logger)
	if err != nil {
		logger.Errorf("cannot instantiate extension factory: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	if err := initExtensionFactory(extensionFactory); err != nil {
		logger.Errorf("cannot initialize extension factory: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	storageRootExtensions, objectExtensions, err := initDefaultExtensions(extensionFactory, flagExtensionFolder, logger)
	if err != nil {
		logger.Errorf("cannot initialize default extensions: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	tempFile := fmt.Sprintf("%s.tmp", ocflPath)
	reader, writer, ocfs, err := OpenRW(ocflPath, tempFile, logger)
	if err != nil {
		logger.Errorf("cannot create target filesystem: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	defer showStatus(ctx)
	storageRoot, err := ocfl.CreateStorageRoot(ctx, ocfs, VersionIdsVersion[flagVersion], extensionFactory, storageRootExtensions, digest, logger)
	if err != nil {
		ocfs.Discard()
		logger.Errorf("cannot create new storageroot: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	_, err = addObjectByPath(storageRoot, fixityAlgs, objectExtensions, false, objectID, userName, userAddress, message, srcPath)
	if err != nil {
		logger.Errorf("error adding content to storageroot filesystem '%s': %v", ocfs, err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
	}

	if err := ocfs.Close(); err != nil {
		logger.Errorf("error closing filesystem '%s': %v", ocfs, err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
	} else {
		if reader != nil && reader != (*os.File)(nil) {
			if err := reader.Close(); err != nil {
				logger.Errorf("error closing reader: %v", err)
				logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			}
		}
		if err := writer.Close(); err != nil {
			logger.Errorf("error closing writer: %v", err)
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		}
		if err := os.Rename(tempFile, ocflPath); err != nil {
			logger.Errorf("cannot rename '%s' -> '%s': %v", tempFile, ocflPath, err)
		}
	}

}
