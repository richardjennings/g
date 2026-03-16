package g

import "fmt"

// Status returns a FfileSet containing all files from commit, index and
// working directory with the corresponding status.
func (r *Repository) Status() (*FfileSet, error) {
	idx, err := r.Index()
	if err != nil {
		return nil, fmt.Errorf("reading index: %w", err)
	}

	commitSha, err := r.currentCommit()
	if err != nil {
		return nil, fmt.Errorf("reading current commit: %w", err)
	}
	return r.status(idx, commitSha)
}

func (r *Repository) status(idx *Index, commitSha Sha) (*FfileSet, error) {
	var commitFiles, indexFiles, wtFiles []*FileStatus
	var err error

	if commitSha.IsSet() {
		commitFiles, err = r.committedFiles(commitSha)
		if err != nil {
			return nil, fmt.Errorf("reading committed files: %w", err)
		}
	}

	indexFiles, err = idx.Files()
	if err != nil {
		return nil, fmt.Errorf("reading index files: %w", err)
	}

	wtFiles, err = r.ls(r.Path())
	if err != nil {
		return nil, fmt.Errorf("listing working tree: %w", err)
	}

	return NewFfileSet(commitFiles, indexFiles, wtFiles)
}

// Free-function shims for migration.
func CurrentStatus() (*FfileSet, error)                    { return defaultRepo.Status() }
func Status(idx *Index, commitSha Sha) (*FfileSet, error)  { return defaultRepo.status(idx, commitSha) }
