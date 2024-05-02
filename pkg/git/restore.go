package git

import (
	"errors"
	"fmt"
	"github.com/richardjennings/mygit/pkg/config"
	"github.com/richardjennings/mygit/pkg/gfs"
	"github.com/richardjennings/mygit/pkg/index"
	"github.com/richardjennings/mygit/pkg/objects"
	"github.com/richardjennings/mygit/pkg/refs"
	"io"
	"os"
	"path/filepath"
)

func Restore(path string, staged bool) error {
	idx, err := index.ReadIndex()
	if err != nil {
		return err
	}
	currentCommit, err := refs.LastCommit()
	if err != nil {
		return err
	}
	currentStatus, err := index.Status(idx, currentCommit)
	if err != nil {
		return err
	}
	if staged {
		// remove file from index
		if err := idx.Rm(path); err != nil {
			return err
		}
		return idx.Write()
	}

	fileStatus, ok := currentStatus.Contains(path)
	// if the path not found or is untracked working directory fileStatus then error
	if !ok || fileStatus.WdStatus == gfs.WDUntracked {
		return fmt.Errorf("error: pathspec '%s' did not match any fileStatus(s) known to git", path)
	}
	// if in index but not committed
	if fileStatus.IdxStatus == gfs.IndexAddedInIndex && fileStatus.WdStatus != gfs.WDWorktreeChangedSinceIndex {
		// there is nothing to do, right ? ...
		return nil
	}

	// update working directory fileStatus with object referenced by index
	file := idx.File(path)
	if file == nil {
		// this should not happen
		return errors.New("index did not return file for some reason")
	}
	obj, err := objects.ReadObject(file.Sha.AsHexBytes())
	if err != nil {
		return err
	}
	fh, err := os.OpenFile(filepath.Join(config.Path(), path), os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = fh.Close() }()
	reader, err := obj.ReadCloser()
	if err != nil {
		return err
	}
	if err := objects.ReadHeadBytes(reader, obj); err != nil {
		return err
	}
	_, err = io.Copy(fh, reader)
	if err != nil {
		return err
	}
	if err := fh.Close(); err != nil {
		return err
	}
	return os.Chtimes(filepath.Join(config.Path(), path), file.Finfo.ModTime(), file.Finfo.ModTime())
}
