package main

import (
	"github.com/richardjennings/g"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use: "init",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		repo, err = g.Init(
			g.WithGitDirectory(gitDirectoryFlag),
			g.WithPath(pathFlag),
		)
		return err
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
