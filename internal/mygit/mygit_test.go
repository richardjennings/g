package mygit

import (
	"bytes"
	"fmt"
	"github.com/richardjennings/mygit/internal/mygit/config"
	"github.com/richardjennings/mygit/internal/mygit/refs"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Init(t *testing.T) {
	dir := testDir(t)
	defer func() { _ = os.RemoveAll(dir) }()
	testConfigure(t, dir)
	if err := Init(); err != nil {
		t.Fatal(err)
	}

	actual := testListFiles(t, dir, true)
	expected := []string{".git", ".git/HEAD", ".git/objects", ".git/refs", ".git/refs/heads"}
	assert.Equal(t, expected, actual)
}

func Test_DefaultBranch(t *testing.T) {
	dir := testDir(t)
	defer func() { _ = os.RemoveAll(dir) }()
	testConfigure(t, dir)
	if err := Init(); err != nil {
		t.Fatal(err)
	}

	actual, err := refs.CurrentBranch()
	assert.NoError(t, err)
	expected := "main"
	assert.Equal(t, expected, actual)
}

func Test_End_To_End(t *testing.T) {
	dir := testDir(t)
	defer func() { _ = os.RemoveAll(dir) }()
	testConfigure(t, dir)

	// git init
	if err := Init(); err != nil {
		t.Fatal(err)
	}

	// list branches - after init there are none
	// git branch
	testBranchLs(t, "")

	// write a file
	// echo "hello" > hello
	if err := os.WriteFile(filepath.Join(dir, "hello"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// status should have an object
	// git status --porcelain
	testStatus(t, "?? hello\n")

	// add the file to the index
	// git add .
	testAdd(t, ".", 1)
	files := testListFiles(t, config.ObjectPath(), false)
	assert.Equal(t, 1, len(files))

	// status should be added
	// git status --porcelain
	testStatus(t, "A  hello\n")

	// create commit
	// git commit -m "test"
	testCommit(t)

	// list branches - main should now show up as it has a commit
	// git branch
	testBranchLs(t, "* main\n")

	files = testListFiles(t, config.ObjectPath(), false)
	// blob, tree object, commit object
	assert.Equal(t, 3, len(files))

	// Test adding a modified file to the index
	// update a file
	// echo "hello world" > hello
	if err := os.WriteFile(filepath.Join(dir, "hello"), []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	// status should be modified
	// git status porcelain
	testStatus(t, " M hello\n")

	// add the file to the index
	// git add hello
	testAdd(t, "hello", 1)

	testStatus(t, "M  hello\n")

	// git commit
	testCommit(t)

	// status should be empty
	// git status --porcelain
	testStatus(t, "")

	// create a branch called test
	// git branch test
	assert.Nil(t, CreateBranch("test"))

	// check it is now listed
	// git branch
	testBranchLs(t, "* main\n  test\n")

	// trying to delete current checkout branch gives error
	// git branch -d main
	err := DeleteBranch("main")
	assert.Equal(t, fmt.Sprintf(DeleteBranchCheckedOutErrFmt, "main", dir), err.Error())

	// delete test branch
	// git branch -d test
	assert.Nil(t, DeleteBranch("test"))

	// should be just main left
	// git branch
	testBranchLs(t, "* main\n")
	testLog(t)

	// create a branch called test2
	// git branch test2
	assert.Nil(t, CreateBranch("test2"))

	// add a file to main and commit
	// echo "world" > world
	if err := os.WriteFile(filepath.Join(dir, "world"), []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}
	// git add world
	testAdd(t, "world", 2)
	// git commit
	testCommit(t)
	// git status --porcelain
	testStatus(t, "")

	// test2 branch does not include world, switch to it and check status
	// git switch test2
	testSwitchBranch(t, "test2")

	// git status --porcelain
	testStatus(t, "")

	// switch back to main, should get file back
	testSwitchBranch(t, "main")
	testStatus(t, "")
}

func testDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "mygit-test")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func testConfigure(t *testing.T, path string) {
	opts := []config.Opt{
		config.WithGitDirectory(config.DefaultGitDirectory),
		config.WithPath(path),
	}
	if err := config.Configure(opts...); err != nil {
		t.Fatal(err)
	}
}

func testListFiles(t *testing.T, path string, dirs bool) []string {
	var files []string
	if err := filepath.Walk(path, func(p string, info fs.FileInfo, err error) error {
		if p == path {
			return nil
		}
		if !dirs && info.IsDir() {
			return nil
		}
		files = append(files, strings.TrimPrefix(p, path+string(filepath.Separator)))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return files
}

func testAdd(t *testing.T, path string, numIdxFiles int) {
	if err := Add(path); err != nil {
		t.Fatal(err)
	}
	files, err := LsFiles()
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, files, numIdxFiles)
}

func testStatus(t *testing.T, expected string) {
	buf := bytes.NewBuffer(nil)
	if err := Status(buf); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expected, buf.String())
}

func testCommit(t *testing.T) []byte {
	sha, err := Commit()
	if err != nil {
		t.Fatal(err)
	}
	if len(sha) != 20 {
		t.Errorf("expected sha len 20 got %d", len(sha))
	}
	return sha
}

func testLog(t *testing.T) []byte {
	buf := bytes.NewBuffer(nil)
	err := Log(buf)
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func testBranchLs(t *testing.T, expected string) {
	buf := bytes.NewBuffer(nil)
	err := ListBranches(buf)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expected, buf.String())
}

func testSwitchBranch(t *testing.T, branch string) {
	if err := SwitchBranch(branch); err != nil {
		t.Fatal(err)
	}
}
