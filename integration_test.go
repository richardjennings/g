package main

import (
	"github.com/richardjennings/mygit/internal/mygit"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Init(t *testing.T) {
	dir := testDir(t)
	m := testGit(t, dir)
	if err := m.Init(); err != nil {
		t.Fatal(err)
	}
	actual := testListFiles(t, dir)
	expected := []string{".git", ".git/HEAD", ".git/objects", ".git/refs", ".git/refs/heads"}
	assert.Equal(t, expected, actual)
}

func testDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "mygit-test")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func testGit(t *testing.T, path string) *mygit.MyGit {
	opts := []mygit.Opt{
		mygit.WithGitDirectory(mygit.DefaultGitDirectory),
		mygit.WithPath(path),
	}
	m, err := mygit.NewMyGit(opts...)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func testListFiles(t *testing.T, path string) []string {
	var files []string
	if err := filepath.Walk(path, func(p string, info fs.FileInfo, err error) error {
		if p == path {
			return nil
		}
		files = append(files, strings.TrimPrefix(p, path+string(filepath.Separator)))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return files
}
