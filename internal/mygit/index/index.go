package index

import (
	"errors"
	"os"
	"sort"
)

const (
	StatusInvalid   indexStatusTyp = iota
	StatusModified                 // different in working directory than Index
	StatusUntracked                // in working directory but not in Index
	StatusAdded                    // in Index but not in last commit
	StatusDeleted                  // in last commit but not in Index
	StatusUnchanged
)

type (
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
	indexStatusTyp uint8

	// represent working directory files and Index Status
	File struct {
		Path   string
		Finfo  os.FileInfo
		Status indexStatusTyp
		Sha    []byte
	}
	IdxFile struct {
		Path string
		//Status indexStatusTyp
		Sha []byte
	}
)

func (idx *Index) IdxFiles() []*IdxFile {
	var files []*IdxFile
	for _, v := range idx.items {
		idx := &IdxFile{Path: string(v.Name), Sha: v.Sha[:]}
		files = append(files, idx)
	}
	return files
}

func (idx *Index) AddWdFile(f *File) error {
	// if delete, remove from Index
	if f.Status == StatusDeleted {
		for i, v := range idx.items {
			if string(v.Name) == f.Path {
				idx.items = append(idx.items[0:i], idx.items[i+1:]...)
				idx.header.NumEntries--
				return nil
			}
		}
		return errors.New("somehow the file was not found in Index items to be removed")
	} else if f.Status == StatusUntracked {
		// just add and sort all of them for now
		item, err := f.toIndexItem()
		if err != nil {
			return err
		}
		idx.items = append(idx.items, item)
		idx.header.NumEntries++
		// and sort @todo more efficient
		sort.Slice(idx.items, func(i, j int) bool {
			return string(idx.items[i].Name) < string(idx.items[j].Name)
		})
	} else if f.Status == StatusModified {
		// @todo add support for changing existing entries when working dir file is changed
		return errors.New("updating modified file in Index not written yet")
	}

	return nil
}

func newIndex() *Index {
	return &Index{header: &indexHeader{
		Sig:        [4]byte{'D', 'I', 'R', 'C'},
		Version:    2,
		NumEntries: 0,
	}}
}
