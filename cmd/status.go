package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"os"
)

var statusCmd = &cobra.Command{
	Use: "status",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := myGit()
		if err != nil {
			log.Fatalln(err)
		}
		return m.Status(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
