package git

import (
	"fmt"
	"github.com/richardjennings/g"
	"io"
)

// Status currently displays the file statuses comparing the working directory
// to the index and the index to the last commit (if any).
func Status(o io.Writer) error {
	files, err := g.CurrentStatus()
	if err != nil {
		return err
	}
	for _, v := range files.Files() {
		if v.IndexStatus() == g.IndexNotUpdated && v.WorkingDirectoryStatus() == g.WDIndexAndWorkingTreeMatch {
			continue
		}
		if _, err := fmt.Fprintf(o, "%s%s %s\n", v.IndexStatus(), v.WorkingDirectoryStatus(), v.Path); err != nil {
			return err
		}
	}
	return nil
}
