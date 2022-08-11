package extension

import (
	"regexp"
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
	rule_1_5 := regexp.MustCompile("[\x00-\x1F\x7F\n\r\t*?:\\[\\]\"<>|(){}&'!\\;#@]")
	rule_2_4_6 := regexp.MustCompile("^[\\s\\-~]*(.*?)\\s*$")

	fname = strings.ToValidUTF8(fname, "_")

	names := strings.Split(fname, "/")
	result := []string{}

	for _, n := range names {
		n = rule_1_5.ReplaceAllString(n, "_")
		n = rule_2_4_6.ReplaceAllString(n, "$1")
		result = append(result, n)
	}

	fname = strings.Join(result, "/")
	if len(result) > 0 {
		if result[0] == "" {
			fname = "/" + fname
		}

	}
	return fname
}
