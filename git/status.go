package git

import (
	"fmt"
	"github.com/richardjennings/g"
	"io"
)

// Status currently displays the file statuses comparing the working directory
// to the index and the index to the last commit (if any).
func Status(o io.Writer) error {
	var err error
	// index
	idx, err := g.ReadIndex()
	if err != nil {
		return err
	}

	commitSha, err := g.LastCommit()
	if err != nil {
		// @todo error types to check for e.g no previous commits as source of error
		return err
	}

	files, err := g.Status(idx, commitSha)

	if err != nil {
		return err
	}

	for _, v := range files.Files() {
		if v.IdxStatus == g.IndexNotUpdated && v.WdStatus == g.WDIndexAndWorkingTreeMatch {
			continue
		}
		if _, err := fmt.Fprintf(o, "%s%s %s\n", v.IdxStatus, v.WdStatus, v.Path); err != nil {
			return err
		}
	}

	return nil
}
