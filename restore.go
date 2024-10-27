package g

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Restore(path string, staged bool) error {
	idx, err := ReadIndex()
	if err != nil {
		return err
	}
	currentCommit, err := CurrentCommit()
	if err != nil {
		return err
	}
	currentStatus, err := Status(idx, currentCommit)
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
	if !ok || fileStatus.WorkingDirectoryStatus() == Untracked {
		return fmt.Errorf("error: pathspec '%s' did not match any fileStatus(s) known to git", path)
	}
	// if in index but not committed
	if fileStatus.IndexStatus() == AddedInIndex && fileStatus.WorkingDirectoryStatus() != WorktreeChangedSinceIndex {
		// there is nothing to do, right ? ...
		return nil
	}

	// update working directory fileStatus with object referenced by index
	file := idx.File(path)
	if file == nil {
		// this should not happen
		return errors.New("index did not return file for some reason")
	}
	obj, err := ReadObject(file.index.Sha)
	if err != nil {
		return err
	}
	fh, err := os.OpenFile(filepath.Join(Path(), path), os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = fh.Close() }()
	reader, err := obj.ReadCloser()
	if err != nil {
		return err
	}
	if err := ReadHeadBytes(reader, obj); err != nil {
		return err
	}
	_, err = io.Copy(fh, reader)
	if err != nil {
		return err
	}
	if err := fh.Close(); err != nil {
		return err
	}
	return os.Chtimes(filepath.Join(Path(), path), file.index.Finfo.ModTime(), file.index.Finfo.ModTime())
}
