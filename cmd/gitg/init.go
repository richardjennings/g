package main

import (
	"github.com/richardjennings/g"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use: "init",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			return err
		}
		return Init()
	},
}

func Init() error {
	return g.Init()
}

func init() {
	rootCmd.AddCommand(initCmd)
}
