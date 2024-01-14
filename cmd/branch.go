package cmd

import (
	"github.com/richardjennings/mygit/internal/mygit"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var branchCmd = &cobra.Command{
	Use:  "branch <path> ...",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			log.Fatalln(err)
		}
		if len(args) == 0 {
			// default to list branches
			return mygit.ListBranches(os.Stdout)
		}
		if len(args) == 1 {
			// create a branch
			return mygit.CreateBranch(args[0])
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(branchCmd)
}
