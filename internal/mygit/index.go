package mygit

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"syscall"
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

	// represent working directory files and index status
	wdFile struct {
		path   string
		finfo  os.FileInfo
		status indexStatusTyp
		sha    []byte
	}
	idxFile struct {
		path   string
		status indexStatusTyp
		sha    []byte
	}
)

const (
	indexStatusInvalid   indexStatusTyp = iota
	indexStatusModified                 // different in working directory than index
	indexStatusUntracked                // in working directory but not in index
	indexStatusAdded                    // in index but not in last commit
	indexStatusDeleted                  // in last commit but not in index
	indexStatusUnchanged
)

func (wdf *wdFile) toIndexItem() (*indexItem, error) {
	if wdf.sha == nil {
		return nil, errors.New("missing sha from working directory file toIndexItem")
	}
	item := &indexItem{indexItemP: &indexItemP{}}
	item.CTimeS = uint32(wdf.finfo.Sys().(*syscall.Stat_t).Ctimespec.Sec)
	item.CTimeN = uint32(wdf.finfo.Sys().(*syscall.Stat_t).Ctimespec.Nsec)
	item.MTimeS = uint32(wdf.finfo.Sys().(*syscall.Stat_t).Mtimespec.Sec)
	item.MTimeN = uint32(wdf.finfo.Sys().(*syscall.Stat_t).Mtimespec.Nsec)
	item.Dev = uint32(wdf.finfo.Sys().(*syscall.Stat_t).Dev)
	item.Ino = uint32(wdf.finfo.Sys().(*syscall.Stat_t).Ino)
	if wdf.finfo.IsDir() {
		item.Mode = uint32(040000)
	} else {
		item.Mode = uint32(0100644)
	}
	item.Uid = wdf.finfo.Sys().(*syscall.Stat_t).Uid
	item.Gid = wdf.finfo.Sys().(*syscall.Stat_t).Gid
	item.Size = uint32(wdf.finfo.Size())
	copy(item.Sha[:], wdf.sha)
	nameLen := len(wdf.path)
	if nameLen < 0xFFF {
		item.Flags = uint16(len(wdf.path))
	} else {
		item.Flags = 0xFFF
	}
	item.Name = []byte(wdf.path)

	return item, nil
}

func (idx *index) idxFiles() []*idxFile {
	var files []*idxFile
	for _, v := range idx.items {
		idx := &idxFile{path: string(v.Name), sha: v.Sha[:]}
		files = append(files, idx)
	}
	return files
}

func (idx *index) addWdFile(f *wdFile) error {
	// if delete, remove from index
	if f.status == indexStatusDeleted {
		for i, v := range idx.items {
			if string(v.Name) == f.path {
				idx.items = append(idx.items[0:i], idx.items[i+1:]...)
				idx.header.NumEntries--
				return nil
			}
		}
		return errors.New("somehow the file was not found in index items to be removed")
	} else if f.status == indexStatusUntracked {
		// just add and sort all of them for now
		item, err := f.toIndexItem()
		if err != nil {
			return err
		}
		idx.items = append(idx.items, item)
		idx.header.NumEntries++
		// and sort @todo more efficient
		sort.Slice(idx.items, func(i, j int) bool {
			if string(idx.items[i].Name) < string(idx.items[j].Name) {
				return true
			}
			return false
		})
	} else if f.status == indexStatusModified {
		// @todo add support for changing existing entries when working dir file is changed
		return errors.New("updating modified file in index not written yet")
	}

	return nil
}

func (m *MyGit) newIndex() *index {
	return &index{header: &indexHeader{
		Sig:        [4]byte{'D', 'I', 'R', 'C'},
		Version:    2,
		NumEntries: 0,
	}}
}

// writeIndex writes an index struct to the Git Index
func (m *MyGit) writeIndex(index *index) error {
	if index.header.NumEntries != uint32(len(index.items)) {
		return errors.New("index numEntries and length of items inconsistent")
	}
	path := filepath.Join(m.path, m.gitDirectory, DefaultIndexFile)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	// use a multi-writer to allow both writing the the file whilst incrementally generating
	// a sha hash of the content as it is written
	h := sha1.New()
	mw := io.MultiWriter(f, h)

	// write header
	if err := binary.Write(mw, binary.BigEndian, index.header); err != nil {
		return err
	}
	// write each item fixed size entry
	for _, item := range index.items {
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
	copy(index.sig[:], sha)
	// write sha hash of index
	if err := binary.Write(f, binary.BigEndian, &sha); err != nil {
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
		if errors.Is(err, os.ErrNotExist) {
			return m.newIndex(), nil
		}
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
func (m *MyGit) wdStatus() ([]*wdFile, error) {
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
			f.status = indexStatusUntracked
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
		// the remaining files without a status might be modified,
		// recalculate the sha hash of the file to be sure ...
		// @todo for now lets just assume they are modified

		h := sha1.New()
		f, err := os.Open(filepath.Join(m.path, v.path))
		if err != nil {
			return nil, err
		}
		header := []byte(fmt.Sprintf("blob %d%s", v.finfo.Size(), string(byte(0))))
		h.Write(header)
		_, err = io.Copy(h, f)
		_ = f.Close()
		if err != nil {
			return nil, err
		}
		if string(h.Sum(nil)) != string(idxFileIdx[v.path].Sha[:]) {
			v.status = indexStatusModified
		} else {
			v.status = indexStatusUnchanged
		}
	}

	return files, nil
}

// ReadWriteIndex @todo remove - used to test writing and reading are correct (enough)
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
	var files []string
	for _, v := range index.idxFiles() {
		files = append(files, v.path)
	}
	return files, nil
}

func (m *MyGit) Add(paths ...string) error {
	index, err := m.readIndex()
	if err != nil {
		return err
	}
	// get working directory files with index status
	wdFiles, err := m.wdStatus()
	var updates []*wdFile
	for _, p := range paths {
		if p == "." {
			// special case meaning add everything
			for _, v := range wdFiles {
				switch v.status {
				case indexStatusUntracked, indexStatusModified, indexStatusDeleted:
					updates = append(updates, v)
				}
			}
		} else {
			// @todo add support for paths other than just '.'
			return errors.New("only supports '.' currently ")
		}
	}
	for _, v := range updates {
		switch v.status {
		case indexStatusUntracked, indexStatusModified:
			// add the file to the object store
			obj, err := m.storeBlob(filepath.Join(m.path, v.path))
			if err != nil {
				return err
			}
			v.sha = obj.sha
		}
		if err := index.addWdFile(v); err != nil {
			return err
		}
	}
	// once all wdFiles are added to index struct, write it out
	return m.writeIndex(index)
}

// Status currently displays the
func (m *MyGit) Status(o io.Writer) error {
	files, err := m.wdStatus()
	if err != nil {
		return err
	}
	var s string
	for _, v := range files {
		switch v.status {
		case indexStatusInvalid:
			s = "x"
		case indexStatusModified:
			s = "M"
		case indexStatusDeleted:
			s = "D"
		case indexStatusUntracked:
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
