package mygit

import (
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
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

func (m *MyGit) writeObject(header []byte, content []byte) ([]byte, error) {
	h := sha1.New()
	h.Write(header)
	h.Write(content)
	sha := h.Sum(nil)
	// create object path if needed
	npath := filepath.Join(m.path, m.gitDirectory, ObjectsDirectory, hex.EncodeToString(sha)[0:2])
	if err := os.MkdirAll(npath, 0744); err != nil {
		return nil, err
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
	if err := os.Rename(tf.Name(), filepath.Join(npath, hex.EncodeToString(sha)[2:])); err != nil {
		return nil, err
	}
	return sha, nil
}

func (m *MyGit) storeCommit(treeSha []byte, parents [][]byte, author string, committer string, msg string) ([]byte, error) {
	content := []byte(fmt.Sprintf(
		"tree %s\nauthor %s %d +0000\ncommitter %s %d +0000\n\n%s",
		hex.EncodeToString(treeSha),
		author,
		time.Now().Unix(),
		committer,
		time.Now().Unix(),
		msg,
	))
	header := []byte(fmt.Sprintf("commit %d%s", len(content), string(byte(0))))
	return m.writeObject(header, content)
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
	return m.writeObject(header, content)
}

func (m *MyGit) storeBlob(path string) (*fileObject, error) {
	// @todo adapt into writeObjectFile or similar
	finfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	prefix := fmt.Sprintf("blob %d%s", finfo.Size(), string(byte(0)))
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	h := sha1.New()
	h.Write([]byte(prefix))
	_, err = io.Copy(h, f)
	_ = f.Close()
	if err != nil {
		return nil, err
	}
	sha := h.Sum(nil)

	// create object path if needed
	npath := filepath.Join(m.path, m.gitDirectory, ObjectsDirectory, hex.EncodeToString(sha)[0:2])
	if err := os.MkdirAll(npath, 0744); err != nil {
		return nil, err
	}
	// write temporary file
	tf, err := os.CreateTemp(npath, "tmp_obj_")
	if err != nil {
		return nil, err
	}
	defer func() { _ = tf.Close() }()
	z := zlib.NewWriter(tf)
	z.Write([]byte(prefix))
	defer z.Close()
	f, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	if _, err = io.Copy(z, f); err != nil {
		return nil, err
	}

	return &fileObject{sha: sha, path: path}, os.Rename(tf.Name(), filepath.Join(npath, hex.EncodeToString(sha)[2:]))
}
