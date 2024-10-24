package g

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
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
		Parents        []Sha
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
		o += fmt.Sprintf("parent: %s\n", v.AsHexString())
	}
	o += fmt.Sprintf("%s <%s> %s\n", c.Author, c.AuthorEmail, c.AuthoredTime.String())
	o += fmt.Sprintf("%s <%s> %s\n", c.Committer, c.CommitterEmail, c.CommittedTime.String())
	o += fmt.Sprintf("message: \n%s\n", c.Message)
	return o
}

// ObjectTree creates a Tree Object with child Objects representing the files and
// paths in the provided files.
func ObjectTree(files []*File) *Object {
	root := &Object{}
	var n *Object  // current node
	var pn *Object // previous node
	// mp holds a cache of file paths to objectTree nodes
	mp := make(map[string]*Object)
	for _, v := range files {
		parts := strings.Split(strings.TrimPrefix(v.Path, WorkingDirectory()), string(filepath.Separator))
		if len(parts) == 1 {
			root.Objects = append(root.Objects, &Object{Typ: ObjectTypeBlob, Path: v.Path, Sha: v.Sha.AsByteSlice()})
			continue // top level file
		}
		pn = root
		for i, p := range parts {
			if i == len(parts)-1 {
				pn.Objects = append(pn.Objects, &Object{Typ: ObjectTypeBlob, Path: v.Path, Sha: v.Sha.AsByteSlice()})
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

// FlattenTree turns a TreeObject structure into a flat list of file paths
func (o *Object) FlattenTree() []*File {
	var objFiles []*File
	if o.Typ == ObjectTypeBlob {
		s, _ := NewSha(o.Sha)
		f := []*File{{Path: o.Path, Sha: s}}
		return f
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

func ReadObject(sha Sha) (*Object, error) {
	var err error
	var o *Object

	// check if a loose file or in a packfile
	if _, err := os.Stat(filepath.Join(ObjectPath(), string(sha.AsHexBytes()[0:2]), string(sha.AsHexBytes()[2:]))); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return lookupInPackfiles(sha)
		} else {
			return nil, err
		}
	}

	o = &Object{Sha: sha.AsHexBytes()}
	o.ReadCloser = ObjectReadCloser(sha.AsHexBytes())
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
		o.Typ = ObjectTypeCommit
	case "tree":
		o.Typ = ObjectTypeTree
	case "blob":
		o.Typ = ObjectTypeBlob
	default:
		return nil, fmt.Errorf("unknown %s", string(header[0]))
	}
	o.Length, err = strconv.Atoi(string(header[1][:len(header[1])-1]))

	return o, err
}

func ObjectReadCloser(sha []byte) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) {
		path := filepath.Join(ObjectPath(), string(sha[0:2]), string(sha[2:]))
		f, err := os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			return nil, err
		}
		defer func() { _ = f.Close() }()
		return zlib.NewReader(f)
	}
}

// ReadObjectTree reads an object from the object store
func ReadObjectTree(sha Sha) (*Object, error) {
	obj, err := ReadObject(sha)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, fmt.Errorf("object %s not found", sha.AsHexString())
	}
	switch obj.Typ {
	case ObjectTypeCommit:
		commit, err := readCommit(obj)
		if err != nil {
			return obj, err
		}
		sha, err := NewSha(commit.Tree)
		if err != nil {
			return obj, err
		}
		co, err := ReadObjectTree(sha)
		if err != nil {
			return nil, err
		}
		obj.Objects = append(obj.Objects, co)
		return obj, nil
	case ObjectTypeTree:
		tree, err := ReadTree(obj)
		if err != nil {
			return nil, err
		}
		for _, v := range tree.Items {
			sha, err := NewSha(v.Sha)
			if err != nil {
				return obj, err
			}
			o, err := ReadObjectTree(sha)
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
	case ObjectTypeBlob:
		// lets not read the whole blob
		return obj, nil
	default:
		return nil, errors.New("unhandled object type")
	}

}

func ReadTree(obj *Object) (*Tree, error) {
	var err error
	var p []byte

	tree := &Tree{
		Sha: obj.Sha,
	}
	r, err := obj.ReadCloser()
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	if err := ReadHeadBytes(r, obj); err != nil {
		// Tree objects can be totally empty ...
		if errors.Is(err, io.EOF) {
			return tree, nil
		}
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
			itm.Typ = ObjectTypeTree
			if err != nil {
				return nil, err
			}
		} else {
			itm.Typ = ObjectTypeBlob
		}
		itm.Path = string(item[1][:len(item[1])-1])

		if err == io.EOF {
			break
		}
		tree.Items = append(tree.Items, itm)
	}
	return tree, nil
}

func ReadHeadBytes(r io.ReadCloser, obj *Object) error {
	n, err := r.Read(make([]byte, obj.HeaderLength))
	if err != nil {
		return err
	}
	if n != obj.HeaderLength {
		return fmt.Errorf("read %d not %d", n, obj.HeaderLength)
	}
	return nil
}

func ReadCommit(sha Sha) (*Commit, error) {
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
	if err := ReadHeadBytes(r, obj); err != nil {
		return nil, err
	}
	c := &Commit{Sha: obj.Sha}

	s := bufio.NewScanner(r)
	gpgsig := false

	for {
		if !s.Scan() {
			break
		}
		l := s.Bytes()
		p := bytes.SplitN(l, []byte(" "), 2)
		t := string(p[0])
		if c.Tree == nil {
			if t != "tree" {
				return nil, fmt.Errorf("expected tree got %s", t)
			}
			c.Tree = p[1]
			continue
		}
		// can be parent
		if t == "parent" {
			sha, err := NewSha(p[1])
			if err != nil {
				return nil, err
			}
			c.Parents = append(c.Parents, sha)
			continue
		}
		if c.Author == "" {
			// should be author
			if t != "author" {
				return nil, fmt.Errorf("expected author got %s", t)
			}
			// decode author line
			if err := readAuthor(p[1], c); err != nil {
				return nil, err
			}
			continue
		}
		if c.Committer == "" {
			// should be committer
			if t != "committer" {
				return nil, fmt.Errorf("expected committer got %s", t)
			}
			// decode committer line
			if err := readCommitter(p[1], c); err != nil {
				return nil, err
			}
			continue
		}
		// can be GPG Signature
		if t == "gpgsig" {
			gpgsig = true
			continue
		}
		if gpgsig {
			if len(p[1]) == 0 {
				continue
			}
			// @todo build up signature lines
			if string(p[1]) != "-----END PGP SIGNATURE-----" {
				c.Sig = append(c.Sig, l...)
				c.Sig = append(c.Sig, []byte("\n")...)
				continue
			}
			gpgsig = false
			continue
		}
		if len(c.Message) == 0 && len(l) < 2 {
			continue
		}
		// now we have the message body hopefully
		c.Message = append(c.Message, l...)
		c.Message = append(c.Message, []byte("\n")...)
	}

	return c, nil
}

func readAuthor(b []byte, c *Commit) error {
	s := bytes.Index(b, []byte("<"))
	e := bytes.Index(b, []byte(">"))
	c.Author = string(b[0 : s-1])
	c.AuthorEmail = string(b[s+1 : e])
	ut, err := strconv.ParseInt(string(b[e+2:e+2+10]), 10, 64)
	if err != nil {
		return err
	}
	// @todo timezone part
	c.AuthoredTime = time.Unix(ut, 0)
	return nil
}

func readCommitter(b []byte, c *Commit) error {
	s := bytes.Index(b, []byte("<"))
	e := bytes.Index(b, []byte(">"))
	c.Committer = string(b[0 : s-1])
	c.CommitterEmail = string(b[s+1 : e])
	ut, err := strconv.ParseInt(string(b[e+2:e+2+10]), 10, 64)
	if err != nil {
		return err
	}
	// @todo timezone part
	c.CommittedTime = time.Unix(ut, 0)
	return nil
}

func CommittedFiles(sha Sha) ([]*File, error) {
	obj, err := ReadObjectTree(sha)
	if err != nil {
		return nil, err
	}
	return obj.FlattenTree(), nil
}

// WriteTree writes an Object Tree to the object store.
func (o *Object) WriteTree() ([]byte, error) {
	// resolve child tree Objects
	for i, v := range o.Objects {
		if v.Typ == ObjectTypeTree {
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
		if fo.Typ == ObjectTypeTree {
			mode = "40000"
		} else {
			mode = "100644"
		}
		// @todo replace base..
		content = append(content, []byte(fmt.Sprintf("%s %s%s%s", mode, filepath.Base(fo.Path), string(byte(0)), fo.Sha))...)
	}
	header := []byte(fmt.Sprintf("tree %d%s", len(content), string(byte(0))))
	return WriteObject(header, content, "", ObjectPath())
}

// WriteObject writes an object to the object store
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

// WriteBlob writes a file to the object store as a blob and returns
// a Blob Object representation.
func WriteBlob(path string) (*Object, error) {
	path = filepath.Join(Path(), path)
	finfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	header := []byte(fmt.Sprintf("blob %d%s", finfo.Size(), string(byte(0))))
	sha, err := WriteObject(header, nil, path, ObjectPath())
	return &Object{Sha: sha, Path: path}, err
}

func WriteCommit(c *Commit) ([]byte, error) {
	var parentCommits string
	for _, v := range c.Parents {
		parentCommits += fmt.Sprintf("parent %s\n", v)
	}
	content := []byte(fmt.Sprintf(
		"tree %s\n%sauthor %s %d +0000\ncommitter %s %d +0000\n\n%s",
		hex.EncodeToString(c.Tree),
		parentCommits,
		c.Author,
		c.AuthoredTime.Unix(),
		c.Committer,
		c.CommittedTime.Unix(),
		c.Message,
	))
	header := []byte(fmt.Sprintf("commit %d%s", len(content), string(byte(0))))
	b, err := WriteObject(header, content, "", ObjectPath())
	if err != nil {
		return nil, err
	}
	branch, err := CurrentBranch()
	if err != nil {
		return nil, err
	}
	sha, err := NewSha(b)
	if err != nil {
		return nil, err
	}
	return sha.AsHexBytes(), UpdateBranchHead(branch, sha)
}
