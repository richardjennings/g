package mygit

import (
	"errors"
	"fmt"
	"github.com/richardjennings/mygit/pkg/mygit/config"
	"github.com/richardjennings/mygit/pkg/mygit/gfs"
	"github.com/richardjennings/mygit/pkg/mygit/index"
	"github.com/richardjennings/mygit/pkg/mygit/objects"
	"github.com/richardjennings/mygit/pkg/mygit/refs"
	"io"
	"log"
	"os"
	"os/exec"
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
	var updates []*gfs.File
	for _, p := range paths {
		if p == "." {
			// special case meaning add everything
			for _, v := range wdFiles.Files() {
				switch v.WdStatus {
				case gfs.WDUntracked, gfs.WDWorktreeChangedSinceIndex, gfs.WDDeletedInWorktree:
					updates = append(updates, v)
				}
			}
		} else {
			found := false
			for _, v := range wdFiles.Files() {
				if v.Path == p {
					switch v.WdStatus {
					case gfs.WDUntracked, gfs.WDWorktreeChangedSinceIndex, gfs.WDDeletedInWorktree:
						updates = append(updates, v)
					}
					found = true
					break
				}
			}
			if !found {
				// try directory @todo more efficient implementation
				for _, v := range wdFiles.Files() {
					if strings.HasPrefix(v.Path, p+string(filepath.Separator)) {
						switch v.WdStatus {
						case gfs.WDUntracked, gfs.WDWorktreeChangedSinceIndex, gfs.WDDeletedInWorktree:
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
		switch v.WdStatus {
		case gfs.WDUntracked, gfs.WDWorktreeChangedSinceIndex:
			// add the file to the object store
			obj, err := objects.WriteBlob(v.Path)
			if err != nil {
				return err
			}
			v.Sha, _ = gfs.NewSha(obj.Sha)
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
func Commit(message []byte) ([]byte, error) {
	idx, err := index.ReadIndex()
	if err != nil {
		return nil, err
	}
	root := objects.ObjectTree(idx.Files())
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
	commit := &objects.Commit{
		Tree:          tree,
		Parents:       previousCommits,
		Author:        fmt.Sprintf("%s <%s>", config.AuthorName(), config.AuthorEmail()),
		AuthoredTime:  time.Now(),
		Committer:     fmt.Sprintf("%s <%s>", config.CommitterName(), config.CommitterEmail()),
		CommittedTime: time.Now(),
	}
	if message != nil {
		commit.Message = message
	} else {
		// empty commit file
		if err := os.WriteFile(config.EditorFile(), []byte{}, 0600); err != nil {
			log.Fatalln(err)
		}
		ed, args := config.Editor()
		args = append(args, config.EditorFile())
		cmd := exec.Command(ed, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		if err != nil {
			log.Fatalln(err)
		}
		msg, err := os.ReadFile(args[0])
		if err != nil {
			log.Fatalln(msg)
		}
		commit.Message = msg
	}
	if len(commit.Message) == 0 {
		return nil, errors.New("Aborting commit due to empty commit message.")
	}
	return objects.WriteCommit(commit)
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

	files, err := index.Status(idx, commitSha)

	if err != nil {
		return err
	}

	for _, v := range files.Files() {
		if v.IdxStatus == gfs.IndexNotUpdated && v.WdStatus == gfs.WDIndexAndWorkingTreeMatch {
			continue
		}
		if _, err := fmt.Fprintf(o, "%s%s %s\n", v.IdxStatus, v.WdStatus, v.Path); err != nil {
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

	currentCommit, err := refs.LastCommit()
	if err != nil {
		// @todo error types to check for e.g no previous commits as source of error
		return err
	}

	currentStatus, err := index.Status(idx, currentCommit)
	if err != nil {
		return err
	}

	// get commit files
	commitFiles, err := objects.CommittedFiles(commitSha)
	if err != nil {
		return err
	}

	commitSet := gfs.NewFileSet(commitFiles)

	var errorWdFiles []*gfs.File
	var errorIdxFiles []*gfs.File
	var deleteFiles []*gfs.File

	for _, v := range currentStatus.Files() {
		if v.IdxStatus == gfs.IndexUpdatedInIndex {
			errorIdxFiles = append(errorIdxFiles, v)
			continue
		}
		if _, ok := commitSet.Contains(v.Path); ok {
			if v.WdStatus == gfs.WDUntracked {
				errorWdFiles = append(errorWdFiles, v)
				continue
			}
		} else {
			// should be deleted
			deleteFiles = append(deleteFiles, v)
		}
	}
	var errMsg = ""
	if len(errorIdxFiles) > 0 {
		filestr := ""
		for _, v := range errorIdxFiles {
			filestr += fmt.Sprintf("\t%s\n", v.Path)
		}
		errMsg = fmt.Sprintf("error: The following untracked working tree files would be overwritten by checkout:\n %sPlease move or remove them before you switch branches.\nAborting", filestr)
	}
	if len(errorWdFiles) > 0 {
		filestr := ""
		for _, v := range errorWdFiles {
			filestr += fmt.Sprintf("\t%s\n", v.Path)
		}
		if errMsg != "" {
			errMsg += "\n"
		}
		errMsg += fmt.Sprintf("error: The following untracked working tree files would be overwritten by checkout:\n %sPlease move or remove them before you switch branches.\nAborting", filestr)
	}

	if errMsg != "" {
		return errors.New(errMsg)
	}

	for _, v := range deleteFiles {
		if err := os.Remove(filepath.Join(config.Path(), v.Path)); err != nil {
			return err
		}
	}

	idx = index.NewIndex()

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
		v.WdStatus = gfs.WDUntracked
		if err := idx.Add(v); err != nil {
			return err
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

func Restore(path string, staged bool) error {
	idx, err := index.ReadIndex()
	if err != nil {
		return err
	}
	currentCommit, err := refs.LastCommit()
	if err != nil {
		return err
	}
	currentStatus, err := index.Status(idx, currentCommit)
	if err != nil {
		return err
	}
	if staged {
		// remove file from index
		if err := idx.Rm(path); err != nil {
			return err
		}
		return idx.Write()
	}

	fileStatus, ok := currentStatus.Contains(path)
	// if the path not found or is untracked working directory fileStatus then error
	if !ok || fileStatus.WdStatus == gfs.WDUntracked {
		return fmt.Errorf("error: pathspec '%s' did not match any fileStatus(s) known to git", path)
	}
	// if in index but not committed
	if fileStatus.IdxStatus == gfs.IndexAddedInIndex && fileStatus.WdStatus != gfs.WDWorktreeChangedSinceIndex {
		// there is nothing to do, right ? ...
		return nil
	}

	// update working directory fileStatus with object referenced by index
	file := idx.File(path)
	if file == nil {
		// this should not happen
		return errors.New("index did not return file for some reason")
	}
	obj, err := objects.ReadObject(file.Sha.AsHexBytes())
	if err != nil {
		return err
	}
	fh, err := os.OpenFile(filepath.Join(config.Path(), path), os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = fh.Close() }()
	reader, err := obj.ReadCloser()
	if err != nil {
		return err
	}
	if err := objects.ReadHeadBytes(reader, obj); err != nil {
		return err
	}
	_, err = io.Copy(fh, reader)
	if err != nil {
		return err
	}
	if err := fh.Close(); err != nil {
		return err
	}
	return os.Chtimes(filepath.Join(config.Path(), path), file.Finfo.ModTime(), file.Finfo.ModTime())
}
