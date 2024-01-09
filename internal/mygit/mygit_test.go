package mygit

import (
	"bytes"
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

func Test_AddFile_Status_Commit(t *testing.T) {
	dir := testDir(t)
	defer func() { _ = os.RemoveAll(dir) }()
	testConfigure(t, dir)
	if err := Init(); err != nil {
		t.Fatal(err)
	}

	// write a file
	if err := os.WriteFile(filepath.Join(dir, "hello"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// status should have an object
	buf := bytes.NewBuffer(nil)
	if err := Status(buf); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, " ?? hello\n", buf.String())

	// add the file to the index
	if err := Add("."); err != nil {
		t.Fatal(err)
	}
	files := testListFiles(t, config.ObjectPath(), false)
	assert.Equal(t, 1, len(files))

	// status should be added
	buf = bytes.NewBuffer(nil)
	if err := Status(buf); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "A  hello\n", buf.String())

	// create commit
	sha, err := Commit()
	if err != nil {
		t.Fatal(err)
	}
	if len(sha) != 20 {
		t.Errorf("expected sha len 20 got %d", len(sha))
	}
	files = testListFiles(t, config.ObjectPath(), false)
	// blob, tree object, commit object
	assert.Equal(t, 3, len(files))
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
