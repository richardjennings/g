package g

import (
	"bytes"
	"fmt"
	"strings"
)

func IsIgnored(path string, rules [][]byte) bool {

	// make the path relative
	path = strings.TrimPrefix(path, Path())

	// ignore the git directory regardless
	if strings.HasPrefix(path, fmt.Sprintf("/%s/", config.GitDirectory)) {
		return true
	}

	for _, v := range rules {
		prefixMatch := false
		dirMatch := false

		if len(v) == 0 {
			// blank line separator
			continue
		}
		if v[0] == '#' {
			// comment
			continue
		}
		if v[0] == '\\' && v[1] == '#' {
			// escaped #
			v = v[1:]
		}
		for i, vv := range v {
			// allow escaped spaces
			if vv == '\\' && v[i+1] == ' ' && i < len(v)-1 {
				v = append(v[:i], v[:i+1]...)
			}
		}
		if v[0] == '/' {
			// starts with /
			prefixMatch = true
		} else if l := bytes.LastIndex(v, []byte{'/'}); l > -1 && (l < len(v)-1 || l == 0) {
			// of has / in it but not the end
			prefixMatch = true
			// easier to add explicit '/'
			v = append([]byte{'/'}, v...)
		}
		if v[len(v)-1] == '/' {
			dirMatch = true
		}

		// check for suffix match
		if !prefixMatch && !dirMatch {
			// other things to check before using suffix ...
			if bytes.HasSuffix([]byte(path), v) {
				return true
			}
		}

		if dirMatch {
			if path[len(path)-1] == '/' {
				return true
			}
		}

		if prefixMatch {
			if bytes.HasPrefix([]byte(path), v) {
				return true
			}
		}
	}
	return false
}
