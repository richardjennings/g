package commits

import (
	"github.com/richardjennings/mygit/internal/mygit/refs"
	"time"
)

type (
	Commit struct {
		Tree          []byte
		Parents       []string
		Author        string
		AuthoredTime  time.Time
		Committer     string
		CommittedTime time.Time
		Message       string
	}
)

func PreviousCommits() ([]string, error) {
	// @todo check for a previous commit better
	currentBranch, err := refs.CurrentBranch()
	if err != nil {
		return nil, err
	}
	var previousCommits []string
	previousCommit, err := refs.HeadSHA(string(currentBranch))
	if err != nil {
		return nil, err
	}
	if previousCommit != nil {
		previousCommits = append(previousCommits, string(previousCommit))
	}
	return previousCommits, nil
}
