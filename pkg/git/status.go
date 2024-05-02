package git

import (
	"fmt"
	"github.com/richardjennings/mygit/pkg/gfs"
	"github.com/richardjennings/mygit/pkg/index"
	"github.com/richardjennings/mygit/pkg/refs"
	"io"
)

// Status currently displays the file statuses comparing the working directory
// to the index and the index to the last commit (if any).
func Status(o io.Writer) error {
	var err error
	// index
	idx, err := index.ReadIndex()
	if err != nil {
		return err
	}

	commitSha, err := refs.LastCommit()
	if err != nil {
		// @todo error types to check for e.g no previous commits as source of error
		return err
	}

	files, err := index.Status(idx, commitSha)

	if err != nil {
		return err
	}

	for _, v := range files.Files() {
		if v.IdxStatus == gfs.IndexNotUpdated && v.WdStatus == gfs.WDIndexAndWorkingTreeMatch {
			continue
		}
		if _, err := fmt.Fprintf(o, "%s%s %s\n", v.IdxStatus, v.WdStatus, v.Path); err != nil {
			return err
		}
	}

	return nil
}
