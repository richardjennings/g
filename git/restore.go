package git

import (
	"errors"
	"fmt"
	"github.com/richardjennings/g"
	"io"
	"os"
	"path/filepath"
)

func Restore(path string, staged bool) error {
	idx, err := g.ReadIndex()
	if err != nil {
		return err
	}
	currentCommit, err := g.LastCommit()
	if err != nil {
		return err
	}
	currentStatus, err := g.Status(idx, currentCommit)
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
	if !ok || fileStatus.WorkingDirectoryStatus() == g.WDUntracked {
		return fmt.Errorf("error: pathspec '%s' did not match any fileStatus(s) known to git", path)
	}
	// if in index but not committed
	if fileStatus.IndexStatus() == g.IndexAddedInIndex && fileStatus.WorkingDirectoryStatus() != g.WDWorktreeChangedSinceIndex {
		// there is nothing to do, right ? ...
		return nil
	}

	// update working directory fileStatus with object referenced by index
	file := idx.File(path)
	if file == nil {
		// this should not happen
		return errors.New("index did not return file for some reason")
	}
	obj, err := g.ReadObject(file.Sha)
	if err != nil {
		return err
	}
	fh, err := os.OpenFile(filepath.Join(g.Path(), path), os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = fh.Close() }()
	reader, err := obj.ReadCloser()
	if err != nil {
		return err
	}
	if err := g.ReadHeadBytes(reader, obj); err != nil {
		return err
	}
	_, err = io.Copy(fh, reader)
	if err != nil {
		return err
	}
	if err := fh.Close(); err != nil {
		return err
	}
	return os.Chtimes(filepath.Join(g.Path(), path), file.Finfo.ModTime(), file.Finfo.ModTime())
}
