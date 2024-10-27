package main

import (
	"github.com/richardjennings/g"
	"github.com/spf13/cobra"
)

var restoreStaged bool

var restoreCmd = &cobra.Command{
	Use:  "restore",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			return err
		}
		return g.Restore(args[0], restoreStaged)
	},
}

func init() {
	restoreCmd.Flags().BoolVar(&restoreStaged, "staged", true, "--staged")
	rootCmd.AddCommand(restoreCmd)
}
