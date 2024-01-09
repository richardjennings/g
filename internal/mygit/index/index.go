package index

import (
	"errors"
	"github.com/richardjennings/mygit/internal/mygit/fs"
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
func (idx *Index) Files() []*fs.File {
	var files []*fs.File
	for _, v := range idx.items {
		idx := &fs.File{Path: string(v.Name), Sha: v.Sha[:]}
		files = append(files, idx)
	}
	return files
}

// Add adds a fs.File to the Index Struct. A call to idx.Write is required
// to flush the changes to the filesystem.
func (idx *Index) Add(f *fs.File) error {
	// if delete, remove from Index
	if f.Status == fs.StatusDeleted {
		for i, v := range idx.items {
			if string(v.Name) == f.Path {
				idx.items = append(idx.items[0:i], idx.items[i+1:]...)
				idx.header.NumEntries--
				return nil
			}
		}
		return errors.New("somehow the file was not found in Index items to be removed")
	} else if f.Status == fs.StatusUntracked {
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
	} else if f.Status == fs.StatusModified {
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

func item(f *fs.File) (*indexItem, error) {
	if f.Sha == nil {
		return nil, errors.New("missing Sha from working directory file toIndexItem")
	}
	item := &indexItem{indexItemP: &indexItemP{}}
	switch runtime.GOOS {
	case "darwin":
	case "linux":
	default:
		return nil, errors.New("setItemOsSpecificStat not implemented, unsupported OS")
	}
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
	copy(item.Sha[:], f.Sha)
	nameLen := len(f.Path)
	if nameLen < 0xFFF {
		item.Flags = uint16(len(f.Path))
	} else {
		item.Flags = 0xFFF
	}
	item.Name = []byte(f.Path)

	return item, nil
}

func newIndex() *Index {
	return &Index{header: &indexHeader{
		Sig:        [4]byte{'D', 'I', 'R', 'C'},
		Version:    2,
		NumEntries: 0,
	}}
}
