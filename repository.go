package g

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

// Repository represents a git repository rooted at a specific path.
type Repository struct {
	cnf *Cnf
}

// Open creates a Repository configured with the given options.
func Open(opts ...Opt) (*Repository, error) {
	c := &Cnf{
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
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("applying option: %w", err)
		}
	}
	if c.Path == DefaultPath {
		p, err := filepath.Abs(DefaultPath)
		if err != nil {
			return nil, fmt.Errorf("resolving path: %w", err)
		}
		c.Path = p
	}

	// read .gitignore
	c.GitIgnore = make([][]byte, 0)
	file, err := os.Open(c.GitIgnoreFileName)
	if err == nil {
		defer func() { _ = file.Close() }()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			c.GitIgnore = append(c.GitIgnore, scanner.Bytes())
		}
	}

	return &Repository{cnf: c}, nil
}

// Path helpers

func (r *Repository) Path() string {
	return r.cnf.Path
}

func (r *Repository) gitPath() string {
	return filepath.Join(r.cnf.Path, r.cnf.GitDirectory)
}

func (r *Repository) objectPath() string {
	return filepath.Join(r.cnf.Path, r.cnf.GitDirectory, r.cnf.ObjectsDirectory)
}

func (r *Repository) workingDirectory() string {
	return r.cnf.Path + string(filepath.Separator)
}

func (r *Repository) indexFilePath() string {
	return filepath.Join(r.cnf.Path, r.cnf.GitDirectory, r.cnf.IndexFile)
}

func (r *Repository) refsDirectory() string {
	return filepath.Join(r.cnf.Path, r.cnf.GitDirectory, r.cnf.RefsDirectory)
}

func (r *Repository) refsHeadPrefix() string {
	return filepath.Join(r.cnf.RefsDirectory, r.cnf.RefsHeadsDirectory) + string(os.PathSeparator)
}

func (r *Repository) refsHeadsDirectory() string {
	return filepath.Join(r.cnf.Path, r.cnf.GitDirectory, r.cnf.RefsDirectory, r.cnf.RefsHeadsDirectory)
}

func (r *Repository) packedRefsFile() string {
	return filepath.Join(r.cnf.Path, r.cnf.GitDirectory, r.cnf.PackedRefsFile)
}

func (r *Repository) objectPackfileDirectory() string {
	return filepath.Join(r.cnf.Path, r.cnf.GitDirectory, r.cnf.ObjectsDirectory, r.cnf.PackfileDirectory)
}

func (r *Repository) gitHeadPath() string {
	return filepath.Join(r.cnf.Path, r.cnf.GitDirectory, r.cnf.HeadFile)
}

func (r *Repository) defaultBranch() string {
	return r.cnf.DefaultBranch
}

func (r *Repository) editor() (string, []string) {
	return r.cnf.Editor, r.cnf.EditorArgs
}

func (r *Repository) Editor() (string, []string) {
	return r.editor()
}

func (r *Repository) editorFile() string {
	return fmt.Sprintf("%s/COMMIT_EDITMSG", r.gitPath())
}

func (r *Repository) EditorFile() string {
	return r.editorFile()
}
