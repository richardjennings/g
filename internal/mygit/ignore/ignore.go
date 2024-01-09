package ignore

import (
	"github.com/richardjennings/mygit/internal/mygit/config"
	"path/filepath"
	"strings"
)

func IsIgnored(path string) bool {
	// remove absolute portion of Path
	path = strings.TrimPrefix(path, config.Path())
	path = strings.TrimPrefix(path, string(filepath.Separator))
	if path == "" {
		return true
	}
	// @todo fix literal string prefix matching and iteration
	for _, v := range config.Config.GitIgnore {
		if strings.HasPrefix(path, v) {
			return true
		}
	}
	// @todo remove special git case
	if strings.HasPrefix(path, config.DefaultGitDirectory) {
		return true
	}
	if strings.HasPrefix(path, config.Config.GitDirectory) {
		return true
	}
	return false
}
