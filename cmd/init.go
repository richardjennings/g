package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var initCmd = &cobra.Command{
	Use: "init",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := myGit()
		if err != nil {
			log.Fatalln(err)
		}
		return m.Init()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
