package g

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_Integration(t *testing.T) {

	// create a working directory
	dir, err := os.MkdirTemp("", "")
	e(err, t)

	defer func() { _ = os.RemoveAll(dir) }()

	// add a couple of files
	e(os.WriteFile(filepath.Join(dir, "a"), []byte("a"), 0644), t)
	e(os.WriteFile(filepath.Join(dir, "b"), []byte("b"), 0644), t)

	// configure with dir working directory
	e(Configure(WithPath(dir)), t)

	// init, creating a .git folder
	e(Init(), t)

	// commit 'a' to branch 'main'
	{
		// check branch is main
		assertCurrentBranch(t, "main")
		// check that file 'a' is recorded as untracked in the current status,
		assertStatus(t, map[string]IndexStatus{"a": UntrackedInIndex}, map[string]WDStatus{"a": Untracked})
		// add 'a' to the index
		assertAddFiles(t, []string{"a"})
		// check its status is correct
		assertStatus(t, map[string]IndexStatus{"a": AddedInIndex}, map[string]WDStatus{"a": IndexAndWorkingTreeMatch})
		// create a commit
		commitSha := assertCreateCommit(t, &Commit{
			Author:        fmt.Sprintf("%s <%s>", "tester", "tester@test.com"),
			AuthoredTime:  time.Now(),
			Committer:     fmt.Sprintf("%s <%s>", "tester", "tester@test.com"),
			CommittedTime: time.Now(),
			Message:       []byte("this is a commit message"),
		})
		// check branch is still main
		assertCurrentBranch(t, "main")
		// check current HEAD sha matches the new commit sha
		assertCurrentCommit(t, commitSha)
		// check last commit message matches the created commit message
		assertLookupCommit(t, commitSha, func(t *testing.T, commit *Commit) {
			if string(commit.Message) != "this is a commit message\n" {
				t.Errorf("expected commit message to be 'this is a commit message'")
			}
		})

	}

	// modify 'a',
	// commit,
	// create a branch 'test',
	// switch to branch 'test',
	// check file 'b' is still present untracked
	{
		// modify file a
		e(os.WriteFile(filepath.Join(dir, "a"), []byte("aa"), 0644), t)
		// add 'a' to the index
		assertAddFiles(t, []string{"a"})
		// create a commit
		commitSha := assertCreateCommit(t, &Commit{
			Author:        fmt.Sprintf("%s <%s>", "tester", "tester@test.com"),
			AuthoredTime:  time.Now(),
			Committer:     fmt.Sprintf("%s <%s>", "tester", "tester@test.com"),
			CommittedTime: time.Now(),
			Message:       []byte("this is a another commit message"),
		})
		// check last commit message matches the created commit message
		assertLookupCommit(t, commitSha, func(t *testing.T, commit *Commit) {
			if string(commit.Message) != "this is a another commit message\n" {
				t.Errorf("expected commit message to be 'this is a another commit message'")
			}
		})
		// check 'a' is still correct
		assertStatus(t, map[string]IndexStatus{"a": NotUpdated}, map[string]WDStatus{"a": IndexAndWorkingTreeMatch})
		// change to a new branch
		e(CreateBranch("test"), t)
		// switch to the new branch
		assertSwitchBranch(t, "test", assertNoErrorFiles)
		// check branch is now test
		assertCurrentBranch(t, "test")
		// check current commit is the last commit sha
		assertCurrentCommit(t, commitSha)
		// check 'a' is still correct
		assertStatus(t, map[string]IndexStatus{"a": NotUpdated}, map[string]WDStatus{"a": IndexAndWorkingTreeMatch})
		// check 'b' is still correct
		assertStatus(t, map[string]IndexStatus{"b": UntrackedInIndex}, map[string]WDStatus{"b": Untracked})
	}

	// commit b
	// switch to 'main' branch
	// check 'b' is no longer in the working directory
	{
		// add 'b' to the index
		assertAddFiles(t, []string{"b"})
		// commit
		commitSha := assertCreateCommit(t, &Commit{
			Author:        fmt.Sprintf("%s <%s>", "tester", "tester@test.com"),
			AuthoredTime:  time.Now(),
			Committer:     fmt.Sprintf("%s <%s>", "tester", "tester@test.com"),
			CommittedTime: time.Now(),
			Message:       []byte("this is yet another commit message"),
		})
		// check commit
		assertLookupCommit(t, commitSha, func(t *testing.T, commit *Commit) {
			if string(commit.Message) != "this is yet another commit message\n" {
				t.Errorf("expected commit message to be 'this is yet another commit message'")
			}
		})
		// check status of 'b'
		assertStatus(t, map[string]IndexStatus{"b": NotUpdated}, map[string]WDStatus{"b": IndexAndWorkingTreeMatch})
		// switch to the main branch
		assertSwitchBranch(t, "main", assertNoErrorFiles)
		// check file 'b' is not in the status
		assertNotInStatus(t, []string{"b"})
	}

	// switch back to branch 'test'
	// add a new file 'c'
	// add 'c' to the index
	// switch back to 'main'
	// 'c' should still be in the index
	{
		// switch back to branch 'test'
		assertSwitchBranch(t, "test", assertNoErrorFiles)
		// create a new file 'c'
		e(os.WriteFile(filepath.Join(dir, "c"), []byte("c"), 0644), t)
		// check 'c' has the correct status
		assertStatus(t, map[string]IndexStatus{"c": UntrackedInIndex}, map[string]WDStatus{"c": Untracked})
		// add 'c' to the index
		assertAddFiles(t, []string{"c"})
		// switch to branch 'main'
		assertSwitchBranch(t, "main", assertNoErrorFiles)
		// check 'c' has the correct status
		assertStatus(t, map[string]IndexStatus{"c": AddedInIndex}, map[string]WDStatus{"c": IndexAndWorkingTreeMatch})
	}

	// check for semantics when switching branch
	//

}

func e(err error, t *testing.T) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertLookupCommit(t *testing.T, sha Sha, f func(*testing.T, *Commit)) {
	t.Helper()
	commit, err := ReadCommit(sha)
	e(err, t)
	f(t, commit)
}

func assertCurrentCommit(t *testing.T, commitSha Sha) {
	t.Helper()
	branch, err := CurrentBranch()
	e(err, t)
	sha, err := HeadSHA(branch)
	e(err, t)
	if sha.String() != commitSha.String() {
		e(fmt.Errorf("expected current commit SHA %s to match SHA %s", sha, commitSha), t)
	}
}

func assertCreateCommit(t *testing.T, commit *Commit) Sha {
	t.Helper()
	commitSha, err := CreateCommit(commit)
	e(err, t)
	if !commitSha.IsSet() {
		e(errors.New("expected commit SHA to be set"), t)
	}
	return commitSha
}

func assertAddFiles(t *testing.T, filePaths []string) {
	t.Helper()
	idx, err := ReadIndex()
	e(err, t)
	fh, err := CurrentStatus()
	e(err, t)
	for _, v := range filePaths {
		f, ok := fh.idx[v]
		if !ok {
			t.Errorf("file %s not found in index", v)
		}
		e(idx.addFromWorkTree(f), t)
	}
	e(idx.Write(), t)
}

func assertCurrentBranch(t *testing.T, expected string) {
	t.Helper()
	branch, err := CurrentBranch()
	e(err, t)
	if branch != expected {
		t.Errorf("expected branch to be '%s' got '%s'", expected, branch)
	}
}

func assertNotInStatus(t *testing.T, filePaths []string) {
	t.Helper()
	fh, err := CurrentStatus()
	e(err, t)
	for _, v := range filePaths {
		if _, ok := fh.idx[v]; ok {
			t.Errorf("file %s was not expected to be in status", v)
		}
	}
}

var assertNoErrorFiles = func(t *testing.T, paths []string) {
	if len(paths) > 0 {
		t.Errorf("expected 0 files, got %d", len(paths))
	}
}

func assertSwitchBranch(t *testing.T, name string, f func(t *testing.T, fh []string)) {
	t.Helper()
	errFiles, err := SwitchBranch(name)
	if err != nil {
		t.Error(err)
	}
	f(t, errFiles)
}

func assertStatus(t *testing.T, i map[string]IndexStatus, w map[string]WDStatus) {
	t.Helper()
	fs, err := CurrentStatus()
	e(err, t)
	for k, v := range i {
		f, ok := fs.Contains(k)
		if !ok {
			t.Fatalf("expected file '%s'", k)
		}
		if f.idxStatus != v {
			t.Errorf("expected file '%s' to have index status '%s' got '%s'", k, v, f.idxStatus)
		}
	}
	for k, v := range w {
		f, ok := fs.Contains(k)
		if !ok {
			t.Fatalf("expected file '%s'", k)
		}
		if f.wdStatus != v {
			t.Errorf("expected file '%s' to have working directory status '%d' got '%d'", k, v, f.wdStatus)
		}
	}
}
