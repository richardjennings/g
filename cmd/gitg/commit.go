package main

import (
	"errors"
	"fmt"
	"github.com/richardjennings/g"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"time"
)

var commitMessage string

var commitCmd = &cobra.Command{
	Use: "commit",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			log.Fatalln(err)
		}
		var msg []byte
		if cmd.Flags().Changed("message") {
			msg = []byte(commitMessage)
		}
		sha, err := Commit(msg)
		if err != nil {
			return err
		}
		fmt.Println(sha.AsHexString())
		return nil
	},
}

// Commit writes a git commit object from the files in the index
func Commit(message []byte) (g.Sha, error) {
	commit := &g.Commit{
		Author:        fmt.Sprintf("%s <%s>", g.AuthorName(), g.AuthorEmail()),
		AuthoredTime:  time.Now(),
		Committer:     fmt.Sprintf("%s <%s>", g.CommitterName(), g.CommitterEmail()),
		CommittedTime: time.Now(),
	}
	if message != nil {
		commit.Message = message
	} else {
		// empty commit file
		if err := os.WriteFile(g.EditorFile(), []byte{}, 0600); err != nil {
			log.Fatalln(err)
		}
		ed, args := g.Editor()
		args = append(args, g.EditorFile())
		cmd := exec.Command(ed, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			log.Fatalln(err)
		}
		msg, err := os.ReadFile(args[0])
		if err != nil {
			log.Fatalln(msg)
		}
		commit.Message = msg
	}

	if len(commit.Message) == 0 {
		return g.Sha{}, errors.New("aborting commit due to empty commit message")
	}
	return g.CreateCommit(commit)
}

func init() {
	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "--message")
	rootCmd.AddCommand(commitCmd)
}
