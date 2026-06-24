package util

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"emperror.dev/emperror"
	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/ocflerrors"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
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
	rule_1_5 := regexp.MustCompile("[\u0000-\u001F\u007F\n\r\t*?:\\[\\]\"<>|(){}&'!;]")
	rule_2_4_6 := regexp.MustCompile(`^[\s\-~]*(.*?)\s*$`)

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

func CheckPrefix(list []string, suffix string, f func(str1, str2 string) bool) {
	slices.Sort(list)
	for j := 0; j < len(list)-1; j++ {
		prefix := strings.TrimSuffix(list[j], suffix) + suffix
		prefix2 := strings.TrimSuffix(list[j+1], suffix) + suffix
		if strings.HasPrefix(prefix2, prefix) {
			if !f(list[j], list[j+1]) {
				return
			}
		}
	}

}

func Fullpath(path string) (string, error) {
	path = filepath.ToSlash(filepath.Clean(path))
	// if empty use current folder
	if path == "" {
		currdir, err := os.Getwd()
		if err != nil {
			return "", errors.Wrap(err, "cannot get current directory")
		}
		return filepath.ToSlash(currdir), nil
	}
	// replace starting ~ with user home
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Wrap(err, "cannot get home directory")
		}
		path = filepath.ToSlash(filepath.Join(home, path[1:]))
	}
	// if it is an absolute path, all fine
	if filepath.IsAbs(path) || path[0] == '/' {
		return path, nil
	}
	currdir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "cannot get current directory")
	}
	return filepath.ToSlash(filepath.Join(currdir, path)), nil
}

// CopyMapStringSlice deep copy map of string slices
func CopyMapStringSlice(dest, src map[string][]string) {
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

func GetObjectVersion(sourceFS fs.FS) (version.OCFLVersion, error) {
	ver, err := GetVersion(sourceFS, ".", "ocfl_object_")
	if err != nil {
		return "", errors.Wrap(err, "getting object version")
	}
	return ver, nil
}

func GetStorageRootVersion(sourceFS fs.FS) (version.OCFLVersion, error) {
	ver, err := GetVersion(sourceFS, ".", "ocfl_")
	if err != nil {
		return "", errors.Wrap(err, "getting storage root version")
	}
	return ver, nil
}

func GetVersion(fsys fs.FS, folder, prefix string) (ver version.OCFLVersion, err error) {
	rString := fmt.Sprintf("0=%s([0-9]+\\.[0-9]+)", prefix)
	r, err := regexp.Compile(rString)
	if err != nil {
		return "", errors.Wrapf(err, "cannot compile %s", rString)
	}
	if folder == "" {
		folder = "."
	}
	files, err := fs.ReadDir(fsys, folder)
	if err != nil {
		return "", errors.Wrapf(err, "cannot get files in folder %s", folder)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		matches := r.FindStringSubmatch(file.Name())
		if matches != nil {
			if ver != "" {
				return "", ocflerrors.ErrVersionMultiple
			}
			ver = version.OCFLVersion(matches[1])
			cnt, err := fs.ReadFile(fsys, path.Join(folder, file.Name()))
			if err != nil {
				return "", errors.Wrapf(err, "cannot read %s/%s", folder, file.Name())
			}

			t := fmt.Sprintf("%s%s", prefix, ver)
			if string(cnt) != t+"\n" && string(cnt) != t+"\r\n" {
				return ver, ocflerrors.ErrInvalidContent
				//addValidationErrors(ctx, GetValidationError(version, E007).AppendDescription("%s: %s != %s", file.Name(), cnt, t+"\\n"))
			}
		}
	}
	if ver == "" {
		return "", ocflerrors.ErrVersionNone
	}
	return ver, nil
}

func validVersion(ctx context.Context, fsys fs.FS, ver version.OCFLVersion, folder, prefix string) bool {
	v, _ := GetVersion(fsys, folder, prefix)
	return v == ver
}

// SliceContains Contains reports whether vs is present in s
func SliceContains[E comparable](s []E, vs []E) bool {
	for _, v := range vs {
		if !slices.Contains(s, v) {
			return false
		}
	}
	return true
}

func SliceInsertSorted[E constraints.Ordered](data []E, v E) []E {
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
var rule_2_4_6 = regexp.MustCompile(`^[\s\-~]*(.*?)\s*$`)

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
