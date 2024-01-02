package mygit

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

const (
	DefaultGitDirectory = ".git"
	DefaultPath         = "."
	DefaultHeadFile     = "HEAD"
	DefaultIndexFile    = "index"
	ObjectsDirectory    = "objects"
	RefsDirectory       = "refs"
	RefsHeadsDirectory  = "heads"
)

type (
	MyGit struct {
		gitDirectory string
		path         string
		gitIgnore    []string
	}
	Opt func(m *MyGit) error
)

func WithPath(path string) Opt {
	return func(m *MyGit) error {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		m.path = path
		return nil
	}
}

func WithGitDirectory(name string) Opt {
	return func(m *MyGit) error {
		m.gitDirectory = name
		return nil
	}
}

func NewMyGit(opts ...Opt) (*MyGit, error) {
	m := &MyGit{}
	m.gitIgnore = []string{ //@todo read from .gitignore
		".idea",
	}
	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// list working directory files that are not ignored
func (m *MyGit) wdFiles() ([]*wdFile, error) {
	var wdFiles []*wdFile
	if err := filepath.Walk(m.path, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		// do not add ignored files
		if !m.isIgnored(path) {
			wdFiles = append(wdFiles, &wdFile{
				path:  strings.TrimPrefix(path, m.path+string(filepath.Separator)),
				finfo: info,
			})
		}
		return nil
	}); err != nil {
		return wdFiles, err
	}
	return wdFiles, nil
}

func (m *MyGit) files() ([]string, error) {
	var files []string
	if err := filepath.Walk(m.path, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		// do not add ignored files
		if !m.isIgnored(path) {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return files, err
	}
	// sort alphanumeric filename
	sort.Slice(files, func(i, j int) bool {
		return files[i] < files[j]
	})
	return files, nil
}

func (m *MyGit) isIgnored(path string) bool {
	// remove absolute portion of path
	path = strings.TrimPrefix(path, m.path)
	path = strings.TrimPrefix(path, string(filepath.Separator))
	if path == "" {
		return true
	}
	// @todo fix literal string prefix matching and iteration
	for _, v := range m.gitIgnore {
		if strings.HasPrefix(path, v) {
			return true
		}
	}
	// @todo remove special git case
	if strings.HasPrefix(path, DefaultGitDirectory) {
		return true
	}
	if strings.HasPrefix(path, m.gitDirectory) {
		return true
	}
	return false
}
