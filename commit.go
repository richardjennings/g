package g

// CreateCommit writes the Commit provided in the Object Store
func CreateCommit(commit *Commit) (Sha, error) {
	idx, err := ReadIndex()
	if err != nil {
		return Sha{}, err
	}
	root := ObjectTree(idx.Files())
	tree, err := root.WriteTree()
	if err != nil {
		return Sha{}, err
	}
	previousCommits, err := PreviousCommits()
	if err != nil {
		return Sha{}, err
	}
	commit.Tree = tree
	commit.Parents = previousCommits
	return writeCommit(commit)
}
