package index

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/richardjennings/mygit/pkg/config"
	"github.com/richardjennings/mygit/pkg/gfs"
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
func (idx *Index) Files() []*gfs.File {
	var files []*gfs.File
	for _, v := range idx.items {
		s, _ := gfs.NewSha(v.Sha[:])
		idx := &gfs.File{Path: string(v.Name), Sha: s, Finfo: fromIndexItemP(v.indexItemP)}
		files = append(files, idx)
	}
	return files
}

func (idx *Index) File(path string) *gfs.File {
	for _, v := range idx.items {
		if string(v.Name) == path {
			s, _ := gfs.NewSha(v.Sha[:])
			return &gfs.File{Path: string(v.Name), Sha: s, Finfo: fromIndexItemP(v.indexItemP)}
		}
	}
	return nil
}

// Rm removes a gfs.File from the Index
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
func (idx *Index) Add(f *gfs.File) error {
	// if delete, remove from Index
	if f.WdStatus == gfs.WDDeletedInWorktree {
		for i, v := range idx.items {
			if string(v.Name) == f.Path {
				idx.items = append(idx.items[0:i], idx.items[i+1:]...)
				idx.header.NumEntries--
				return nil
			}
		}
		return errors.New("somehow the file was not found in Index items to be removed")
	} else if f.WdStatus == gfs.WDUntracked {
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
	} else if f.WdStatus == gfs.WDWorktreeChangedSinceIndex {
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
	path := config.IndexFilePath()
	f, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	// use a multi-writer to allow both writing the the file whilst incrementally generating
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

func item(f *gfs.File) (*indexItem, error) {
	if f.Sha == nil {
		return nil, errors.New("missing Sha from working directory file toIndexItem")
	}
	if f.Finfo == nil {
		info, err := os.Stat(filepath.Join(config.Path(), f.Path))
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
	case *gfs.Finfo:
		fi := f.Finfo.(*gfs.Finfo)
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

func fromIndexItemP(p *indexItemP) *gfs.Finfo {
	f := &gfs.Finfo{
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
