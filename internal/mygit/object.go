package mygit

import (
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type (
	fileObject struct {
		path string
		sha  []byte
	}
)

func (m *MyGit) writeObject(header []byte, content []byte, contentFile string) ([]byte, error) {
	h := sha1.New()
	h.Write(header)
	h.Write(content)
	sha := h.Sum(nil)
	// create object path if needed
	npath := filepath.Join(m.path, m.gitDirectory, ObjectsDirectory, hex.EncodeToString(sha)[0:2])
	if err := os.MkdirAll(npath, 0744); err != nil {
		return nil, err
	}

	// if object exists with sha already we can avoid writing again
	_, err := os.Stat(filepath.Join(npath, hex.EncodeToString(sha)[2:]))
	if err == nil || !errors.Is(err, fs.ErrNotExist) {
		// file exists
		return sha, err
	}

	tf, err := os.CreateTemp(npath, "tmp_obj_")
	if err != nil {
		return nil, err
	}
	defer func() { _ = tf.Close() }()
	z := zlib.NewWriter(tf)
	defer func() { _ = z.Close() }()
	if _, err := z.Write(header); err != nil {
		return nil, err
	}
	if _, err := z.Write(content); err != nil {
		return nil, err
	}
	if contentFile != "" {
		f, err := os.Open(contentFile)
		if err != nil {
			return nil, err
		}
		defer func() { _ = f.Close() }()
		if _, err := io.Copy(z, f); err != nil {
			return nil, err
		}
	}
	if err := os.Rename(tf.Name(), filepath.Join(npath, hex.EncodeToString(sha)[2:])); err != nil {
		return nil, err
	}
	return sha, nil
}

func (m *MyGit) storeCommit(treeSha []byte, parents [][]byte, author string, committer string, msg string) ([]byte, error) {
	var parentCommits string
	for _, v := range parents {
		parentCommits += fmt.Sprintf("parent %s\n", v)
	}
	content := []byte(fmt.Sprintf(
		"tree %s\n%sauthor %s %d +0000\ncommitter %s %d +0000\n\n%s",
		hex.EncodeToString(treeSha),
		parentCommits,
		author,
		time.Now().Unix(),
		committer,
		time.Now().Unix(),
		msg,
	))
	header := []byte(fmt.Sprintf("commit %d%s", len(content), string(byte(0))))
	return m.writeObject(header, content, "")
}

func (m *MyGit) storeTree(fos []*fileObject) ([]byte, error) {
	// sort fileObjects
	sort.Slice(fos, func(i, j int) bool {
		return fos[i].path < fos[j].path
	})
	var content []byte
	for _, fo := range fos {
		// @todo replace base..
		content = append(content, []byte(fmt.Sprintf("100644 %s%s%s", filepath.Base(fo.path), string(byte(0)), fo.sha))...)
	}
	header := []byte(fmt.Sprintf("tree %d%s", len(content), string(byte(0))))
	return m.writeObject(header, content, "")
}

func (m *MyGit) storeBlob(path string) (*fileObject, error) {
	finfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	header := []byte(fmt.Sprintf("blob %d%s", finfo.Size(), string(byte(0))))
	sha, err := m.writeObject(header, nil, path)
	return &fileObject{sha: sha, path: path}, err
}
