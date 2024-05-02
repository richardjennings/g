package git

import (
	"fmt"
	"github.com/richardjennings/mygit/pkg/mygit/config"
	"github.com/richardjennings/mygit/pkg/mygit/refs"
	"io"
)

const DeleteBranchCheckedOutErrFmt = "error: Cannot delete branch '%s' checked out at '%s'"

func DeleteBranch(name string) error {
	// Delete Branch removes any branch that is not checked out
	// @todo more correct semantics
	currentBranch, err := refs.CurrentBranch()
	if err != nil {
		return err
	}
	if name == currentBranch {
		return fmt.Errorf(DeleteBranchCheckedOutErrFmt, name, config.Path())
	}
	return refs.DeleteBranch(name)
}

func CreateBranch(name string) error {
	return refs.CreateBranch(name)
}

func ListBranches(o io.Writer) error {
	var err error
	currentBranch, err := refs.CurrentBranch()
	if err != nil {
		return err
	}
	branches, err := refs.ListBranches()
	if err != nil {
		return err
	}
	for _, v := range branches {
		if v == currentBranch {
			_, err = o.Write([]byte(fmt.Sprintf("* %v\n", v)))
		} else {
			_, err = o.Write([]byte(fmt.Sprintf("  %v\n", v)))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

