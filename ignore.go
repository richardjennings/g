package g

import (
	"path/filepath"
	"strings"
)

func IsIgnored(path string) bool {
	// remove absolute portion of Path
	path = strings.TrimPrefix(path, Path())
	path = strings.TrimPrefix(path, string(filepath.Separator))
	if path == "" {
		return true
	}
	// @todo fix literal string prefix matching and iteration
	for _, v := range Config.GitIgnore {
		if strings.HasPrefix(path, v) {
			return true
		}
	}
	return strings.HasPrefix(path, Config.GitDirectory+string(filepath.Separator))
}
