package g

import (
	"fmt"
	"os"
	"path/filepath"
)

// RestoreStaged removes a staged change from the index.
// If the file is in the previous commit, removing it from the index means
// updating the index to specify the commit sha. The timestamp for which would
// be the same as when the files were originally switched to.
//
// If the file is not in a previous commit, removing it from the index means
// simply removing it from the index.
func RestoreStaged(path string) error {
	status, err := CurrentStatus()
	if err != nil {
		return err
	}
	f, ok := status.idx[path]
	if !ok {
		return fmt.Errorf("file %s not found in index", path)
	}
	idx, err := ReadIndex()
	if err != nil {
		return err
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
	// @todo how does git handle this specfically ?
	item.MTimeS = item.MTimeS - 100
	item.CTimeS = item.CTimeS - 100
	if err := idx.upsertItem(item); err != nil {
		return err
	}
	return idx.Write()
}

func Restore(path string, staged bool) error {
	if staged {
		return RestoreStaged(path)
	}

	currentStatus, err := CurrentStatus()
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
		// there is nothing to do
		return nil
	}

	// write the file
	if err := writeObjectToWorkingTree(fileStatus.index.Sha, fileStatus.Path()); err != nil {
		return err
	}

	// update modification time to match index
	return os.Chtimes(filepath.Join(Path(), path), fileStatus.index.Finfo.ModTime(), fileStatus.index.Finfo.ModTime())
}
