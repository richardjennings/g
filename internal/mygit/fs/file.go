package fs

import (
	"github.com/richardjennings/mygit/internal/mygit/config"
	"github.com/richardjennings/mygit/internal/mygit/ignore"
	"os"
	"path/filepath"
	"strings"
)

const (
	StatusInvalid   FileStatus = iota
	StatusModified             // different in working directory than Index
	StatusUntracked            // in working directory but not in Index
	StatusAdded                // in Index but not in last commit
	StatusDeleted              // in last commit but not in Index
	StatusUnchanged
)

type (
	File struct {
		Path   string
		Status FileStatus
		Sha    []byte
		Finfo  os.FileInfo
	}
	FileStatus uint8
)

func (ist FileStatus) String() string {
	switch ist {
	case StatusModified:
		return "M"
	case StatusAdded:
		return "A"
	case StatusDeleted:
		return "D"
	case StatusUntracked:
		return "??"
	case StatusUnchanged:
		return ""
	default:
		return "x"
	}
}

// Ls recursively lists files in path
func Ls(path string) ([]*File, error) {
	var files []*File
	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		// do not add ignored files
		if !ignore.IsIgnored(path) {
			files = append(files, &File{
				Path:  strings.TrimPrefix(path, config.WorkingDirectory()),
				Finfo: info,
			})
		}
		return nil
	}); err != nil {
		return files, err
	}
	return files, nil
}
