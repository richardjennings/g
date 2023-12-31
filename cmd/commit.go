package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var commitCmd = &cobra.Command{
	Use: "commit",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := myGit()
		if err != nil {
			log.Fatalln(err)
		}
		return m.Commit()
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
}
