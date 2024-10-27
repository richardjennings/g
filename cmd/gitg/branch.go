package main

import (
	"fmt"
	"github.com/richardjennings/g"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var branchDelete bool

var branchCmd = &cobra.Command{
	Use:  "branch <path> ...",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			return err
		}
		if len(args) == 0 {
			// default to list branches
			return ListBranches(os.Stdout)
		}
		if len(args) == 1 {
			if branchDelete {
				return DeleteBranch(args[0])
			} else {
				// create a branch
				return CreateBranch(args[0])
			}
		}
		return nil
	},
}

const DeleteBranchCheckedOutErrFmt = "error: Cannot delete branch '%s' checked out at '%s'"

func DeleteBranch(name string) error {
	// Delete Branch removes any branch that is not checked out
	// @todo more correct semantics
	currentBranch, err := g.CurrentBranch()
	if err != nil {
		return err
	}
	if name == currentBranch {
		return fmt.Errorf(DeleteBranchCheckedOutErrFmt, name, g.Path())
	}
	return g.DeleteBranch(name)
}

func CreateBranch(name string) error {
	return g.CreateBranch(name)
}

func ListBranches(o io.Writer) error {
	var err error
	currentBranch, err := g.CurrentBranch()
	if err != nil {
		return err
	}
	branches, err := g.ListBranches()
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

func init() {
	branchCmd.Flags().BoolVarP(&branchDelete, "delete", "d", false, "--delete <branch>")
	rootCmd.AddCommand(branchCmd)
}
