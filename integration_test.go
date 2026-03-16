package g

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	e(initRepo(), t)

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

	// restore --staged a file that is commited, has been modified, added to
	// index
	// restore a working tree file that is in a previous commit
	{
		// commit change to 'c'
		_ = assertCreateCommit(t, &Commit{
			Author:        fmt.Sprintf("%s <%s>", "tester", "tester@test.com"),
			AuthoredTime:  time.Now(),
			Committer:     fmt.Sprintf("%s <%s>", "tester", "tester@test.com"),
			CommittedTime: time.Now(),
			Message:       []byte("this is yet another commit message"),
		})
		// change 'c'
		e(os.WriteFile(filepath.Join(dir, "c"), []byte("cc"), 0644), t)
		// check status is correct
		assertStatus(t, map[string]IndexStatus{"c": NotUpdated}, map[string]WDStatus{"c": WorktreeChangedSinceIndex})
		// add to index
		assertAddFiles(t, []string{"c"})
		// check status
		assertStatus(t, map[string]IndexStatus{"c": UpdatedInIndex}, map[string]WDStatus{"c": IndexAndWorkingTreeMatch})
		// restore --staged c
		assertRestore(t, "c", true)
		// check status
		assertStatus(t, map[string]IndexStatus{"c": NotUpdated}, map[string]WDStatus{"c": WorktreeChangedSinceIndex})
		// git restore c
		assertRestore(t, "c", false)
		// check status
		assertStatus(t, map[string]IndexStatus{"c": NotUpdated}, map[string]WDStatus{"c": IndexAndWorkingTreeMatch})
	}

}

func e(err error, t *testing.T) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertRestore(t *testing.T, path string, staged bool) {
	if err := Restore(path, staged); err != nil {
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
			t.Errorf("expected file '%s' to have working directory status '%s' got '%s'", k, v, f.wdStatus)
		}
	}
}

// initTestRepo creates a temp directory, configures g, and calls Init.
// It returns the temp dir path and a cleanup function.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "g-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	t.Setenv("GIT_AUTHOR_NAME", "tester")
	t.Setenv("GIT_AUTHOR_EMAIL", "tester@test.com")

	e(Configure(WithPath(dir), WithGitDirectory(".git")), t)
	e(initRepo(), t)
	return dir
}

func makeCommit(t *testing.T, msg string) Sha {
	t.Helper()
	return assertCreateCommit(t, &Commit{
		Author:        fmt.Sprintf("%s <%s>", "tester", "tester@test.com"),
		AuthoredTime:  time.Now(),
		Committer:     fmt.Sprintf("%s <%s>", "tester", "tester@test.com"),
		CommittedTime: time.Now(),
		Message:       []byte(msg),
	})
}

// Test_Restore_NonExistentFile tests that restoring a file that does not exist
// returns an error.
func Test_Restore_NonExistentFile(t *testing.T) {
	_ = initTestRepo(t)

	// create and commit a file so the repo has at least one commit
	err := Restore("nonexistent.txt", false)
	if err == nil {
		t.Fatal("expected error when restoring a non-existent file, got nil")
	}
}

// Test_Restore_UntrackedFile tests that restoring an untracked file returns a
// pathspec error.
func Test_Restore_UntrackedFile(t *testing.T) {
	dir := initTestRepo(t)

	// create a file but do NOT add it to the index
	e(os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("hello"), 0644), t)

	err := Restore("untracked.txt", false)
	if err == nil {
		t.Fatal("expected error when restoring an untracked file, got nil")
	}
	if !strings.Contains(err.Error(), "pathspec") {
		t.Errorf("expected pathspec error message, got: %s", err.Error())
	}
}

// Test_RestoreStaged_CommittedFile tests the RestoreStaged upsert path where
// the file IS in a previous commit. This differs from the Rm path (tested
// in the main integration test) where the file is only added, not committed.
func Test_RestoreStaged_CommittedFile(t *testing.T) {
	dir := initTestRepo(t)

	// create and commit a file
	e(os.WriteFile(filepath.Join(dir, "x"), []byte("original"), 0644), t)
	assertAddFiles(t, []string{"x"})
	makeCommit(t, "initial commit with x")

	// verify committed status
	assertStatus(t,
		map[string]IndexStatus{"x": NotUpdated},
		map[string]WDStatus{"x": IndexAndWorkingTreeMatch},
	)

	// modify the file, add it to the index (UpdatedInIndex)
	e(os.WriteFile(filepath.Join(dir, "x"), []byte("modified"), 0644), t)
	assertAddFiles(t, []string{"x"})
	assertStatus(t,
		map[string]IndexStatus{"x": UpdatedInIndex},
		map[string]WDStatus{"x": IndexAndWorkingTreeMatch},
	)

	// restore --staged x (upsert path: file exists in commit)
	assertRestore(t, "x", true)

	// after restore --staged, the index should reflect the committed version
	// while the working tree still has the modification
	assertStatus(t,
		map[string]IndexStatus{"x": NotUpdated},
		map[string]WDStatus{"x": WorktreeChangedSinceIndex},
	)
}

// Test_SwitchBranch_WithConflicts tests that switching branches when there are
// local modifications to a file that differs between branches returns the
// conflicting file list.
func Test_SwitchBranch_WithConflicts(t *testing.T) {
	dir := initTestRepo(t)

	// create and commit a file on main
	e(os.WriteFile(filepath.Join(dir, "conflict.txt"), []byte("main-version"), 0644), t)
	assertAddFiles(t, []string{"conflict.txt"})
	makeCommit(t, "commit on main")

	// create a new branch and switch to it
	e(CreateBranch("feature"), t)
	assertSwitchBranch(t, "feature", assertNoErrorFiles)

	// modify and commit the file on the feature branch
	e(os.WriteFile(filepath.Join(dir, "conflict.txt"), []byte("feature-version"), 0644), t)
	assertAddFiles(t, []string{"conflict.txt"})
	makeCommit(t, "commit on feature")

	// switch back to main
	assertSwitchBranch(t, "main", assertNoErrorFiles)

	// modify the file locally (without committing) so it conflicts with feature
	e(os.WriteFile(filepath.Join(dir, "conflict.txt"), []byte("local-change"), 0644), t)

	// try switching to feature -- should report conflicting files
	errFiles, err := SwitchBranch("feature")
	if err != nil {
		t.Fatalf("expected no error from SwitchBranch, got: %v", err)
	}
	if len(errFiles) == 0 {
		t.Fatal("expected non-empty error file list when switching with conflicts")
	}

	found := false
	for _, f := range errFiles {
		if f == "conflict.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'conflict.txt' in conflict list, got: %v", errFiles)
	}
}

// Test_NestedDirectories_WriteTree tests that files in nested subdirectories
// are correctly handled by ObjectTree and WriteTree, and that the commit
// round-trips properly.
func Test_NestedDirectories_WriteTree(t *testing.T) {
	dir := initTestRepo(t)

	// create nested directory structure
	e(os.MkdirAll(filepath.Join(dir, "a", "b", "c"), 0755), t)
	e(os.WriteFile(filepath.Join(dir, "top.txt"), []byte("top"), 0644), t)
	e(os.WriteFile(filepath.Join(dir, "a", "mid.txt"), []byte("mid"), 0644), t)
	e(os.WriteFile(filepath.Join(dir, "a", "b", "deep.txt"), []byte("deep"), 0644), t)
	e(os.WriteFile(filepath.Join(dir, "a", "b", "c", "deepest.txt"), []byte("deepest"), 0644), t)

	// add all files
	assertAddFiles(t, []string{"top.txt"})
	assertAddFiles(t, []string{filepath.Join("a", "mid.txt")})
	assertAddFiles(t, []string{filepath.Join("a", "b", "deep.txt")})
	assertAddFiles(t, []string{filepath.Join("a", "b", "c", "deepest.txt")})

	// commit
	commitSha := makeCommit(t, "nested directories commit")

	// verify the commit is readable
	assertLookupCommit(t, commitSha, func(t *testing.T, commit *Commit) {
		if !strings.Contains(string(commit.Message), "nested directories commit") {
			t.Errorf("unexpected commit message: %s", commit.Message)
		}
	})

	// read back the committed files via the object tree
	committedFiles, err := CommittedFiles(commitSha)
	if err != nil {
		t.Fatalf("failed to read committed files: %v", err)
	}

	expectedPaths := map[string]bool{
		"top.txt":                                    false,
		filepath.Join("a", "mid.txt"):                false,
		filepath.Join("a", "b", "deep.txt"):          false,
		filepath.Join("a", "b", "c", "deepest.txt"):  false,
	}

	for _, f := range committedFiles {
		if _, ok := expectedPaths[f.Path()]; ok {
			expectedPaths[f.Path()] = true
		}
	}

	for p, found := range expectedPaths {
		if !found {
			t.Errorf("expected path '%s' in committed files but not found", p)
		}
	}
}

// Test_Init_Idempotent tests that calling Init twice does not produce an error.
func Test_Init_Idempotent(t *testing.T) {
	dir, err := os.MkdirTemp("", "g-test-init-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	t.Setenv("GIT_AUTHOR_NAME", "tester")
	t.Setenv("GIT_AUTHOR_EMAIL", "tester@test.com")

	e(Configure(WithPath(dir), WithGitDirectory(".git")), t)

	// first init
	e(initRepo(), t)

	// verify .git directory exists
	info, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		t.Fatalf("expected .git directory to exist after first Init: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected .git to be a directory")
	}

	// second init -- should not error
	e(initRepo(), t)

	// verify .git directory still exists and is intact
	info, err = os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		t.Fatalf("expected .git directory to exist after second Init: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected .git to be a directory after second Init")
	}

	// verify we can still use the repo after double init
	e(os.WriteFile(filepath.Join(dir, "test.txt"), []byte("test"), 0644), t)
	assertAddFiles(t, []string{"test.txt"})
	makeCommit(t, "commit after double init")
	assertCurrentBranch(t, "main")
}
