package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

var lsFilesCmd = &cobra.Command{
	Use: "ls-files",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := myGit()
		if err != nil {
			log.Fatalln(err)
		}
		files, err := m.LsFiles()
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
