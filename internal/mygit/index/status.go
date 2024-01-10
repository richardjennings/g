package index

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/richardjennings/mygit/internal/mygit/config"
	"github.com/richardjennings/mygit/internal/mygit/fs"
	"github.com/richardjennings/mygit/internal/mygit/objects"
	"io"
	"os"
	"path/filepath"
	"time"
)

// CommitStatus returns a slice of files with status flags indicating if a file in a
// commit is absent in the index, modified in the index, or added in the index.
func (idx *Index) CommitStatus(sha []byte) ([]*fs.File, error) {
	var files []*fs.File
	f := idx.Files()
	if sha == nil {
		for _, v := range f {
			files = append(files, &fs.File{Path: v.Path, Sha: v.Sha, Status: fs.StatusAdded})
		}
		return files, nil
	}
	obj, err := objects.ReadObjectTree(sha)
	if err != nil {
		return nil, err
	}
	objFiles := obj.FlattenTree()
	objMap := make(map[string]*fs.File)
	itemMap := make(map[string]*fs.File)
	for _, o := range objFiles {
		objMap[o.Path] = o
	}
	for _, o := range f {
		itemMap[o.Path] = o
	}
	for _, v := range f {
		item, ok := objMap[v.Path]
		if !ok {
			files = append(files, &fs.File{Path: v.Path, Sha: v.Sha, Status: fs.StatusAdded})
		} else if string(item.Sha) != hex.EncodeToString(v.Sha) {
			files = append(files, &fs.File{Path: v.Path, Sha: v.Sha, Status: fs.StatusModified})
		}
	}
	for _, v := range objFiles {
		if _, ok := itemMap[v.Path]; !ok {
			files = append(files, &fs.File{Path: v.Path, Sha: v.Sha, Status: fs.StatusDeleted})
		}
	}

	return files, nil
}

// FsStatus retrieves a recursive list of files at path and generates a status flag
// for each file representing if the file is absent from the index, is missing from
// the index, or has been modified.
func FsStatus(path string) ([]*fs.File, error) {
	idx, err := ReadIndex()
	if err != nil {
		return nil, err
	}
	files, err := fs.Ls(path)
	if err != nil {
		return nil, err
	}
	// create an Index for wd files
	wdFileIdx := make(map[string]*fs.File)
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
			f.Status = fs.StatusUntracked
		}
	}

	// check for any files in idxFileIdx that are not in wdFileIdx,
	// these will have removed from working directory Status
	for v := range idxFileIdx {
		if _, ok := wdFileIdx[v]; !ok {
			// add a deleted file to wd files
			files = append(files, &fs.File{Path: v, Status: fs.StatusDeleted})
		}
	}

	// now check file properties to detect no change for files without a Status already
	// modification time,
	// size,
	// mode ?, //@todo not bother for now
	for _, v := range files {
		if v.Status != fs.StatusInvalid {
			continue // skip files with added or deleted Status
		}
		i, ok := idxFileIdx[v.Path]
		if !ok {
			// this really should not happen... right ?
			return nil, errors.New("file was meant to be in Index map but was not somehow")
		}
		mt := time.Unix(int64(i.MTimeS), int64(i.MTimeN))
		if v.Finfo.ModTime().Equal(mt) && v.Finfo.Size() == int64(i.Size) {
			v.Status = fs.StatusUnchanged
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
			v.Status = fs.StatusModified
		} else {
			v.Status = fs.StatusUnchanged
		}
	}

	return files, nil
}
