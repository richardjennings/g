package git

import (
	"fmt"
	"github.com/richardjennings/g"
	"io"
)

// Log prints out the commit log for the current branch
func Log(o io.Writer) error {
	branch, err := g.CurrentBranch()
	if err != nil {
		return err
	}
	commitSha, err := g.HeadSHA(branch)
	if err != nil {
		return err
	}
	for c, err := g.ReadCommit(commitSha); c != nil && err == nil; c, err = g.ReadCommit(c.Parents[0]) {
		_, _ = fmt.Fprintf(o, "commit %s\nAuthor: %s <%s>\nDate:   %s\n\n%8s\n", c.Sha, c.Author, c.AuthorEmail, c.AuthoredTime.String(), c.Message)
		if len(c.Parents) == 0 {
			break
		}
	}

	return nil
}
