package main

import (
	"github.com/richardjennings/g"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:  "switch",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			return err
		}
		return SwitchBranch(args[0])
	},
}

func SwitchBranch(name string) error {
	errFiles, err := g.SwitchBranch(name)
	if err != nil {
		return err
	}
	// @todo print out errFiles
	_ = errFiles
	return nil
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
