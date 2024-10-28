package main

import (
	"fmt"
	"github.com/richardjennings/g"
	"github.com/spf13/cobra"
)

var lsFilesCmd = &cobra.Command{
	Use: "ls-files",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			return err
		}
		files, err := LsFiles()
		if err != nil {
			return err
		}
		for _, v := range files {
			fmt.Println(v)
		}
		return nil
	},
}

// LsFiles returns a list of files in the index
func LsFiles() ([]string, error) {
	idx, err := g.ReadIndex()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, v := range idx.Files() {
		files = append(files, v.Path())
	}
	return files, nil
}

func init() {
	rootCmd.AddCommand(lsFilesCmd)
}
