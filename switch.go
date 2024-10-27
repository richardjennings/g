package g

import (
	"os"
	"path/filepath"
)

// switchBranchDelta describes the changes need to move between two branches
// Any files that prevent the switch operation from being safe are recorded as errorFiles.
type switchBranchDelta struct {
	add        []*FileStatus // Files to be added to the working directory
	addSkip    []*FileStatus // Files to be added that are already correct
	staged     []*FileStatus // Files that are staged in the index and not an error
	remove     []*FileStatus // Files to be removed from the working directory
	ignore     []*FileStatus // Working Directory changes that should be left alone
	errorFiles []*FileStatus // The following untracked working tree files would be overwritten by checkout ...
}

func newSwitchBranchDelta(name string) (*switchBranchDelta, error) {
	// the delta
	delta := &switchBranchDelta{}

	// get all files in working directory, index and current commit with the
	// index and wd statuses set.
	curFiles, err := CurrentStatus()
	if err != nil {
		return nil, err
	}

	// get all the files in the branch HEAD commit being switched to
	commitFiles, err := CommittedFilesForBranchHead(name)
	if err != nil {
		return nil, err
	}

	// @todo commitFiles returns gibberish idxStatus - which maybe does not
	// matter right now

	// compare and generate a Delta
	for _, c := range curFiles.Files() {

		// check if file is in new commit files
		n, ok := commitFiles.idx[c.path]

		if !ok {
			// this file is not in the new commit. We only need to remove it
			// if it is not modified from the previous branch head commit.
			if c.wdStatus == IndexAndWorkingTreeMatch && c.idxStatus == NotUpdated {
				delta.remove = append(delta.remove, c)
				continue
			}

			// if this file is added in the index, (and is not in the new commit)
			// we can safely add to the new index
			if c.idxStatus == AddedInIndex {
				// what if the working directory has changes ? can we change
				// branch leaving working directory changes and keep the index ?
				delta.staged = append(delta.staged, c)
				delta.ignore = append(delta.ignore, c)
				continue
			}

			// this file should be safe to ignore
			delta.ignore = append(delta.ignore, c)
			continue
		}

		if c.idxStatus == UpdatedInIndex {
			// @todo switch error conditions
			// this can be an error condition as the file is in the new commit
			// to determine if it is an error we need to determine if the file
			// has changed in the new commit
		}

		// the filepath is in both the current commit and the new commit

		// if the file has both local changes and has a different version in the new
		// commit, record as errorFile so as not to lose local changes.
		if c.wdStatus == WorktreeChangedSinceIndex {
			// now how to tell
			if c.commit.Sha.String() != n.commit.Sha.String() {
				delta.errorFiles = append(delta.errorFiles, c)
			}
		}
	}

	for _, n := range commitFiles.Files() {
		// check if file is in current files
		c, ok := curFiles.idx[n.path]
		if ok {
			// if the current branch has the same hash for the same commited
			// file - add to addSkip. @todo What about if the same hash is in
			// the index ?
			if c.commit != nil && c.commit.Sha.Matches(n.commit.Sha) {
				// and the file is consistent with the commit
				if c.wdStatus == IndexAndWorkingTreeMatch && c.idxStatus == NotUpdated {
					// no need to delete/recreate it
					delta.addSkip = append(delta.addSkip, c)
					continue
				}
			}
		}
		delta.add = append(delta.add, n)

	}

	return delta, nil
}

func SwitchBranch(name string) ([]string, error) {
	delta, err := newSwitchBranchDelta(name)
	if err != nil {
		return nil, err
	}
	if len(delta.errorFiles) != 0 {
		// do not return an error as detecting errorFiles may not be
		// an error condition
		paths := make([]string, len(delta.errorFiles))
		for i, v := range delta.errorFiles {
			paths[i] = v.path
		}
		return paths, nil
	}

	// remove the files that need to be removed
	for _, v := range delta.remove {
		if err := os.Remove(filepath.Join(Path(), v.Path())); err != nil {
			return nil, err
		}
	}

	// add the files that need to be added
	for _, v := range delta.add {
		if err := writeObjectToWorkingTree(v.commit.Sha, v.Path()); err != nil {
			return nil, err
		}
	}

	// rebuild the index
	idx := NewIndex()
	for _, v := range delta.addSkip {
		if err := idx.addFromCommit(v); err != nil {
			return nil, err
		}
	}
	for _, v := range delta.add {
		if err := idx.addFromCommit(v); err != nil {
			return nil, err
		}
	}
	for _, v := range delta.staged {
		if err := idx.addFromIndex(v); err != nil {
			return nil, err
		}
	}

	if err := idx.Write(); err != nil {
		return nil, err
	}

	// update HEAD
	if err := UpdateHead(name); err != nil {
		return nil, err
	}

	return nil, nil
}
