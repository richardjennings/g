package g

import (
	"bufio"
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
	DefaultBranchName         = "main"
	DefaultEditor             = "vim"
	DefaultPackedRefsFile     = "info/refs"
	DefaultPackfileDirectory  = "pack"
	DefaultGitIgnoreFileName  = ".gitignore"
)

var config = &Cnf{
	GitDirectory:       DefaultGitDirectory,
	Path:               DefaultPath,
	HeadFile:           DefaultHeadFile,
	IndexFile:          DefaultIndexFile,
	ObjectsDirectory:   DefaultObjectsDirectory,
	RefsDirectory:      DefaultRefsDirectory,
	RefsHeadsDirectory: DefaultRefsHeadsDirectory,
	PackedRefsFile:     DefaultPackedRefsFile,
	PackfileDirectory:  DefaultPackfileDirectory,
	DefaultBranch:      DefaultBranchName,
	Editor:             DefaultEditor,
	GitIgnoreFileName:  DefaultGitIgnoreFileName,
}

type (
	Cnf struct {
		// GitDirector configures where the name of the git directory
		// This is usually .git
		GitDirectory string
		// Path configures where the Git Directory to interact with is
		// relative to the present working directory. This is usually .
		Path string

		HeadFile           string
		IndexFile          string
		ObjectsDirectory   string
		RefsDirectory      string
		RefsHeadsDirectory string
		PackedRefsFile     string
		PackfileDirectory  string
		DefaultBranch      string
		GitIgnore          [][]byte
		Editor             string
		EditorArgs         []string
		GitIgnoreFileName  string
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

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return err
		}
	}
	if config.Path == DefaultPath {
		p, err := filepath.Abs(DefaultPath)
		if err != nil {
			return err
		}
		config.Path = p
	}

	// read .gitignore
	// @todo there can be multiple, and some of the rules are relative to those
	// files ...
	config.GitIgnore = make([][]byte, 0)
	file, err := os.Open(config.GitIgnoreFileName)
	if err != nil {
		return nil
	}
	defer func() { _ = file.Close() }()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		config.GitIgnore = append(config.GitIgnore, scanner.Bytes())
	}
	return nil
}

func Path() string {
	return config.Path
}

func GitPath() string {
	return filepath.Join(config.Path, config.GitDirectory)
}

func ObjectPath() string {
	return filepath.Join(config.Path, config.GitDirectory, config.ObjectsDirectory)
}

func WorkingDirectory() string {
	return config.Path + string(filepath.Separator)
}

func IndexFilePath() string {
	return filepath.Join(config.Path, config.GitDirectory, config.IndexFile)
}

func RefsDirectory() string {
	return filepath.Join(config.Path, config.GitDirectory, config.RefsDirectory)
}

func RefsHeadPrefix() string {
	return filepath.Join(config.RefsDirectory, config.RefsHeadsDirectory) + string(os.PathSeparator)
}

func RefsHeadsDirectory() string {
	return filepath.Join(config.Path, config.GitDirectory, config.RefsDirectory, config.RefsHeadsDirectory)
}

func PackedRefsFile() string {
	return filepath.Join(config.Path, config.GitDirectory, config.PackedRefsFile)
}

func ObjectPackfileDirectory() string {
	return filepath.Join(config.Path, config.GitDirectory, config.ObjectsDirectory, config.PackfileDirectory)
}

func GitHeadPath() string {
	return filepath.Join(config.Path, config.GitDirectory, config.HeadFile)
}

func Pager() (string, []string) {
	return "/usr/bin/less", []string{"-X", "-F"}
}

func Editor() (string, []string) {
	return config.Editor, config.EditorArgs
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

func DefaultBranch() string {
	return config.DefaultBranch
}
