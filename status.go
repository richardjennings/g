package g

func CurrentStatus() (*FileSet, error) {
	// index
	idx, err := ReadIndex()
	if err != nil {
		return nil, err
	}

	commitSha, err := LastCommit()
	if err != nil {
		// @todo error types to check for e.g no previous commits as source of error
		return nil, err
	}
	return Status(idx, commitSha)
}

// Status returns a FileSet containing all files from commit, index and working directory
// with the corresponding status.
func Status(idx *Index, commitSha Sha) (*FileSet, error) {
	var commitFiles []*File
	var err error
	if commitSha.IsSet() {
		commitFiles, err = CommittedFiles(commitSha)
		if err != nil {
			return nil, err
		}
	}
	files := NewFileSet(commitFiles)
	files.MergeFromIndex(NewFileSet(idx.Files()))
	workingDirectoryFiles, err := Ls(Path())
	if err != nil {
		return nil, err
	}
	files.MergeFromWD(NewFileSet(workingDirectoryFiles))
	return files, nil
}
