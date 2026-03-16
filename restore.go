package g

import (
	"fmt"
	"os"
	"path/filepath"
)

// restoreStaged removes a staged change from the index.
// If the file is in the previous commit, removing it from the index means
// updating the index to specify the commit sha. The timestamp for which would
// be the same as when the files were originally switched to.
//
// If the file is not in a previous commit, removing it from the index means
// simply removing it from the index.
func (r *Repository) restoreStaged(path string) error {
	status, err := r.Status()
	if err != nil {
		return fmt.Errorf("reading status: %w", err)
	}
	f, ok := status.idx[path]
	if !ok {
		return fmt.Errorf("file %s not found in index", path)
	}
	idx, err := r.Index()
	if err != nil {
		return fmt.Errorf("reading index: %w", err)
	}
	if f.commit == nil {
		// if the file is not commited at all, the correct behaviour of staged
		// is to simply remove the file form the index such that it is no longer
		// being tracked
		if err := idx.Rm(path); err != nil {
			return fmt.Errorf("removing from index: %w", err)
		}
		return idx.Write()
	}

	item, err := newItem(f.wd.Finfo, f.commit.Sha, f.path)
	if err != nil {
		return fmt.Errorf("creating index item: %w", err)
	}
	// @todo how does git handle this specfically ?
	item.MTimeS = item.MTimeS - 100
	item.CTimeS = item.CTimeS - 100
	if err := idx.upsertItem(item); err != nil {
		return fmt.Errorf("updating index: %w", err)
	}
	return idx.Write()
}

// RestoreStaged is a free-function shim that delegates to defaultRepo.
func RestoreStaged(path string) error { return defaultRepo.restoreStaged(path) }

// Restore restores a file in the working directory to match the index version,
// or restores a staged change if staged is true.
func (r *Repository) Restore(path string, staged bool) error {
	if staged {
		return r.restoreStaged(path)
	}

	currentStatus, err := r.Status()
	if err != nil {
		return fmt.Errorf("reading status: %w", err)
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
	if err := r.writeObjectToWorkingTree(fileStatus.index.Sha, fileStatus.Path()); err != nil {
		return fmt.Errorf("restoring %s: %w", path, err)
	}

	// update modification time to match index
	return os.Chtimes(filepath.Join(r.Path(), path), fileStatus.index.Finfo.ModTime(), fileStatus.index.Finfo.ModTime())
}

// Restore is a free-function shim that delegates to defaultRepo.
func Restore(path string, staged bool) error { return defaultRepo.Restore(path, staged) }
