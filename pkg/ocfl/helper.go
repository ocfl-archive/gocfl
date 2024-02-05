package ocfl

import (
	"bytes"
	"compress/gzip"
	"context"
	"emperror.dev/emperror"
	"emperror.dev/errors"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// FixFilename
/**********************************************************************
 * 1) Forbid/escape ASCII control characters (bytes 1-31 and 127) in filenames, including newline, escape, and tab.
 *    I know of no user or program that actually requires this capability. As far as I can tell, this capability
 *    exists only to make it hard to write correct software, to ease the job of attackers, and to create
 *    interoperability problems. Chuck it.
 * 2) Forbid/escape leading “-”. This way, you can always distinguish option flags from filenames, eliminating a host
 *    of stupid emperror. Nobody in their right mind writes programs that depend on having dash-prefixed files on a Unix
 *    system. Even on Windows systems they’re a bad idea, because many programs use “-” instead of “/” to identify options.
 * 3) Forbid/escape filenames that aren’t a valid UTF-8 encoding. This way, filenames can always be correctly displayed.
 *    Trying to use environment values like LC_ALL (or other LC_* values) or LANG is just a hack that often fails. This
 *    will take time, as people slowly transition and minor tool problems get fixed, but I believe that transition is
 *    already well underway.
 * 4) Forbid/escape leading/trailing space characters — at least trailing spaces. Adjacent spaces are somewhat dodgy,
 *    too. These confuse users when they happen, with no utility. In particular, filenames that are only space characters
 *    are nothing but trouble. Some systems may want to go further and forbid space characters outright, but I doubt that’ll
 *    be acceptable everywhere, and with the other approaches these are less necessary. As noted above, an interesting
 *    alternative would be quietly convert (in the API) all spaces into unbreakable spaces.
 * 5) Forbid/escape “problematic” characters that get specially interpreted by shells, other interpreters (such as perl),
 *    and HTML/XML. This is less important, and I would expect this to happen (at most) on specific systems. With the steps
 *    above, a lot of programs and statements like “cat *” just work correctly. But funny characters cause troubles for shell
 *    scripts and perl, because they need to quote them when typing in commands.. and they often forget to do so. They can
 *    also be a cause for trouble when they’re passed down to other programs, especially if they run “exec” and so on. They’re
 *    also helpful for web applications, again, because the characters that should be escapes are sometimes not escaped. A short
 *    list would be “*”, “?”, and “[”; by eliminating those three characters and control characters from filenames, and removing
 *    the space character from IFS, you can process filenames in shells without quoting variable references — eliminating a
 *    common source of emperror. Forbidding/escaping “<” and “>” would eliminate a source of nasty errors for perl programs, web
 *    applications, and anyone using HTML or XML. A more stringent list would be “*?:[]"<>|(){}&'!\;” (this is Glindra’s “safe”
 *    list with ampersand, single-quote, bang, backslash, and semicolon added). This list is probably a little extreme, but let’s
 *    try and see. As noted earlier, I’d need to go through a complete analysis of all characters for a final list; for security,
 *    you want to identify everything that is permissible, and disallow everything else, but its manifestation can be either way
 *    as long as you’ve considered all possible cases. But if this set can be determined locally, based on local requirements,
 *    there’s less need to get complete agreement on a list.
 * 6) Forbid/escape leading “~” (tilde). Shells specially interpret such filenames. This is definitely low priority.
 *
 * https://www.dwheeler.com/essays/fixing-unix-linux-filenames.html
 */
func FixFilename(fname string) string {
	rule_1_5 := regexp.MustCompile("[\x00-\x1F\x7F\n\r\t*?:\\[\\]\"<>|(){}&'!\\;]")
	rule_2_4_6 := regexp.MustCompile("^[\\s\\-~]*(.*?)\\s*$")

	fname = strings.ToValidUTF8(fname, "")

	names := strings.Split(fname, "/")
	result := []string{}

	for _, n := range names {
		n = rule_1_5.ReplaceAllString(n, "_")
		n = rule_2_4_6.ReplaceAllString(n, "$1")
		result = append(result, n)
	}

	fname = filepath.ToSlash(filepath.Join(result...))
	if len(result) > 0 {
		if result[0] == "" {
			fname = "/" + fname
		}

	}
	return fname
}

func Fullpath(path string) (string, error) {
	path = filepath.ToSlash(filepath.Clean(path))
	if path == "" {
		currdir, err := os.Getwd()
		if err != nil {
			return "", errors.Wrap(err, "cannot get current directory")
		}
		return filepath.ToSlash(currdir), nil
	}
	if strings.HasPrefix(path, "/") {
		// absolute path on un*x
		if runtime.GOOS != "windows" {
			return path, nil
		}
		// UNC path
		if strings.HasPrefix(path, "//") {
			return path, nil
		}

		currdir, err := os.Getwd()
		if err != nil {
			return "", errors.Wrap(err, "cannot get current directory")
		}
		currdir = filepath.ToSlash(currdir)
		// this is a problem. current dir is unc, but path is not
		if strings.HasPrefix(currdir, "/") {
			return "", errors.Errorf("current directory '%s' is UNC path, but path '%s' is not", currdir, path)
		}
		// no drive letter in current dir
		if len(currdir) > 1 && currdir[1] != ':' {
			return "", errors.Wrapf(err, "no drive letter in current folder '%s' path with leading '/' not allowed - %s", currdir, path)
		}
		if len(currdir) > 1 {
			return filepath.ToSlash(filepath.Join(currdir[0:2], path)), nil
		}
		return path, nil
	}
	// absolute path on windows with drive letter
	if runtime.GOOS == "windows" && len(path) > 1 && path[1] == ':' {
		return path, nil
	}
	currdir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "cannot get current directory")
	}
	return filepath.ToSlash(filepath.Join(currdir, path)), nil
}

// deep copy map of string slices
func copyMapStringSlice(dest, src map[string][]string) {
	for key, val := range src {
		dest[key] = make([]string, len(val))
		copy(dest[key], val)
	}
}

func GetErrorStacktrace(err error) errors.StackTrace {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	var stack errors.StackTrace

	errors.UnwrapEach(err, func(err error) bool {
		e := emperror.ExposeStackTrace(err)
		st, ok := e.(stackTracer)
		if !ok {
			return true
		}

		stack = st.StackTrace()
		return true
	})

	if len(stack) > 2 {
		stack = stack[:len(stack)-2]
	}
	return stack
	// fmt.Printf("%+v", st[0:2]) // top two frames
}

func getVersion(ctx context.Context, fsys fs.FS, folder, prefix string) (version OCFLVersion, err error) {
	rString := fmt.Sprintf("0=%s([0-9]+\\.[0-9]+)", prefix)
	r, err := regexp.Compile(rString)
	if err != nil {
		return "", errors.Wrapf(err, "cannot compile %s", rString)
	}
	files, err := fs.ReadDir(fsys, folder)
	if err != nil {
		return "", errors.Wrapf(err, "cannot get %s files", folder)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		matches := r.FindStringSubmatch(file.Name())
		if matches != nil {
			if version != "" {
				return "", errVersionMultiple
			}
			version = OCFLVersion(matches[1])
			cnt, err := fs.ReadFile(fsys, fmt.Sprintf("%s/%s", folder, file.Name()))
			if err != nil {
				return "", errors.Wrapf(err, "cannot read %s/%s", folder, file.Name())
			}

			t := fmt.Sprintf("%s%s", prefix, version)
			if string(cnt) != t+"\n" && string(cnt) != t+"\r\n" {
				return version, errInvalidContent
				//addValidationErrors(ctx, GetValidationError(version, E007).AppendDescription("%s: %s != %s", file.Name(), cnt, t+"\\n"))
			}
		}
	}
	if version == "" {
		return "", errVersionNone
	}
	return version, nil
}

func validVersion(ctx context.Context, fsys fs.FS, version OCFLVersion, folder, prefix string) bool {
	v, _ := getVersion(ctx, fsys, folder, prefix)
	return v == version
}

// Contains reports whether vs is present in s
func sliceContains[E comparable](s []E, vs []E) bool {
	for _, v := range vs {
		if !slices.Contains(s, v) {
			return false
		}
	}
	return true
}

func sliceInsertSorted[E constraints.Ordered](data []E, v E) []E {
	var dummy E
	i, _ := slices.BinarySearch(data, v) // find slot
	data = append(data, dummy)           // extend the slice
	copy(data[i+1:], data[i:])           // make room
	data[i] = v
	return data
}

func sliceInsertAt[E comparable](data []E, i int, v E) []E {
	if i == len(data) {
		// Insert at end is the easy case.
		return append(data, v)
	}

	// Make space for the inserted element by shifting
	// values at the insertion index up one index. The call
	// to append does not allocate memory when cap(data) is
	// greater ​than len(data).
	data = append(data[:i+1], data[i:]...)

	// Insert the new element.
	data[i] = v

	// Return the updated slice.
	return data
}

func showStatus(ctx context.Context) error {
	status, err := GetValidationStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot get status of validation")
	}
	status.Compact()
	for _, _err := range status.Errors {
		fmt.Println(_err.Error())
		//logger.Infof("ERROR: %v", err)
	}
	/*
		for _, warning := range status.Warnings {
			fmt.Println(warning.Error())
			//logger.Infof("WARN:  %v", err)
		}
		fmt.Println("\n")
	*/
	return nil
}

// CleanPath
/**********************************************************************
 * 1) Forbid/escape ASCII control characters (bytes 1-31 and 127) in filenames, including newline, escape, and tab.
 *    I know of no user or program that actually requires this capability. As far as I can tell, this capability
 *    exists only to make it hard to write correct software, to ease the job of attackers, and to create
 *    interoperability problems. Chuck it.
 * 2) Forbid/escape leading “-”. This way, you can always distinguish option flags from filenames, eliminating a host
 *    of stupid emperror. Nobody in their right mind writes programs that depend on having dash-prefixed files on a Unix
 *    system. Even on Windows systems they’re a bad idea, because many programs use “-” instead of “/” to identify options.
 * 3) Forbid/escape filenames that aren’t a valid UTF-8 encoding. This way, filenames can always be correctly displayed.
 *    Trying to use environment values like LC_ALL (or other LC_* values) or LANG is just a hack that often fails. This
 *    will take time, as people slowly transition and minor tool problems get fixed, but I believe that transition is
 *    already well underway.
 * 4) Forbid/escape leading/trailing space characters — at least trailing spaces. Adjacent spaces are somewhat dodgy,
 *    too. These confuse users when they happen, with no utility. In particular, filenames that are only space characters
 *    are nothing but trouble. Some systems may want to go further and forbid space characters outright, but I doubt that’ll
 *    be acceptable everywhere, and with the other approaches these are less necessary. As noted above, an interesting
 *    alternative would be quietly convert (in the API) all spaces into unbreakable spaces.
 * 5) Forbid/escape “problematic” characters that get specially interpreted by shells, other interpreters (such as perl),
 *    and HTML/XML. This is less important, and I would expect this to happen (at most) on specific systems. With the steps
 *    above, a lot of programs and statements like “cat *” just work correctly. But funny characters cause troubles for shell
 *    scripts and perl, because they need to quote them when typing in commands.. and they often forget to do so. They can
 *    also be a cause for trouble when they’re passed down to other programs, especially if they run “exec” and so on. They’re
 *    also helpful for web applications, again, because the characters that should be escapes are sometimes not escaped. A short
 *    list would be “*”, “?”, and “[”; by eliminating those three characters and control characters from filenames, and removing
 *    the space character from IFS, you can process filenames in shells without quoting variable references — eliminating a
 *    common source of emperror. Forbidding/escaping “<” and “>” would eliminate a source of nasty errors for perl programs, web
 *    applications, and anyone using HTML or XML. A more stringent list would be “*?:[]"<>|(){}&'!\;” (this is Glindra’s “safe”
 *    list with ampersand, single-quote, bang, backslash, and semicolon added). This list is probably a little extreme, but let’s
 *    try and see. As noted earlier, I’d need to go through a complete analysis of all characters for a final list; for security,
 *    you want to identify everything that is permissible, and disallow everything else, but its manifestation can be either way
 *    as long as you’ve considered all possible cases. But if this set can be determined locally, based on local requirements,
 *    there’s less need to get complete agreement on a list.
 * 6) Forbid/escape leading “~” (tilde). Shells specially interpret such filenames. This is definitely low priority.
 *
 * https://www.dwheeler.com/essays/fixing-unix-linux-filenames.html
 */
var rule_1_5 = regexp.MustCompile("[\x00-\x1F\x7F\n\r\t*?:\\[\\]\"<>|(){}&'!\\;#@]")
var rule_2_4_6 = regexp.MustCompile("^[\\s\\-~]*(.*?)\\s*$")

var ErrFilenameTooLong = errors.New("filename too long")
var ErrPathnameTooLong = errors.New("pathname too long")

func CleanPath(fname string, MaxFilenameLength, MaxPathnameLength int) (string, error) {

	fname = strings.ToValidUTF8(fname, "_")

	names := strings.Split(fname, "/")
	result := []string{}

	for _, n := range names {
		n = rule_1_5.ReplaceAllString(n, "_")
		n = rule_2_4_6.ReplaceAllString(n, "$1")

		lenN := len(n)
		if lenN > MaxFilenameLength {
			return "", errors.Wrapf(ErrFilenameTooLong, "filename: %s", n)
		}
		if lenN > 0 {
			result = append(result, n)
		}
	}

	fname = strings.Join(result, "/")

	if len(fname) > MaxPathnameLength {
		return "", errors.Wrapf(ErrPathnameTooLong, "pathname: %s", fname)
	}

	return fname, nil
}

func ReadFile(object Object, name, version, storageType, storageName string, fsys fs.FS) ([]byte, error) {
	var targetname string
	switch storageType {
	case "area":
		path, err := object.GetAreaPath(storageName)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get area path for '%s'", storageName)
		}
		targetname = object.GetInventory().BuildManifestNameVersion(fmt.Sprintf("%s/%s", path, name), version)
		//targetname = fmt.Sprintf("%s/content/%s/indexer_%s.jsonl%s", version, path, version, ext)
		fsys = object.GetFS()
	case "path":
		path, err := object.GetAreaPath("content")
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		targetname = object.GetInventory().BuildManifestNameVersion(fmt.Sprintf("%s/%s/%s", path, storageName, name), version)
		//targetname = fmt.Sprintf("%s/content/%s/indexer_%s.jsonl%s", v, sl.IndexerConfig.StorageName, v, ext)
		fsys = object.GetFS()
	case "extension":
		targetname = strings.TrimLeft(fmt.Sprintf("%s/%s", storageName, name), "/")
	default:
		return nil, errors.Errorf("unsupported storage type '%s'", storageType)
	}

	return fs.ReadFile(fsys, targetname)
}

func ReadJsonL(object Object, name, version, compress, storageType, storageName string, fsys fs.FS) ([]byte, error) {
	var ext string
	switch compress {
	case "brotli":
		ext = ".br"
	case "gzip":
		ext = ".gz"
	case "none":
	default:
		return nil, errors.Errorf("invalid compression '%s'", compress)
	}
	var targetname string
	switch storageType {
	case "area":
		path, err := object.GetAreaPath(storageName)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get area path for '%s'", storageName)
		}
		targetname = object.GetInventory().BuildManifestNameVersion(fmt.Sprintf("%s/%s_%s.jsonl%s", path, name, version, ext), version)
		//targetname = fmt.Sprintf("%s/content/%s/indexer_%s.jsonl%s", version, path, version, ext)
		fsys = object.GetFS()
	case "path":
		path, err := object.GetAreaPath("content")
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		targetname = object.GetInventory().BuildManifestNameVersion(fmt.Sprintf("%s/%s/%s_%s.jsonl%s", path, storageName, name, version, ext), version)
		//targetname = fmt.Sprintf("%s/content/%s/indexer_%s.jsonl%s", v, sl.IndexerConfig.StorageName, v, ext)
		fsys = object.GetFS()
	case "extension":
		targetname = strings.TrimLeft(fmt.Sprintf("%s/%s_%s.jsonl%s", storageName, name, version, ext), "/")
	default:
		return nil, errors.Errorf("unsupported storage type '%s'", storageType)
	}

	var reader io.Reader
	f, err := fsys.Open(targetname)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open '%v/%s'", fsys, targetname)
	}
	switch compress {
	case "brotli":
		reader = brotli.NewReader(f)
	case "gzip":
		reader, err = gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, errors.Wrapf(err, "cannot open gzip reader on '%s'", targetname)
		}
	case "none":
		reader = f
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		if f != nil {
			f.Close()
		}
		return nil, errors.Wrapf(err, "cannot read '%s'", targetname)
	}
	if f != nil {
		if err := f.Close(); err != nil {
			return nil, errors.Wrapf(err, "cannot close '%s'", targetname)
		}
	}
	return data, nil
}

func WriteJsonL(object Object, name string, brotliData []byte, compress, storageType, storageName string, fsys fs.FS) error {
	var bufReader = bytes.NewBuffer(brotliData)
	var ext string
	var reader io.Reader
	switch compress {
	case "brotli":
		ext = ".br"
		reader = bufReader
	case "gzip":
		ext = ".gz"
		brotliReader := brotli.NewReader(bufReader)
		pr, pw := io.Pipe()
		go func() {
			defer pw.Close()
			gzipWriter := gzip.NewWriter(pw)
			defer gzipWriter.Close()
			if _, err := io.Copy(gzipWriter, brotliReader); err != nil {
				pw.CloseWithError(errors.Wrapf(err, "error on gzip compressor"))
			}
		}()
		reader = pr
	case "none":
		reader = brotli.NewReader(bufReader)
	default:
		return errors.Errorf("invalid compression '%s'", compress)
	}

	head := object.GetInventory().GetHead()
	switch strings.ToLower(storageType) {
	case "area":
		targetname := fmt.Sprintf("%s_%s.jsonl%s", name, head, ext)
		if _, err := object.AddReader(io.NopCloser(reader), []string{targetname}, storageName, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", targetname)
		}
	case "path":
		path, err := object.GetAreaPath("content")
		if err != nil {
			return errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		targetname := fmt.Sprintf("%s/%s/%s_%s.jsonl%s", path, storageName, name, head, ext)

		//targetname := fmt.Sprintf("%s/%s_%s.jsonl%s", name, storageName, head, ext)
		if _, err := object.AddReader(io.NopCloser(reader), []string{targetname}, "", true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", targetname)
		}
	case "extension":
		targetname := strings.TrimLeft(fmt.Sprintf("%s/%s_%s.jsonl%s", storageName, name, head, ext), "/")
		fp, err := writefs.Create(fsys, targetname)
		if err != nil {
			return errors.Wrapf(err, "cannot create '%v/%s'", fsys, targetname)
		}
		defer fp.Close()
		if _, err := io.Copy(fp, reader); err != nil {
			return errors.Wrapf(err, "cannot write '%v/%s'", fsys, targetname)
		}
	default:
		return errors.Errorf("unsupported storage type '%s'", storageType)
	}

	return nil
}
