package g

func CurrentStatus() (*FfileSet, error) {
	// index
	idx, err := ReadIndex()
	if err != nil {
		return nil, err
	}

	commitSha, err := CurrentCommit()
	if err != nil {
		return nil, err
	}
	return Status(idx, commitSha)
}

// Status returns a FfileSet containing all files from commit, index and working directory
// with the corresponding status.
func Status(idx *Index, commitSha Sha) (*FfileSet, error) {
	var commitFiles, indexFiles, wtFiles []*FileStatus
	var err error

	// set commit files
	if commitSha.IsSet() {
		commitFiles, err = CommittedFiles(commitSha)
		if err != nil {
			return nil, err
		}
	}

	indexFiles = idx.Files()

	// set working tree files
	wtFiles, err = Ls(Path())
	if err != nil {
		return nil, err
	}

	return NewFfileSet(commitFiles, indexFiles, wtFiles)
}
