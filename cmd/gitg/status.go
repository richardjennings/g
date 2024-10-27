package main

import (
	"fmt"
	"github.com/richardjennings/g"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var statusCmd = &cobra.Command{
	Use: "status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			return err
		}
		return Status(os.Stdout)
	},
}

// Status currently displays the file statuses comparing the working directory
// to the index and the index to the last commit (if any).
func Status(o io.Writer) error {
	files, err := g.CurrentStatus()
	if err != nil {
		return err
	}
	for _, v := range files.Files() {
		if v.IndexStatus() == g.NotUpdated && v.WorkingDirectoryStatus() == g.IndexAndWorkingTreeMatch {
			continue
		}
		if _, err := fmt.Fprintf(o, "%s%s %s\n", v.IndexStatus().StatusString(), v.WorkingDirectoryStatus().StatusString(), v.Path()); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
