package g

import "fmt"

// Commit writes the Commit provided in the Object Store.
func (r *Repository) Commit(commit *Commit) (Sha, error) {
	idx, err := r.Index()
	if err != nil {
		return Sha{}, fmt.Errorf("reading index: %w", err)
	}
	idxFiles, err := idx.Files()
	if err != nil {
		return Sha{}, fmt.Errorf("reading index files: %w", err)
	}
	root := ObjectTree(idxFiles, r.workingDirectory())
	tree, err := r.writeTreeRecursive(root)
	if err != nil {
		return Sha{}, fmt.Errorf("writing tree: %w", err)
	}
	previousCommits, err := r.previousCommits()
	if err != nil {
		return Sha{}, fmt.Errorf("reading previous commits: %w", err)
	}
	commit.Tree = tree
	commit.Parents = previousCommits
	return r.writeCommit(commit)
}

// Free-function shims for migration.
func CreateCommit(commit *Commit) (Sha, error) { return defaultRepo.Commit(commit) }
