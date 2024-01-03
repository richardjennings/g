package cmd

import (
	"github.com/richardjennings/mygit/internal/mygit"
	"github.com/spf13/cobra"
	"log"
)

var addCmd = &cobra.Command{
	Use:  "add <path> ...",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			log.Fatalln(err)
		}
		return mygit.Add(args...)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
