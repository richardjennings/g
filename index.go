package g

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"syscall"
)

type (
	// Index represents the Git Index
	Index struct {
		header *indexHeader
		items  []*indexItem
		sig    [20]byte
	}
	indexHeader struct {
		Sig        [4]byte
		Version    uint32
		NumEntries uint32
	}
	indexItem struct {
		*indexItemP
		Name []byte
	}
	indexItemP struct {
		CTimeS uint32
		CTimeN uint32
		MTimeS uint32
		MTimeN uint32
		Dev    uint32
		Ino    uint32
		Mode   uint32
		Uid    uint32
		Gid    uint32
		Size   uint32
		Sha    [20]byte
		Flags  uint16 // length of filename
	}
)

// Files lists the files in the index
func (idx *Index) Files() []*File {
	var files []*File
	for _, v := range idx.items {
		s, _ := NewSha(v.Sha[:])
		idx := &File{Path: string(v.Name), Sha: s, Finfo: fromIndexItemP(v.indexItemP)}
		files = append(files, idx)
	}
	return files
}

func (idx *Index) File(path string) *File {
	for _, v := range idx.items {
		if string(v.Name) == path {
			s, _ := NewSha(v.Sha[:])
			return &File{Path: string(v.Name), Sha: s, Finfo: fromIndexItemP(v.indexItemP)}
		}
	}
	return nil
}

// Rm removes a File from the Index
// A call to idx.Write is required to persist the change.
func (idx *Index) Rm(path string) error {
	for i, v := range idx.items {
		if string(v.Name) == path {
			idx.items = append(idx.items[:i], idx.items[i+1:]...)
			idx.header.NumEntries--
			return nil
		}
	}
	return fmt.Errorf("error: pathspec '%s' did not match any file(s) known to git", path)
}

// Add adds a fs.File to the Index Struct. A call to idx.Write is required
// to flush the changes to the filesystem.
func (idx *Index) Add(f *File) error {
	// if delete, remove from Index
	if f.WdStatus == WDDeletedInWorktree {
		for i, v := range idx.items {
			if string(v.Name) == f.Path {
				idx.items = append(idx.items[0:i], idx.items[i+1:]...)
				idx.header.NumEntries--
				return nil
			}
		}
		return errors.New("somehow the file was not found in Index items to be removed")
	} else if f.WdStatus == WDUntracked {
		// just add and sort all of them for now
		item, err := item(f)
		if err != nil {
			return err
		}
		idx.items = append(idx.items, item)
		idx.header.NumEntries++
		// and sort @todo more efficient
		sort.Slice(idx.items, func(i, j int) bool {
			return string(idx.items[i].Name) < string(idx.items[j].Name)
		})
	} else if f.WdStatus == WDWorktreeChangedSinceIndex {
		for i, v := range idx.items {
			if string(v.Name) == f.Path {
				item, err := item(f)
				if err != nil {
					return err
				}
				idx.items[i] = item
			}
		}
	}

	return nil
}

// Write writes an Index struct to the Git Index
func (idx *Index) Write() error {
	if idx.header.NumEntries != uint32(len(idx.items)) {
		return errors.New("index numEntries and length of items inconsistent")
	}
	path := IndexFilePath()
	f, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	// use a multi-writer to allow both writing the file whilst incrementally generating
	// a Sha hash of the content as it is written
	h := sha1.New()
	mw := io.MultiWriter(f, h)

	// write header
	if err := binary.Write(mw, binary.BigEndian, idx.header); err != nil {
		return err
	}
	// write each item fixed size entry
	for _, item := range idx.items {
		if err := binary.Write(mw, binary.BigEndian, item.indexItemP); err != nil {
			return err
		}
		// write name
		if _, err := mw.Write(item.Name); err != nil {
			return err
		}
		// write padding
		padding := make([]byte, 8-(62+len(item.Name))%8)
		if _, err := mw.Write(padding); err != nil {
			return err
		}
	}
	// use the generated hash
	sha := h.Sum(nil)
	copy(idx.sig[:], sha)
	// write Sha hash of Index
	if err := binary.Write(f, binary.BigEndian, &sha); err != nil {
		return err
	}

	return f.Close()
}

func item(f *File) (*indexItem, error) {
	if f.Finfo == nil {
		info, err := os.Stat(filepath.Join(Path(), f.Path))
		if err != nil {
			return nil, err
		}
		f.Finfo = info
	}
	item := &indexItem{indexItemP: &indexItemP{}}
	switch runtime.GOOS {
	case "darwin":
	case "linux":
	default:
		return nil, errors.New("setItemOsSpecificStat not implemented, unsupported OS")
	}
	switch f.Finfo.(type) {
	case *Finfo:
		fi := f.Finfo.(*Finfo)
		item.CTimeS = fi.CTimeS
		item.CTimeN = fi.CTimeN
		item.MTimeS = fi.MTimeS
		item.MTimeN = fi.MTimeN
		item.Dev = fi.Dev
		item.Ino = fi.Ino
		item.Mode = fi.MMode
		item.Uid = fi.Uid
	default:
		setItemOsSpecificStat(f.Finfo, item)
		item.Dev = uint32(f.Finfo.Sys().(*syscall.Stat_t).Dev)
		item.Ino = uint32(f.Finfo.Sys().(*syscall.Stat_t).Ino)
		if f.Finfo.IsDir() {
			item.Mode = uint32(040000)
		} else {
			item.Mode = uint32(0100644)
		}
		item.Uid = f.Finfo.Sys().(*syscall.Stat_t).Uid
		item.Gid = f.Finfo.Sys().(*syscall.Stat_t).Gid
		item.Size = uint32(f.Finfo.Size())
	}
	item.Sha = f.Sha.AsArray()
	nameLen := len(f.Path)
	if nameLen < 0xFFF {
		item.Flags = uint16(len(f.Path))
	} else {
		item.Flags = 0xFFF
	}
	item.Name = []byte(f.Path)

	return item, nil
}

func NewIndex() *Index {
	return &Index{header: &indexHeader{
		Sig:        [4]byte{'D', 'I', 'R', 'C'},
		Version:    2,
		NumEntries: 0,
	}}
}

func fromIndexItemP(p *indexItemP) *Finfo {
	f := &Finfo{
		CTimeS: p.CTimeS,
		CTimeN: p.CTimeN,
		MTimeS: p.MTimeS,
		MTimeN: p.MTimeN,
		Dev:    p.Dev,
		Ino:    p.Ino,
		MMode:  p.Mode,
		Uid:    p.Uid,
		Gid:    p.Gid,
		SSize:  p.Size,
		Sha:    p.Sha,
	}
	return f
}

// Status returns a FileSet containing all files from commit, index and working directory
// with the corresponding status.
func Status(idx *Index, commitSha []byte) (*FileSet, error) {
	var commitFiles []*File
	var err error
	if commitSha != nil {
		commitFiles, err = CommittedFiles(commitSha)
		if err != nil {
			return nil, err
		}
	}
	files := NewFileSet(commitFiles)
	files.MergeFromIndex(NewFileSet(idx.Files()))
	workingDirectoryFiles, err := Ls(Path())
	if err != nil {
		return nil, err
	}
	files.MergeFromWD(NewFileSet(workingDirectoryFiles))
	return files, nil
}

// FsStatus returns a FileSet containing all files from the index and working directory
// with the corresponding status.
func FsStatus(path string) (*FileSet, error) {
	idx, err := ReadIndex()
	if err != nil {
		return nil, err
	}
	idxFiles := idx.Files()
	idxSet := NewFileSet(idxFiles)
	files, err := Ls(path)
	if err != nil {
		return nil, err
	}
	idxSet.MergeFromWD(NewFileSet(files))
	return idxSet, nil
}

// ReadIndex reads the Git Index into an Index struct
func ReadIndex() (*Index, error) {
	path := IndexFilePath()
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewIndex(), nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()
	// populate indexHeader
	index := &Index{header: &indexHeader{}}
	if err := binary.Read(f, binary.BigEndian, index.header); err != nil {
		return nil, err
	}
	// read num items from header
	for i := 0; i < int(index.header.NumEntries); i++ {
		itemP := &indexItemP{}
		if err := binary.Read(f, binary.BigEndian, itemP); err != nil {
			return nil, err
		}
		// mask 4 bits out of 12bits of item flags to get filename length
		l := itemP.Flags & 0xFFF // 12 1s
		item := indexItem{indexItemP: itemP}
		// read l bytes into Name
		item.Name = make([]byte, l)
		if err := binary.Read(f, binary.BigEndian, &item.Name); err != nil {
			return nil, err
		}
		index.items = append(index.items, &item)
		// now read some bytes to make the total read for the item a multiple of 8
		padding := make([]byte, 8-(62+l)%8)
		if err := binary.Read(f, binary.BigEndian, &padding); err != nil {
			return nil, err
		}
	}
	if err := binary.Read(f, binary.BigEndian, &index.sig); err != nil {
		return nil, err
	}

	return index, nil
}
