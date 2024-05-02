package git

import (
	"fmt"
	"github.com/richardjennings/mygit/pkg/objects"
	"github.com/richardjennings/mygit/pkg/refs"
	"io"
)

// Log prints out the commit log for the current branch
func Log(o io.Writer) error {
	branch, err := refs.CurrentBranch()
	if err != nil {
		return err
	}
	commitSha, err := refs.HeadSHA(branch)
	if err != nil {
		return err
	}
	for c, err := objects.ReadCommit(commitSha); c != nil && err == nil; c, err = objects.ReadCommit(c.Parents[0]) {
		_, _ = fmt.Fprintf(o, "commit %s\nAuthor: %s <%s>\nDate:   %s\n\n%8s\n", c.Sha, c.Author, c.AuthorEmail, c.AuthoredTime.String(), c.Message)
		if len(c.Parents) == 0 {
			break
		}
	}

	return nil
}
