package main

import (
	"github.com/richardjennings/g/git"
	"github.com/spf13/cobra"
	"log"
)

var initCmd = &cobra.Command{
	Use: "init",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			log.Fatalln(err)
		}
		return git.Init()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
