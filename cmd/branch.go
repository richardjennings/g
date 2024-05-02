package cmd

import (
	"github.com/richardjennings/mygit/pkg/git"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var branchDelete bool

var branchCmd = &cobra.Command{
	Use:  "branch <path> ...",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			log.Fatalln(err)
		}
		if len(args) == 0 {
			// default to list branches
			return git.ListBranches(os.Stdout)
		}
		if len(args) == 1 {
			if branchDelete {
				return git.DeleteBranch(args[0])
			} else {
				// create a branch
				return git.CreateBranch(args[0])
			}
		}
		return nil
	},
}

func init() {
	branchCmd.Flags().BoolVarP(&branchDelete, "delete", "d", false, "--delete <branch>")
	rootCmd.AddCommand(branchCmd)
}
