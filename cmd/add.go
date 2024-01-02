package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var addCmd = &cobra.Command{
	Use:  "add <path> ...",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := myGit()
		if err != nil {
			log.Fatalln(err)
		}
		return m.Add(args...)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
