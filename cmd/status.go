package cmd

import (
	"github.com/richardjennings/mygit/internal/mygit"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var statusCmd = &cobra.Command{
	Use: "status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			log.Fatalln(err)
		}
		return mygit.Status(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
