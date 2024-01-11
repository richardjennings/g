package objects

import (
	"fmt"
	"io"
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
	ObjectInvalid objectType = iota
	ObjectBlob
	ObjectTree
	ObjectCommit
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
