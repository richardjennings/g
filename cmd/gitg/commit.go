package main

import (
	"fmt"
	"github.com/richardjennings/g/git"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var commitMessage string

var commitCmd = &cobra.Command{
	Use: "commit",
	Run: func(cmd *cobra.Command, args []string) {
		if err := configure(); err != nil {
			log.Fatalln(err)
		}
		var msg []byte
		if cmd.Flags().Changed("message") {
			msg = []byte(commitMessage)
		}
		sha, err := git.Commit(msg)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(sha.AsHexString())
	},
}

func init() {
	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "--message")
	rootCmd.AddCommand(commitCmd)
}
