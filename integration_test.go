package g

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_Library(t *testing.T) {

	// create a working directory
	dir, err := os.MkdirTemp("", "")
	e(err, t)

	// add a couple of files
	e(os.WriteFile(filepath.Join(dir, "a"), []byte("a"), 0644), t)
	e(os.WriteFile(filepath.Join(dir, "b"), []byte("b"), 0644), t)

	// configure with dir working directory
	e(Configure(WithPath(dir)), t)

	// init, creating a .git folder
	e(Init(), t)

	// create a representation of the repository state
	fs, err := CurrentStatus()
	e(err, t)

	a, ok := fs.Contains("a")
	if !ok {
		e(errors.New("expected file 'a' to be in the current status"), t)
	}

	// check the status of 'a'
	if a.IndexStatus() != IndexUntracked {
		e(errors.New("expected file 'a' to be untracked in the index"), t)
	}

	if a.WorkingDirectoryStatus() != WDUntracked {
		e(errors.New("expected file 'a' to be untracked in the working directory"), t)
	}

	// get the index
	idx, err := ReadIndex()
	e(err, t)

	// add file 'a' to the index and object store
	e(idx.Add(a), t)

	// write the index
	e(idx.Write(), t)

	// create a new representation of the repository state
	fs, err = CurrentStatus()
	e(err, t)

	// get file 'a' again
	a, ok = fs.Contains("a")
	if !ok {
		e(errors.New("expected file 'a' to be in the current status"), t)
	}

	// check the status of 'a'
	if a.IndexStatus() != IndexAddedInIndex {
		e(errors.New("expected file 'a' to be added in the index"), t)
	}

	if a.WorkingDirectoryStatus() != WDIndexAndWorkingTreeMatch {
		e(errors.New("expected file 'a' to have index and working tree match status"), t)
	}

	// create a tree of objects in the index
	tree := ObjectTree(idx.Files())

	// write the tree to the object store
	treeSha, err := tree.WriteTree()

	// check for no previous commits
	pc, err := PreviousCommits()
	e(err, t)
	if pc != nil {
		e(errors.New("expected no previous commits"), t)
	}

	// create a commit
	commit := &Commit{
		Tree:          treeSha,
		Parents:       pc,
		Author:        fmt.Sprintf("%s <%s>", "tester", "tester@test.com"),
		AuthoredTime:  time.Now(),
		Committer:     fmt.Sprintf("%s <%s>", "tester", "tester@test.com"),
		CommittedTime: time.Now(),
		Message:       []byte("this is a commit message"),
	}

	// write the commit
	commitSha, err := WriteCommit(commit)
	e(err, t)
	if !commitSha.IsSet() {
		e(errors.New("expected commit SHA to be set"), t)
	}

	// get current branch
	branch, err := CurrentBranch()
	e(err, t)
	if branch != "main" {
		e(errors.New("expected branch to be 'main'"), t)
	}

	// get current commit sha
	headSha, err := HeadSHA(branch)
	e(err, t)
	if headSha.String() != commitSha.String() {
		e(errors.New("expected head SHA to match previous commit SHA"), t)
	}

	// read the current commit
	commit, err = ReadCommit(headSha)
	e(err, t)

	if string(commit.Message) != "this is a commit message\n" {
		e(errors.New("expected commit message to be 'this is a commit message'"), t)
	}

}

func e(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}
