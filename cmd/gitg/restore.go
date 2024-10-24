package main

import (
	"fmt"
	"github.com/richardjennings/g/git"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var restoreStaged bool

var restoreCmd = &cobra.Command{
	Use:  "restore",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := configure(); err != nil {
			log.Fatalln(err)
		}
		if err := git.Restore(args[0], restoreStaged); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	restoreCmd.Flags().BoolVar(&restoreStaged, "staged", true, "--staged")
	rootCmd.AddCommand(restoreCmd)
}
