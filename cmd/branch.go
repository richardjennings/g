package cmd

import (
	"github.com/richardjennings/mygit/internal/mygit"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var branchCmd = &cobra.Command{
	Use:  "branch <path> ...",
	Args: cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			log.Fatalln(err)
		}
		return mygit.ListBranches(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(branchCmd)
}
