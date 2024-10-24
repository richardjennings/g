package git

import (
	"bytes"
	"fmt"
	"github.com/richardjennings/g"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Test_DefaultBranch(t *testing.T) {
	dir := testDir(t)
	defer func() { _ = os.RemoveAll(dir) }()
	testConfigure(t, dir)
	if err := Init(); err != nil {
		t.Fatal(err)
	}

	actual, err := g.CurrentBranch()
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
	writeFile(t, dir, "hello", []byte("hello"))

	// status should have an object
	// git status --porcelain
	testStatus(t, "?? hello\n")

	// add the file to the index
	// git add .
	testAdd(t, ".", 1)
	files := testListFiles(t, g.ObjectPath(), false)
	assert.Equal(t, 1, len(files))

	// status should be added
	// git status --porcelain
	testStatus(t, "A  hello\n")

	// create commit
	// git commit -m "test"
	testCommit(t, []byte("78"))

	// list branches - main should now show up as it has a commit
	// git branch
	testBranchLs(t, "* main\n")

	files = testListFiles(t, g.ObjectPath(), false)
	// blob, tree object, commit object
	assert.Equal(t, 3, len(files))

	// Test adding a modified file to the index
	// update a file
	// echo "hello world" > hello
	writeFile(t, dir, "hello", []byte("hello world"))

	// status should be modified
	// git status porcelain
	testStatus(t, " M hello\n")

	// add the file to the index
	// git add hello
	testAdd(t, "hello", 1)

	testStatus(t, "M  hello\n")

	// git commit
	testCommit(t, []byte("104"))

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
	writeFile(t, dir, "world", []byte("world"))

	// git add world
	testAdd(t, "world", 2)
	// git commit
	testCommit(t, []byte("143"))
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

	// test restore staged
	writeFile(t, dir, "o", []byte("o"))
	testAdd(t, "o", 3)
	testStatus(t, "A  o\n")
	testRestore(t, "o", true)
	testStatus(t, "?? o\n")

	// test restore
	testAdd(t, "o", 3)
	testCommit(t, []byte("oo"))
	testStatus(t, "")
	writeFile(t, dir, "o", []byte("ok"))
	testStatus(t, " M o\n")
	testRestore(t, "o", false)
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
	opts := []g.Opt{
		g.WithGitDirectory(g.DefaultGitDirectory),
		g.WithPath(path),
	}
	if err := g.Configure(opts...); err != nil {
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

func testRestore(t *testing.T, path string, staged bool) {
	if err := Restore(path, staged); err != nil {
		t.Fatal(err)
	}
}

func testCommit(t *testing.T, message []byte) []byte {
	sha, err := Commit(message)
	if err != nil {
		t.Fatal(err)
	}
	if len(sha) != 40 {
		t.Errorf("expected sha len 40 got %d", len(sha))
	}
	commitSha, err := g.NewSha(sha)
	if err != nil {
		t.Error(err)
		return sha
	}

	// read object
	c, err := g.ReadCommit(commitSha)
	if err != nil {
		t.Error(err)
		return sha
	}
	if c == nil {
		t.Error("commit not found")
		return sha
	}
	if string(c.Message) != string(message)+"\n" {
		t.Errorf("expected commit message %s, got %s", string(message), string(c.Message))
		return sha
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

func writeFile(t *testing.T, dir string, path string, content []byte) {
	if err := os.WriteFile(filepath.Join(dir, path), content, 0644); err != nil {
		t.Fatal(err)
	}
	// racy git https://mirrors.edge.kernel.org/pub/software/scm/git/docs/technical/racy-git.txt
	if err := os.Chtimes(filepath.Join(dir, path), time.Now(), time.Now()); err != nil {
		t.Fatal(err)
	}
}
