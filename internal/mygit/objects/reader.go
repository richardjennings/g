package objects

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"errors"
	"github.com/richardjennings/mygit/internal/mygit/config"
	"io"
	"os"
	"path/filepath"
)

func (o *Object) FlattenTree() []*ObjectFile {
	var objFiles []*ObjectFile
	if o.Typ == ObjectBlob {
		return []*ObjectFile{{Path: o.Path, Sha: o.Sha}}
	}
	for _, v := range o.Objects {
		objs := v.FlattenTree()
		for i, _ := range objs {
			objs[i].Path = filepath.Join(o.Path, objs[i].Path)
		}
		objFiles = append(objFiles, objs...)
	}

	return objFiles
}

func ReadCommitTree(sha []byte) (*Object, error) {
	obj, err := ReadObject(sha)
	if err != nil {
		return nil, err
	}
	if obj.Typ != ObjectCommit {
		return nil, errors.New("expected commit")
	}
	return obj, nil
	//return ReadObject(sha)
}

// ReadObject reads an object from the object store
// header
//
//	type lenNIL
func ReadObject(sha []byte) (*Object, error) {
	path := filepath.Join(config.ObjectPath(), string(sha[0:2]), string(sha[2:]))
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	z, err := zlib.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer func() { _ = z.Close() }()
	buf := bufio.NewReader(z)

	// read parts by null byte
	p, err := buf.ReadBytes(0)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	header := bytes.Fields(p)
	o := &Object{Sha: sha}
	switch string(header[0]) {
	case "commit":
		o.Typ = ObjectCommit
		content, err := buf.ReadBytes(0)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		tree := bytes.Fields(content)
		if len(tree) < 2 {
			return nil, errors.New("expected at least 2 parts")
		}

		_ = z.Close()

		co, err := ReadObject(tree[1])
		if err != nil {
			return nil, err
		}
		o.Objects = append(o.Objects, co)
		return o, nil
	case "tree":
		o.Typ = ObjectTree
		sha := make([]byte, 20)

		// there should be a null byte after file path, then 20 byte sha
		for {
			p, err = buf.ReadBytes(0)

			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return nil, err
			}
			_, err = io.ReadFull(buf, sha)
			// buf.ReadBytes just keeps returning the same data with no error ?

			item := bytes.Fields(p)
			co := &Object{}
			co.Sha = []byte(hex.EncodeToString(sha))
			if string(item[0]) == "40000" {
				co.Typ = ObjectTree
				co, err = ReadObject(co.Sha)
				if err != nil {
					return nil, err
				}
			} else {
				co.Typ = ObjectBlob
			}
			co.Path = string(item[1][:len(item[1])-1])

			o.Objects = append(o.Objects, co)

			if err == io.EOF {
				break
			}

			//if co.Typ == ObjectBlob {
			//	continue
			//}

			//ooo, err := ReadObject(oo.Sha)
			//if err != nil {
			//	return nil, err
			//}
			//if ooo != nil {
			//	oo.Objects = append(oo.Objects, ooo)
			//}
		}
		return o, nil
	case "blob":
		// lets not read the whole blob
		return nil, nil

	default:
		return nil, errors.New("unhandled object type")
	}
}
