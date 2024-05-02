package cmd

import (
	"fmt"
	"github.com/richardjennings/mygit/pkg/mygit"
	"github.com/spf13/cobra"
	"log"
)

var lsFilesCmd = &cobra.Command{
	Use: "ls-files",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			log.Fatalln(err)
		}
		files, err := mygit.LsFiles()
		if err != nil {
			return err
		}
		for _, v := range files {
			fmt.Println(v)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsFilesCmd)
}
