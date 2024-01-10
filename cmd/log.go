package cmd

import (
	"github.com/richardjennings/mygit/internal/mygit"
	"github.com/spf13/cobra"
	"log"
)

var logCmd = &cobra.Command{
	Use:  "log",
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			log.Fatalln(err)
		}
		return mygit.Log()
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}
