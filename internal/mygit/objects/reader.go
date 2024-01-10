package objects

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/richardjennings/mygit/internal/mygit/config"
	"github.com/richardjennings/mygit/internal/mygit/fs"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

// FlattenTree turns a TreeObject structure into a flat list of file paths
func (o *Object) FlattenTree() []*fs.File {
	var objFiles []*fs.File
	if o.Typ == ObjectBlob {
		return []*fs.File{{Path: o.Path, Sha: o.Sha}}
	}
	for _, v := range o.Objects {
		objs := v.FlattenTree()
		for i := range objs {
			objs[i].Path = filepath.Join(o.Path, objs[i].Path)
		}
		objFiles = append(objFiles, objs...)
	}

	return objFiles
}

func ReadObject(sha []byte) (*Object, error) {
	var err error
	o := &Object{Sha: sha}
	o.ReadCloser = ObjectReadCloser(sha)
	z, err := o.ReadCloser()
	if err != nil {
		return o, err
	}
	defer func() { _ = z.Close() }()
	buf := bufio.NewReader(z)
	p, err := buf.ReadBytes(0)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	o.HeaderLength = len(p)
	header := bytes.Fields(p)

	switch string(header[0]) {
	case "commit":
		o.Typ = ObjectCommit
	case "tree":
		o.Typ = ObjectTree
	case "blob":
		o.Typ = ObjectBlob
	default:
		return nil, fmt.Errorf("unknown %s", string(header[0]))
	}
	o.Length, err = strconv.Atoi(string(header[1][:len(header[1])-1]))
	return o, err
}

func ObjectReadCloser(sha []byte) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) {
		path := filepath.Join(config.ObjectPath(), string(sha[0:2]), string(sha[2:]))
		f, err := os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			return nil, err
		}
		defer func() { _ = f.Close() }()
		return zlib.NewReader(f)
	}
}

// ReadObjectTree reads an object from the object store
func ReadObjectTree(sha []byte) (*Object, error) {
	obj, err := ReadObject(sha)
	if err != nil {
		return nil, err
	}
	switch obj.Typ {
	case ObjectCommit:
		commit, err := readCommit(obj)
		if err != nil {
			return obj, err
		}
		co, err := ReadObjectTree(commit.Tree)
		if err != nil {
			return nil, err
		}
		obj.Objects = append(obj.Objects, co)
		return obj, nil
	case ObjectTree:
		tree, err := ReadTree(obj)
		if err != nil {
			return nil, err
		}
		for _, v := range tree.Items {
			o, err := ReadObject(v.Sha)
			if err != nil {
				return nil, err
			}
			o.Path = v.Path
			if o.Typ != v.Typ {
				return nil, errors.New("types did not match somehow")
			}
			obj.Objects = append(obj.Objects, o)
		}
		return obj, nil
	case ObjectBlob:
		// lets not read the whole blob
		return nil, nil
	}
	return nil, errors.New("unhandled object type")

}

func ReadTree(obj *Object) (*Tree, error) {
	var err error
	var p []byte

	tree := &Tree{}
	r, err := obj.ReadCloser()
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	if err := readHeadBytes(r, obj); err != nil {
		return nil, err
	}
	//
	sha := make([]byte, 20)
	buf := bufio.NewReader(r)
	// there should be a null byte after file path, then 20 byte sha
	for {
		itm := &TreeItem{}
		p, err = buf.ReadBytes(0)

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		_, err = io.ReadFull(buf, sha)
		item := bytes.Fields(p)
		itm.Sha = []byte(hex.EncodeToString(sha))
		if string(item[0]) == "40000" {
			itm.Typ = ObjectTree
			if err != nil {
				return nil, err
			}
		} else {
			itm.Typ = ObjectBlob
		}
		itm.Path = string(item[1][:len(item[1])-1])

		if err == io.EOF {
			break
		}
		tree.Items = append(tree.Items, itm)
	}
	return tree, nil
}

func readHeadBytes(r io.ReadCloser, obj *Object) error {
	n, err := r.Read(make([]byte, obj.HeaderLength))
	if err != nil {
		return err
	}
	if n != obj.HeaderLength {
		return fmt.Errorf("read %d not %d", n, obj.HeaderLength)
	}
	return nil
}

func ReadCommit(sha []byte) (*Commit, error) {
	o, err := ReadObject(sha)
	if err != nil {
		return nil, err
	}
	return readCommit(o)
}

// The format for a commit object is simple:
// it specifies the top-level tree for the snapshot of the project at that point;
// the parent commits if any (the commit object described above does not have any parents);
// the author/committer information (which uses your user.name and user.email configuration settings and a timestamp);
// a blank line, and then the commit message.
func readCommit(obj *Object) (*Commit, error) {
	r, err := obj.ReadCloser()
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	if err := readHeadBytes(r, obj); err != nil {
		return nil, err
	}
	c := &Commit{Sha: obj.Sha}

	s := bufio.NewScanner(r)
	var parts [][]byte
	parts = lineParts(s)
	c.Tree = parts[1]
	for {
		parts = lineParts(s)
		if string(parts[0]) == "parent" {
			c.Parents = append(c.Parents, parts[1])
		} else {
			break
		}
	}
	if string(parts[0]) != "author" {
		return nil, fmt.Errorf("expected author got %s", string(parts[0]))
	}
	parts = lineParts(s)
	if string(parts[0]) != "committer" {
		return nil, fmt.Errorf("expected committer got %s", string(parts[0]))
	}
	s.Scan()
	if s.Text() != "" {
		return nil, fmt.Errorf("expected newline got %s", s.Text())
	}
	for {
		if !s.Scan() {
			break
		}
		c.Message = append(c.Message, s.Bytes()...)
	}

	return c, nil
}
func lineParts(s *bufio.Scanner) [][]byte {
	s.Scan()
	return bytes.Split(s.Bytes(), []byte(" "))
}
