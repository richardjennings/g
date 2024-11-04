package g

import (
	"fmt"
	"testing"
)

func TestIsIgnored(t *testing.T) {
	type tc struct {
		Pattern string
		Path    string
		Expect  bool
	}
	for _, tt := range []tc{
		// A blank line matches no files, so it can serve as a separator for
		// readability.
		{Pattern: "", Path: "/test/hello", Expect: false},
		{Pattern: "", Path: "/.git/HEAD", Expect: true},

		// A line starting with # serves as a comment. Put a backslash ("\") in
		// front of the first hash for patterns that begin with a hash.
		{Pattern: "#test", Path: "/test/#test", Expect: false},
		{Pattern: `\#test`, Path: "/test/#test", Expect: true},
		// Trailing spaces are ignored unless they are quoted with backslash
		// ("\").
		// @todo

		// An optional prefix "!" which negates the pattern; any matching file
		// excluded by a previous pattern will become included again. It is not
		// possible to re-include a file if a parent directory of that file is
		// excluded. Git doesn’t list excluded directories for performance
		// reasons, so any patterns on contained files have no effect, no matter
		// where they are defined. Put a backslash ("\") in front of the first
		// "!" for patterns that begin with a literal "!", for example,
		// "\!important!.txt".
		// @todo

		// The slash "/" is used as the directory separator. Separators may
		// occur at the beginning, middle or end of the .gitignore search
		// pattern.

		// If there is a separator at the beginning or middle (or both) of the
		// pattern, then the pattern is relative to the directory level of the
		// particular .gitignore file itself. Otherwise the pattern may also
		// match at any level below the .gitignore level.
		{Pattern: "/a", Path: "/a", Expect: true},
		{Pattern: "a", Path: "/a", Expect: true},
		{Pattern: "a", Path: "/b/a", Expect: true},
		{Pattern: "a", Path: "/c/b/a", Expect: true},
		{Pattern: "/a/b", Path: "/a/b", Expect: true},
		{Pattern: "a/b", Path: "/a/b", Expect: true},
		{Pattern: "a/b", Path: "/d/a/b", Expect: false},

		// If there is a separator at the end of the pattern then the pattern
		// will only match directories, otherwise the pattern can match both
		// files and directories.
		//{Pattern: "/a/b/", Path: "/a/b", Expect: true},

		// For example, a pattern doc/frotz/ matches doc/frotz directory, but
		// not a/doc/frotz directory; however frotz/ matches frotz and a/frotz
		// that is a directory (all paths are relative from the .gitignore file)
		// .
		{Pattern: "doc/frotz/", Path: "/doc/frotz/", Expect: true},
		{Pattern: "doc/frotz/", Path: "/a/doc/frotz", Expect: false},
		{Pattern: "frotz", Path: "/a/frotz", Expect: true},

		// An asterisk "*" matches anything except a slash. The character "?"
		// matches any one character except "/". The range notation, e.g.
		// [a-zA-Z], can be used to match one of the characters in a range. See
		// fnmatch(3) and the  FNM_PATHNAME flag for a more detailed description
		// .
		// @todo

		// Two consecutive asterisks ("**") in patterns matched against full
		// pathname may have special meaning:
		// @todo

		// A leading "**" followed by a slash means match in all directories.
		// For example, "**/foo" matches file or directory "foo" anywhere, the
		// same as pattern "foo". "**/foo/bar" matches file or directory "bar"
		// anywhere that is directly under directory "foo".
		// @todo

		// A trailing "/**" matches everything inside. For example, "abc/**"
		// matches all files inside directory "abc", relative to the location of
		// the .gitignore file, with infinite depth.
		// @todo

		// A slash followed by two consecutive asterisks then a slash matches
		// zero or more directories. For example, "a/**/b" matches "a/b",
		// "a/x/b", "a/x/y/b" and so on.
		// @todo

		// Other consecutive asterisks are considered regular asterisks and will
		// match according to the previous rules.
		// @todo
	} {
		t.Run(fmt.Sprintf("%s with %s", tt.Pattern, tt.Path), func(t *testing.T) {
			actual := IsIgnored(tt.Path, [][]byte{[]byte(tt.Pattern)})
			if actual != tt.Expect {
				t.Errorf("got %v, want %v", actual, tt.Expect)
			}
		})
	}
}
