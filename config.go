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

// config and defaultRepo are the global bridge used during migration.
// They will be removed once all callers use *Repository directly.
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

var defaultRepo = &Repository{cnf: config}

type (
	Cnf struct {
		GitDirectory       string
		Path               string
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

// Configure sets up the global config. Deprecated: use Open() instead.
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
	config.GitIgnore = make([][]byte, 0)
	file, err := os.Open(config.GitIgnoreFileName)
	if err != nil {
		defaultRepo = &Repository{cnf: config}
		return nil
	}
	defer func() { _ = file.Close() }()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		config.GitIgnore = append(config.GitIgnore, scanner.Bytes())
	}
	defaultRepo = &Repository{cnf: config}
	return nil
}

// Free-function shims — delegate to defaultRepo during migration.

func Path() string                       { return defaultRepo.Path() }
func GitPath() string                    { return defaultRepo.gitPath() }
func ObjectPath() string                 { return defaultRepo.objectPath() }
func WorkingDirectory() string           { return defaultRepo.workingDirectory() }
func IndexFilePath() string              { return defaultRepo.indexFilePath() }
func RefsDirectory() string              { return defaultRepo.refsDirectory() }
func RefsHeadPrefix() string             { return defaultRepo.refsHeadPrefix() }
func RefsHeadsDirectory() string         { return defaultRepo.refsHeadsDirectory() }
func PackedRefsFile() string             { return defaultRepo.packedRefsFile() }
func ObjectPackfileDirectory() string    { return defaultRepo.objectPackfileDirectory() }
func GitHeadPath() string                { return defaultRepo.gitHeadPath() }
func DefaultBranch() string              { return defaultRepo.defaultBranch() }
func Editor() (string, []string)         { return defaultRepo.editor() }
func EditorFile() string                 { return defaultRepo.editorFile() }

func Pager() (string, []string) {
	return "/usr/bin/less", []string{"-X", "-F"}
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

// Unused import guard for fmt — used by EditorFile shim via defaultRepo.
var _ = fmt.Sprintf
