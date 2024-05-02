package objects

import (
	"fmt"
	"github.com/richardjennings/mygit/pkg/mygit/config"
	"github.com/richardjennings/mygit/pkg/mygit/gfs"
	"io"
	"path/filepath"
	"strings"
	"time"
)

type (
	Object struct {
		Path         string
		Typ          objectType
		Sha          []byte
		Objects      []*Object
		Length       int
		HeaderLength int
		ReadCloser   func() (io.ReadCloser, error)
		//mode    string
	}
	objectType int
	Commit     struct {
		Sha            []byte
		Tree           []byte
		Parents        [][]byte
		Author         string
		AuthorEmail    string
		AuthoredTime   time.Time
		Committer      string
		CommitterEmail string
		CommittedTime  time.Time
		Sig            []byte
		Message        []byte
	}
	Tree struct {
		Sha   []byte
		Typ   objectType
		Path  string
		Items []*TreeItem
	}
	TreeItem struct {
		Sha  []byte
		Typ  objectType
		Path string
	}
)

const (
	ObjectTypeInvalid objectType = iota
	ObjectTypeBlob
	ObjectTypeTree
	ObjectTypeCommit
)

func (c Commit) String() string {
	var o string
	o += fmt.Sprintf("commit: %s\n", string(c.Sha))
	o += fmt.Sprintf("tree: %s\n", string(c.Tree))
	for _, v := range c.Parents {
		o += fmt.Sprintf("parent: %s\n", string(v))
	}
	o += fmt.Sprintf("%s <%s> %s\n", c.Author, c.AuthorEmail, c.AuthoredTime.String())
	o += fmt.Sprintf("%s <%s> %s\n", c.Committer, c.CommitterEmail, c.CommittedTime.String())
	o += fmt.Sprintf("message: \n%s\n", c.Message)
	return o
}

// ObjectTree creates a Tree Object with child Objects representing the files and
// paths in the provided files.
func ObjectTree(files []*gfs.File) *Object {
	root := &Object{}
	var n *Object  // current node
	var pn *Object // previous node
	// mp holds a cache of file paths to objectTree nodes
	mp := make(map[string]*Object)
	for _, v := range files {
		parts := strings.Split(strings.TrimPrefix(v.Path, config.WorkingDirectory()), string(filepath.Separator))
		if len(parts) == 1 {
			root.Objects = append(root.Objects, &Object{Typ: ObjectTypeBlob, Path: v.Path, Sha: v.Sha.AsBytes()})
			continue // top level file
		}
		pn = root
		for i, p := range parts {
			if i == len(parts)-1 {
				pn.Objects = append(pn.Objects, &Object{Typ: ObjectTypeBlob, Path: v.Path, Sha: v.Sha.AsBytes()})
				continue // leaf
			}
			// key for cached nodes
			key := strings.Join(parts[0:i+1], string(filepath.Separator))
			cached, ok := mp[key]
			if ok {
				n = cached
			} else {
				n = &Object{Typ: ObjectTypeTree, Path: p}
				pn.Objects = append(pn.Objects, n)
				mp[key] = n
			}
			pn = n
		}
	}

	return root
}
