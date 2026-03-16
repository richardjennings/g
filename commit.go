package g

import "fmt"

// CreateCommit writes the Commit provided in the Object Store
func CreateCommit(commit *Commit) (Sha, error) {
	idx, err := ReadIndex()
	if err != nil {
		return Sha{}, fmt.Errorf("reading index: %w", err)
	}
	root := ObjectTree(idx.Files())
	tree, err := root.WriteTree()
	if err != nil {
		return Sha{}, fmt.Errorf("writing tree: %w", err)
	}
	previousCommits, err := PreviousCommits()
	if err != nil {
		return Sha{}, fmt.Errorf("reading previous commits: %w", err)
	}
	commit.Tree = tree
	commit.Parents = previousCommits
	return writeCommit(commit)
}
