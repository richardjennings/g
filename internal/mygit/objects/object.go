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
	fmt.Printf("commit: %s\n", string(c.Sha))
	fmt.Printf("tree: %s\n", string(c.Tree))
	for _, v := range c.Parents {
		fmt.Printf("parent: %s\n", string(v))
	}
	fmt.Printf("%s <%s> %s\n", c.Author, c.AuthorEmail, c.AuthoredTime.String())
	fmt.Printf("%s <%s> %s\n", c.Committer, c.CommitterEmail, c.CommittedTime.String())
	return fmt.Sprintf("message: \n%s\n", c.Message)
}
