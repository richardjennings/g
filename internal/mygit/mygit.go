package mygit

import (
	"fmt"
	"github.com/richardjennings/mygit/internal/mygit/config"
	"github.com/richardjennings/mygit/internal/mygit/fs"
	"github.com/richardjennings/mygit/internal/mygit/index"
	"github.com/richardjennings/mygit/internal/mygit/objects"
	"github.com/richardjennings/mygit/internal/mygit/refs"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Init initializes a git repository
func Init() error {
	path := config.GitPath()
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	for _, v := range []string{
		config.ObjectPath(),
		config.RefsDirectory(),
		config.RefsHeadsDirectory(),
	} {
		if err := os.MkdirAll(v, 0755); err != nil {
			log.Fatalln(err)
		}
	}
	// set default main branch
	return os.WriteFile(config.GitHeadPath(), []byte(fmt.Sprintf("ref: %s\n", config.Config.DefaultBranch)), 0644)
}

// Log prints out the commit log for the current branch
func Log(o io.Writer) error {
	branch, err := refs.CurrentBranch()
	if err != nil {
		return err
	}
	commitSha, err := refs.HeadSHA(branch)
	if err != nil {
		return err
	}
	for c, err := objects.ReadCommit(commitSha); c != nil && err == nil; c, err = objects.ReadCommit(c.Parents[0]) {
		_, _ = fmt.Fprintf(o, "commit %s\nAuthor: %s <%s>\nDate:   %s\n\n%8s\n", c.Sha, c.Author, c.AuthorEmail, c.AuthoredTime.String(), c.Message)
		if len(c.Parents) == 0 {
			break
		}
	}

	return nil
}

// Add adds one or more file paths to the Index.
func Add(paths ...string) error {
	idx, err := index.ReadIndex()
	if err != nil {
		return err
	}
	// get working directory files with idx status
	wdFiles, err := index.FsStatus(config.Path())
	if err != nil {
		return err
	}
	var updates []*fs.File
	for _, p := range paths {
		if p == "." {
			// special case meaning add everything
			for _, v := range wdFiles {
				switch v.Status {
				case fs.StatusUntracked, fs.StatusModified, fs.StatusDeleted:
					updates = append(updates, v)
				}
			}
		} else {
			found := false
			for _, v := range wdFiles {
				if v.Path == p {
					switch v.Status {
					case fs.StatusUntracked, fs.StatusModified, fs.StatusDeleted:
						updates = append(updates, v)
					}
					found = true
					break
				}
			}
			if !found {
				// try directory @todo more efficient implementation
				for _, v := range wdFiles {
					if strings.HasPrefix(v.Path, p+string(filepath.Separator)) {
						switch v.Status {
						case fs.StatusUntracked, fs.StatusModified, fs.StatusDeleted:
							updates = append(updates, v)
						}
						found = true
					}
				}
			}

			if !found {
				return fmt.Errorf("fatal: pathspec '%s' did not match any files (directories not implemented yet)", p)
			}
		}
	}
	for _, v := range updates {
		switch v.Status {
		case fs.StatusUntracked, fs.StatusModified:
			// add the file to the object store
			obj, err := objects.WriteBlob(v.Path)
			if err != nil {
				return err
			}
			v.Sha, _ = fs.NewSha(obj.Sha)
		}
		if err := idx.Add(v); err != nil {
			return err
		}
	}
	// once all files are added to idx struct, write it out
	return idx.Write()
}

// LsFiles returns a list of files in the index
func LsFiles() ([]string, error) {
	idx, err := index.ReadIndex()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, v := range idx.Files() {
		files = append(files, v.Path)
	}
	return files, nil
}

// Commit writes a git commit object from the files in the index
func Commit() ([]byte, error) {
	idx, err := index.ReadIndex()
	if err != nil {
		return nil, err
	}
	root := index.ObjectTree(idx.Files())
	tree, err := root.WriteTree()
	if err != nil {
		return nil, err
	}
	// git has the --allow-empty flag which here defaults to true currently
	// @todo check for changes to be committed.
	previousCommits, err := refs.PreviousCommits()
	if err != nil {
		// @todo error types to check for e.g no previous commits as source of error
		return nil, err
	}
	return objects.WriteCommit(
		&objects.Commit{
			Tree:          tree,
			Parents:       previousCommits,
			Author:        "Richard Jennings <richardjennings@gmail.com>",
			AuthoredTime:  time.Now(),
			Committer:     "Richard Jennings <richardjennings@gmail.com>",
			CommittedTime: time.Now(),
			Message:       []byte("test"),
		},
	)
}

// Status currently displays the file statuses comparing the working directory
// to the index and the index to the last commit (if any).
func Status(o io.Writer) error {
	var err error
	// index
	idx, err := index.ReadIndex()
	if err != nil {
		return err
	}
	commitSha, err := refs.LastCommit()
	if err != nil {
		// @todo error types to check for e.g no previous commits as source of error
		return err
	}
	files, err := idx.CommitIndexStatus(commitSha)
	if err != nil {
		return err
	}
	for _, v := range files {
		if v.Status == fs.StatusUnchanged {
			continue
		}
		if _, err := fmt.Fprintf(o, "%s  %s\n", v.Status, v.Path); err != nil {
			return err
		}
	}

	// working directory
	files, err = index.FsStatus(config.Path())
	if err != nil {
		return err
	}
	for _, v := range files {
		if v.Status == fs.StatusUnchanged {
			continue
		}
		if _, err := fmt.Fprintf(o, " %s %s\n", v.Status, v.Path); err != nil {
			return err
		}
	}
	return nil
}

const DeleteBranchCheckedOutErrFmt = "error: Cannot delete branch '%s' checked out at '%s'"

func DeleteBranch(name string) error {
	// Delete Branch removes any branch that is not checked out
	// @todo more correct semantics
	currentBranch, err := refs.CurrentBranch()
	if err != nil {
		return err
	}
	if name == currentBranch {
		return fmt.Errorf(DeleteBranchCheckedOutErrFmt, name, config.Path())
	}
	return refs.DeleteBranch(name)
}

func CreateBranch(name string) error {
	return refs.CreateBranch(name)
}

func ListBranches(o io.Writer) error {
	var err error
	currentBranch, err := refs.CurrentBranch()
	if err != nil {
		return err
	}
	branches, err := refs.ListBranches()
	if err != nil {
		return err
	}
	for _, v := range branches {
		if v == currentBranch {
			_, err = o.Write([]byte(fmt.Sprintf("* %v\n", v)))
		} else {
			_, err = o.Write([]byte(fmt.Sprintf("  %v\n", v)))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func SwitchBranch(name string) error {

	// index
	idx, err := index.ReadIndex()
	if err != nil {
		return err
	}

	// get commit sha
	commitSha, err := refs.HeadSHA(name)
	if err != nil {
		return err
	}

	if commitSha == nil {
		return fmt.Errorf("fatal: invalid reference: %s", name)
	}

	// get commit files
	commitFiles, err := objects.CommittedFiles(commitSha)
	if err != nil {
		return err
	}

	// @todo determining if changes would be lost by switching branch ....

	// remove files from working directory not in commit files
	// get working directory files
	files, err := fs.Ls(config.Path())
	if err != nil {
		return err
	}
	// status compared to commit
	files, err = index.CompareAsCommit(commitFiles, files)
	if err != nil {
		return err
	}
	idxFiles := idx.Files()
	idxMap := make(map[string]*fs.File)
	for _, v := range idxFiles {
		idxMap[v.Path] = v
	}
	// delete any files not in branch - and, not in index,
	for _, v := range files {
		if v.Status == fs.StatusAdded {
			if _, ok := idxMap[v.Path]; ok {
				if err := os.Remove(filepath.Join(config.Path(), v.Path)); err != nil {
					return err
				}
			}
		}
	}
	// check for index changes

	files, err = index.CompareAsCommit(commitFiles, idxFiles)
	if err != nil {
		return err
	}
	// swap added status for removed
	for _, v := range files {
		if v.Status == fs.StatusAdded {
			v.Status = fs.StatusDeleted
			// remove from index
			if err := idx.Add(v); err != nil {
				return err
			}
		}
	}

	for _, v := range commitFiles {
		obj, err := objects.ReadObject(v.Sha.AsHexBytes())
		if err != nil {
			return err
		}
		r, err := obj.ReadCloser()
		if err != nil {
			return err
		}
		buf := make([]byte, obj.HeaderLength)
		if _, err := r.Read(buf); err != nil {
			return err
		}
		f, err := os.OpenFile(filepath.Join(config.Path(), v.Path), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0655)
		if err != nil {
			return err
		}

		if _, err := io.Copy(f, r); err != nil {
			return err
		}
		if err := r.Close(); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		// add to index if needed
		inIndex := false
		for _, f := range idxFiles {
			if f.Path == v.Path {
				inIndex = true
				break
			}
		}
		if !inIndex {
			v.Status = fs.StatusUntracked
			if err := idx.Add(v); err != nil {
				return err
			}
		}
	}

	if err := idx.Write(); err != nil {
		return err
	}

	// update HEAD
	if err := refs.UpdateHead(name); err != nil {
		return err
	}

	return nil
}
