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
	"strings"
	"time"
)

const (
	objectInvalid objectType = iota
	objectBlob
	objectTree
)

type (
	object struct {
		path    string
		typ     objectType
		sha     []byte
		objects []*object
	}
	objectType int
)

func (m *MyGit) writeObject(header []byte, content []byte, contentFile string) ([]byte, error) {
	var f *os.File
	var err error
	h := sha1.New()
	h.Write(header)
	h.Write(content)
	if contentFile != "" {
		f, err = os.Open(contentFile)
		if err != nil {
			return nil, err
		}
		defer func() { _ = f.Close() }()
		if _, err := io.Copy(h, f); err != nil {
			return nil, err
		}
	}
	sha := h.Sum(nil)
	// create object path if needed
	npath := filepath.Join(m.path, m.gitDirectory, ObjectsDirectory, hex.EncodeToString(sha)[0:2])
	if err := os.MkdirAll(npath, 0744); err != nil {
		return nil, err
	}

	// if object exists with sha already we can avoid writing again
	_, err = os.Stat(filepath.Join(npath, hex.EncodeToString(sha)[2:]))
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
	if f != nil {
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

func (m *MyGit) storeTree(fos []*object) ([]byte, error) {
	var content []byte
	var mode string
	for _, fo := range fos {
		// @todo add executable support
		if fo.typ == objectTree {
			mode = "40000"
		} else {
			mode = "100644"
		}
		// @todo replace base..
		content = append(content, []byte(fmt.Sprintf("%s %s%s%s", mode, filepath.Base(fo.path), string(byte(0)), fo.sha))...)
	}
	header := []byte(fmt.Sprintf("tree %d%s", len(content), string(byte(0))))
	return m.writeObject(header, content, "")
}

func (m *MyGit) storeBlob(path string) (*object, error) {
	finfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	header := []byte(fmt.Sprintf("blob %d%s", finfo.Size(), string(byte(0))))
	sha, err := m.writeObject(header, nil, path)
	return &object{sha: sha, path: path}, err
}

// objectTree takes a list of files and generates an object tree like structure
func (m *MyGit) objectTree(files []string) *object {
	root := &object{}
	var n *object  // current node
	var pn *object // previous node
	// mp holds a cache of file paths to objectTree nodes
	mp := make(map[string]*object)
	for _, v := range files {
		parts := strings.Split(strings.TrimPrefix(v, m.path+string(filepath.Separator)), string(filepath.Separator))
		if len(parts) == 1 {
			root.objects = append(root.objects, &object{typ: objectBlob, path: v})
			continue // top level file
		}
		pn = root
		for i, p := range parts {
			if i == len(parts)-1 {
				pn.objects = append(pn.objects, &object{typ: objectBlob, path: v})
				continue // leaf
			}
			// key for cached nodes
			key := strings.Join(parts[0:i+1], string(filepath.Separator))
			cached, ok := mp[key]
			if ok {
				n = cached
			} else {
				n = &object{typ: objectTree, path: p}
				pn.objects = append(pn.objects, n)
				mp[key] = n
			}
			pn = n
		}
	}

	return root
}

func (m *MyGit) writeObjectTree(node *object) ([]byte, error) {
	// resolve child tree objects
	for i, v := range node.objects {
		// @todo the object blobs should already be in the object store having
		// been added to the index previously ...
		if v.typ == objectBlob {
			fo, err := m.storeBlob(filepath.Join(m.path, v.path))
			if err != nil {
				return nil, err
			}
			node.objects[i].sha = fo.sha
		} else if v.typ == objectTree {
			// if the tree only has blobs, write them and then
			// add the corresponding tree returning the sha
			sha, err := m.writeObjectTree(v)
			if err != nil {
				return nil, err
			}
			node.objects[i].sha = sha
		}
	}
	// write a tree object with the resolved children
	return m.storeTree(node.objects)
}
