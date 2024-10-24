package main

import (
	"fmt"
	"github.com/richardjennings/g/git"
	"github.com/spf13/cobra"
	"os"
)

var statusCmd = &cobra.Command{
	Use: "status",
	Run: func(cmd *cobra.Command, args []string) {
		if err := configure(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := git.Status(os.Stdout); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
