package main

import (
	"fmt"
	"github.com/richardjennings/g/git"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var switchCmd = &cobra.Command{
	Use:  "switch",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := configure(); err != nil {
			log.Fatalln(err)
		}
		if err := git.SwitchBranch(args[0]); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
