package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultGitDirectory       = ".git"
	DefaultPath               = "."
	DefaultHeadFile           = "HEAD"
	DefaultIndexFile          = "index"
	DefaultObjectsDirectory   = "objects"
	DefaultRefsDirectory      = "refs"
	DefaultRefsHeadsDirectory = "heads"
	DefaultBranch             = "refs/heads/main"
	DefaultEditor             = "vim"
)

var Config Cnf

type (
	Cnf struct {
		GitDirectory       string
		Path               string
		HeadFile           string
		IndexFile          string
		ObjectsDirectory   string
		RefsDirectory      string
		RefsHeadsDirectory string
		DefaultBranch      string
		GitIgnore          []string
		Editor             string
		EditorArgs         []string
	}
	Opt func(m *Cnf) error
)

func WithPath(path string) Opt {
	return func(c *Cnf) error {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		c.Path = path
		return nil
	}
}

func WithGitDirectory(name string) Opt {
	return func(m *Cnf) error {
		m.GitDirectory = name
		return nil
	}
}

func Configure(opts ...Opt) error {
	c := &Cnf{
		GitDirectory:       DefaultGitDirectory,
		Path:               DefaultPath,
		HeadFile:           DefaultHeadFile,
		IndexFile:          DefaultIndexFile,
		ObjectsDirectory:   DefaultObjectsDirectory,
		RefsDirectory:      DefaultRefsDirectory,
		RefsHeadsDirectory: DefaultRefsHeadsDirectory,
		DefaultBranch:      DefaultBranch,
		Editor:             DefaultEditor,
		GitIgnore: []string{ //@todo read from .gitignore
			".idea/",
		},
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return err
		}
	}
	if c.Path == "" {
		p, err := filepath.Abs(DefaultPath)
		if err != nil {
			return err
		}
		c.Path = p
	}
	Config = *c
	return nil
}

func Path() string {
	return Config.Path
}

func GitPath() string {
	return filepath.Join(Config.Path, Config.GitDirectory)
}

func ObjectPath() string {
	return filepath.Join(Config.Path, Config.GitDirectory, Config.ObjectsDirectory)
}

func WorkingDirectory() string {
	return Config.Path + string(filepath.Separator)
}

func IndexFilePath() string {
	return filepath.Join(Config.Path, Config.GitDirectory, Config.IndexFile)
}

func RefsDirectory() string {
	return filepath.Join(Config.Path, Config.GitDirectory, Config.RefsDirectory)
}

func RefsHeadsDirectory() string {
	return filepath.Join(Config.Path, Config.GitDirectory, Config.RefsDirectory, Config.RefsHeadsDirectory)
}

func GitHeadPath() string {
	return filepath.Join(Config.Path, Config.GitDirectory, Config.HeadFile)
}

func Pager() (string, []string) {
	return "/usr/bin/less", []string{"-X", "-F"}
}

func Editor() (string, []string) {
	return Config.Editor, Config.EditorArgs
}

func EditorFile() string {
	return fmt.Sprintf("%s/COMMIT_EDITMSG", GitPath())
}

func AuthorName() string {
	if v, ok := os.LookupEnv("GIT_AUTHOR_NAME"); ok {
		return v
	}
	return "default"
}

func AuthorEmail() string {
	if v, ok := os.LookupEnv("GIT_AUTHOR_EMAIL"); ok {
		return v
	}
	return "default@default.com"
}

func CommitterName() string {
	if v, ok := os.LookupEnv("GIT_COMMITTER_NAME"); ok {
		return v
	}
	return AuthorName()
}

func CommitterEmail() string {
	if v, ok := os.LookupEnv("GIT_COMMITTER_EMAIL"); ok {
		return v
	}
	return AuthorEmail()
}
