package index

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/richardjennings/mygit/internal/mygit/config"
	"github.com/richardjennings/mygit/internal/mygit/objects"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func (idx *Index) CommitStatus(sha []byte) ([]*File, error) {
	var files []*File
	f := idx.IdxFiles()
	if sha == nil {
		for _, v := range f {
			files = append(files, &File{Path: v.Path, Sha: v.Sha, Status: StatusAdded})
		}
		return files, nil
	}
	obj, err := objects.ReadObject(sha)
	if err != nil {
		return nil, err
	}
	objFiles := obj.FlattenTree()
	objMap := make(map[string]*objects.ObjectFile)
	itemMap := make(map[string]*IdxFile)
	for _, o := range objFiles {
		objMap[o.Path] = o
	}
	for _, o := range f {
		itemMap[o.Path] = o
	}
	for _, v := range f {
		item, ok := objMap[v.Path]
		if !ok {
			files = append(files, &File{Path: v.Path, Sha: v.Sha, Status: StatusAdded})
		} else if string(item.Sha) != hex.EncodeToString(v.Sha) {
			files = append(files, &File{Path: v.Path, Sha: v.Sha, Status: StatusModified})
		}
	}
	for _, v := range objFiles {
		if _, ok := itemMap[v.Path]; !ok {
			files = append(files, &File{Path: v.Path, Sha: v.Sha, Status: StatusDeleted})
		}
	}

	return files, nil
}

// Status returns a list of files in the working directory that are
// modified, added or deleted.
func WdStatus() ([]*File, error) {
	idx, err := ReadIndex()
	if err != nil {
		return nil, err
	}
	files, err := WdFiles()
	if err != nil {
		return nil, err
	}
	// create an Index for wd files
	wdFileIdx := make(map[string]*File)
	for _, v := range files {
		wdFileIdx[v.Path] = v
	}

	// create an Index for Index files
	idxFileIdx := make(map[string]*indexItem)
	for _, v := range idx.items {
		idxFileIdx[string(v.Name)] = v
	}

	// check for any files in wdFileIdx that are not in idxFileIdx,
	// these will have added to working directory Status
	for v, f := range wdFileIdx {
		if _, ok := idxFileIdx[v]; !ok {
			f.Status = StatusUntracked
		}
	}

	// check for any files in idxFileIdx that are not in wdFileIdx,
	// these will have removed from working directory Status
	for v := range idxFileIdx {
		if _, ok := wdFileIdx[v]; !ok {
			// add a deleted file to wd files
			files = append(files, &File{Path: v, Status: StatusDeleted})
		}
	}

	// now check file properties to detect no change for files without a Status already
	// modification time,
	// size,
	// mode ?, //@todo not bother for now
	for _, v := range files {
		if v.Status != StatusInvalid {
			continue // skip files with added or deleted Status
		}
		i, ok := idxFileIdx[v.Path]
		if !ok {
			// this really should not happen... right ?
			return nil, errors.New("file was meant to be in Index map but was not somehow")
		}
		mt := time.Unix(int64(i.MTimeS), int64(i.MTimeN))
		if v.Finfo.ModTime().Equal(mt) && v.Finfo.Size() == int64(i.Size) {
			v.Status = StatusUnchanged
		}
		// the remaining files without a Status might be modified,
		// recalculate the Sha hash of the file to be sure ...
		// @todo for now lets just assume they are modified

		h := sha1.New()
		f, err := os.Open(filepath.Join(config.Path(), v.Path))
		if err != nil {
			return nil, err
		}
		header := []byte(fmt.Sprintf("blob %d%s", v.Finfo.Size(), string(byte(0))))
		h.Write(header)
		_, err = io.Copy(h, f)
		_ = f.Close()
		if err != nil {
			return nil, err
		}
		if string(h.Sum(nil)) != string(idxFileIdx[v.Path].Sha[:]) {
			v.Status = StatusModified
		} else {
			v.Status = StatusUnchanged
		}
	}

	return files, nil
}

// list working directory files that are not ignored
func WdFiles() ([]*File, error) {
	var wdFiles []*File
	if err := filepath.Walk(config.Path(), func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		// do not add ignored files
		if !isIgnored(path) {
			wdFiles = append(wdFiles, &File{
				Path:  strings.TrimPrefix(path, config.WorkingDirectory()),
				Finfo: info,
			})
		}
		return nil
	}); err != nil {
		return wdFiles, err
	}
	return wdFiles, nil
}

func isIgnored(path string) bool {
	// remove absolute portion of Path
	path = strings.TrimPrefix(path, config.Path())
	path = strings.TrimPrefix(path, string(filepath.Separator))
	if path == "" {
		return true
	}
	// @todo fix literal string prefix matching and iteration
	for _, v := range config.Config.GitIgnore {
		if strings.HasPrefix(path, v) {
			return true
		}
	}
	// @todo remove special git case
	if strings.HasPrefix(path, config.DefaultGitDirectory) {
		return true
	}
	if strings.HasPrefix(path, config.Config.GitDirectory) {
		return true
	}
	return false
}

func (wdf *File) toIndexItem() (*indexItem, error) {
	if wdf.Sha == nil {
		return nil, errors.New("missing Sha from working directory file toIndexItem")
	}
	item := &indexItem{indexItemP: &indexItemP{}}
	switch runtime.GOOS {
	case "darwin":
	case "linux":
	default:
		return nil, errors.New("setItemOsSpecificStat not implemented, unsupported OS")
	}
	setItemOsSpecificStat(wdf.Finfo, item)
	item.Dev = uint32(wdf.Finfo.Sys().(*syscall.Stat_t).Dev)
	item.Ino = uint32(wdf.Finfo.Sys().(*syscall.Stat_t).Ino)
	if wdf.Finfo.IsDir() {
		item.Mode = uint32(040000)
	} else {
		item.Mode = uint32(0100644)
	}
	item.Uid = wdf.Finfo.Sys().(*syscall.Stat_t).Uid
	item.Gid = wdf.Finfo.Sys().(*syscall.Stat_t).Gid
	item.Size = uint32(wdf.Finfo.Size())
	copy(item.Sha[:], wdf.Sha)
	nameLen := len(wdf.Path)
	if nameLen < 0xFFF {
		item.Flags = uint16(len(wdf.Path))
	} else {
		item.Flags = 0xFFF
	}
	item.Name = []byte(wdf.Path)

	return item, nil
}
