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
	previousCommit, err := refs.LastCommit()
	if err != nil {
		return nil, err
	}
	if previousCommit != nil {
		return []string{string(previousCommit)}, nil
	}
	return nil, nil
}
