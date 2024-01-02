package mygit

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type (
	index struct {
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
		CtimeS uint32
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

	// represent working directory files and index status
	wdFile struct {
		path   string
		finfo  os.FileInfo
		status indexStatusTyp
	}
)

const (
	indexStatusInvalid indexStatusTyp = iota
	indexStatusModified
	indexStatusAdded
	indexStatusDeleted
	indexStatusUnchanged
)

func (i *index) fileNames() []string {
	var files []string
	for _, v := range i.items {
		files = append(files, string(v.Name))
	}
	return files
}

// writeIndex writes an index struct to the Git Index
func (m *MyGit) writeIndex(index *index) error {
	path := filepath.Join(m.path, m.gitDirectory, DefaultIndexFile)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	// write header
	if err := binary.Write(f, binary.BigEndian, index.header); err != nil {
		return err
	}
	// write each item fixed size entry
	for _, item := range index.items {
		if err := binary.Write(f, binary.BigEndian, item.indexItemP); err != nil {
			return err
		}
		// write name
		if _, err := f.Write(item.Name); err != nil {
			return err
		}
		// write padding
		padding := make([]byte, 8-(62+len(item.Name))%8)
		if _, err := f.Write(padding); err != nil {
			return err
		}
	}
	// write sha hash of index
	if err := binary.Write(f, binary.BigEndian, &index.sig); err != nil {
		return err
	}
	return nil
}

// readIndex reads the Git Index into an index struct
// @todo better implemented as a reader / writer
func (m *MyGit) readIndex() (*index, error) {
	path := filepath.Join(m.path, m.gitDirectory, DefaultIndexFile)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	// populate indexHeader
	index := &index{header: &indexHeader{}}
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

// status returns a list of files in the working directory that are
// modified, added or deleted.
func (m *MyGit) indexStatus() ([]*wdFile, error) {
	index, err := m.readIndex()
	if err != nil {
		return nil, err
	}
	files, err := m.wdFiles()
	if err != nil {
		return nil, err
	}
	// create an index for wd files
	wdFileIdx := make(map[string]*wdFile)
	for _, v := range files {
		wdFileIdx[v.path] = v
	}

	// create an index for index files
	idxFileIdx := make(map[string]*indexItem)
	for _, v := range index.items {
		idxFileIdx[string(v.Name)] = v
	}

	// check for any files in wdFileIdx that are not in idxFileIdx,
	// these will have added to working directory status
	for v, f := range wdFileIdx {
		if _, ok := idxFileIdx[v]; !ok {
			f.status = indexStatusAdded
		}
	}

	// check for any files in idxFileIdx that are not in wdFileIdx,
	// these will have removed from working directory status
	for v, _ := range idxFileIdx {
		if _, ok := wdFileIdx[v]; !ok {
			// add a deleted file to wd files
			files = append(files, &wdFile{path: v, status: indexStatusDeleted})
		}
	}

	// now check file properties to detect no change for files without a status already
	// modification time,
	// size,
	// mode ?, //@todo not bother for now
	for _, v := range files {
		if v.status != indexStatusInvalid {
			continue // skip files with added or deleted status
		}
		i, ok := idxFileIdx[v.path]
		if !ok {
			// this really should not happen... right ?
			return nil, errors.New("file was meant to be in index map but was not somehow")
		}
		mt := time.Unix(int64(i.MTimeS), int64(i.MTimeN))
		if v.finfo.ModTime().Equal(mt) && v.finfo.Size() == int64(i.Size) {
			v.status = indexStatusUnchanged
		}
	}

	// the remaining files without a status might be modified,
	// recalculate the sha hash of the file to be sure ...

	return files, nil
}

func (m *MyGit) ReadWriteIndex() error {
	index, err := m.readIndex()
	if err != nil {
		return err
	}
	return m.writeIndex(index)
}

// LsFiles returns a list of files in the index
func (m *MyGit) LsFiles() ([]string, error) {
	index, err := m.readIndex()
	if err != nil {
		return nil, err
	}
	return index.fileNames(), nil
}

// Status currently displays the
func (m *MyGit) Status(o io.Writer) error {
	files, err := m.indexStatus()
	if err != nil {
		return err
	}
	var s string
	for _, v := range files {
		switch v.status {
		case indexStatusInvalid:
			s = "x"
		case indexStatusDeleted:
			s = "D"
		case indexStatusAdded:
			s = "??"
		case indexStatusUnchanged:
			continue
		}
		if _, err := fmt.Fprintf(o, "%s %s\n", s, v.path); err != nil {
			return err
		}
	}
	return nil
}
