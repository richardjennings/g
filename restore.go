package g

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func RestoreStaged(path string) error {
	// update the index with the sha from the last commit
	idx, err := ReadIndex()
	if err != nil {
		return err
	}
	status, err := CurrentStatus()
	if err != nil {
		return err
	}
	f, ok := status.idx[path]
	if !ok {
		return fmt.Errorf("file %s not found in index", path)
	}
	if f.commit == nil {
		// if the file is not commited at all, the correct behaviour of staged
		// is to simply remove the file form the index such that it is no longer
		// being tracked
		if err := idx.Rm(path); err != nil {
			return err
		}
		return idx.Write()
	}
	item, err := newItem(f.wd.Finfo, f.commit.Sha, f.path)
	if err != nil {
		return err
	}
	// Work Tree files do not have a hash associated until they are added to the
	// index. Therefore, we use filesystem modification time to compare work tree
	// files against the index. We do not know the previous actual modification
	// time from the commit - as git does not track timestamps like that outside
	// of the local index. So - just subtract 1 from the working directory
	// timestamp, so that the index is detected as different form the working
	// tree.
	item.MTimeN--
	if err := idx.upsertItem(item); err != nil {
		return err
	}
	return idx.Write()
}

func Restore(path string, staged bool) error {
	if staged {
		return RestoreStaged(path)
	}
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
