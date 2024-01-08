package objects

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/richardjennings/mygit/internal/mygit/config"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func (o *Object) WriteTree() ([]byte, error) {
	// resolve child tree Objects
	for i, v := range o.Objects {
		if v.Typ == ObjectTree {
			// if the tree only has blobs, write them and then
			// add the corresponding tree returning the Sha
			sha, err := v.WriteTree()
			if err != nil {
				return nil, err
			}
			o.Objects[i].Sha = sha
		}
	}
	// write a tree obj with the resolved children
	return o.writeTree()
}

func (o *Object) writeTree() ([]byte, error) {
	var content []byte
	var mode string
	for _, fo := range o.Objects {
		// @todo add executable support
		if fo.Typ == ObjectTree {
			mode = "40000"
		} else {
			mode = "100644"
		}
		// @todo replace base..
		content = append(content, []byte(fmt.Sprintf("%s %s%s%s", mode, filepath.Base(fo.Path), string(byte(0)), fo.Sha))...)
	}
	header := []byte(fmt.Sprintf("tree %d%s", len(content), string(byte(0))))
	return WriteObject(header, content, "", config.ObjectPath())
}

// WriteObject writes an object
func WriteObject(header []byte, content []byte, contentFile string, path string) ([]byte, error) {
	var f *os.File
	var err error
	buf := bytes.NewBuffer(nil)
	h := sha1.New()
	z := zlib.NewWriter(buf)
	r := io.MultiWriter(h, z)

	if _, err := r.Write(header); err != nil {
		return nil, err
	}
	if len(content) > 0 {
		if _, err := r.Write(content); err != nil {
			return nil, err
		}
	}
	if contentFile != "" {
		f, err = os.Open(contentFile)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(r, f); err != nil {
			return nil, err
		}
		if err := f.Close(); err != nil {
			return nil, err
		}
	}

	sha := h.Sum(nil)
	path = filepath.Join(path, hex.EncodeToString(sha)[:2])
	// create object sha[:2] directory if needed
	if err := os.MkdirAll(path, 0744); err != nil {
		return nil, err
	}
	path = filepath.Join(path, hex.EncodeToString(sha)[2:])
	// if object exists with Sha already we can avoid writing again
	_, err = os.Stat(path)
	if err == nil || !errors.Is(err, fs.ErrNotExist) {
		// file exists
		return sha, err
	}
	if err := z.Close(); err != nil {
		return nil, err
	}
	return sha, os.WriteFile(path, buf.Bytes(), 0655)
}

func WriteBlob(path string) (*Object, error) {
	path = filepath.Join(config.Path(), path)
	finfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	header := []byte(fmt.Sprintf("blob %d%s", finfo.Size(), string(byte(0))))
	sha, err := WriteObject(header, nil, path, config.ObjectPath())
	return &Object{Sha: sha, Path: path}, err
}
