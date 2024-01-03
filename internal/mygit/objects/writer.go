package objects

import (
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

func WriteObject(header []byte, content []byte, contentFile string, path string) ([]byte, error) {
	var f *os.File
	var err error
	h := sha1.New()
	h.Write(header)
	if len(content) > 0 {
		h.Write(content)
	}
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
	path = filepath.Join(path, hex.EncodeToString(sha)[:2])
	// create object Path if needed
	if err := os.MkdirAll(path, 0744); err != nil {
		return nil, err
	}

	// if object exists with Sha already we can avoid writing again
	_, err = os.Stat(filepath.Join(path, hex.EncodeToString(sha)[2:]))
	if err == nil || !errors.Is(err, fs.ErrNotExist) {
		// file exists
		return sha, err
	}

	tf, err := os.CreateTemp(path, "tmp_obj_")
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
		f, err = os.Open(contentFile)
		if err != nil {
			return nil, err
		}
		defer func() { _ = f.Close() }()
		if _, err := io.Copy(z, f); err != nil {
			return nil, err
		}
	}
	z.Close()
	if err := os.Rename(tf.Name(), filepath.Join(path, hex.EncodeToString(sha)[2:])); err != nil {
		return nil, err
	}
	return sha, nil
}

func StoreBlob(path string) (*Object, error) {
	path = filepath.Join(config.Path(), path)
	finfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	header := []byte(fmt.Sprintf("blob %d%s", finfo.Size(), string(byte(0))))
	sha, err := WriteObject(header, nil, path, config.ObjectPath())
	return &Object{Sha: sha, Path: path}, err
}
