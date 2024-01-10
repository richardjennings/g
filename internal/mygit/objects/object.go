package objects

import (
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
		Sha           []byte
		Tree          []byte
		Parents       [][]byte
		Author        string
		AuthoredTime  time.Time
		Committer     string
		CommittedTime time.Time
		Message       string
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
