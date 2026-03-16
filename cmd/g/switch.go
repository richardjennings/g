package main

import (
	"fmt"
	"strings"

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
	errFiles, err := repo.Switch(name)
	if err != nil {
		return err
	}
	if len(errFiles) > 0 {
		return fmt.Errorf("error: The following untracked working tree files would be overwritten by checkout:\n\t%s\nPlease move or remove them before you switch branches.", strings.Join(errFiles, "\n\t"))
	}
	return nil
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
