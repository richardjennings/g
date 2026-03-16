package main

import (
	"fmt"
	"github.com/richardjennings/g"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
)

var logCmd = &cobra.Command{
	Use:  "log",
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			return err
		}
		cmdPath, cmdArgs := g.Pager()
		c := exec.Command(cmdPath, cmdArgs...)
		w, err := c.StdinPipe()
		if err != nil {
			return err
		}
		c.Stdout = os.Stdout
		err = Log(w)
		if err != nil {
			return err
		}
		w.Close()
		return c.Run()
	},
}

// Log prints out the commit log for the current branch
func Log(o io.Writer) error {
	branch, err := repo.Branch()
	if err != nil {
		return err
	}
	commitSha, err := repo.Head(branch)
	if err != nil {
		return err
	}
	for c, err := repo.ReadCommit(commitSha); c != nil && err == nil; c, err = repo.ReadCommit(c.Parents[0]) {
		_, _ = fmt.Fprintf(o, "commit %s\nAuthor: %s <%s>\nDate:   %s\n\n%8s\n", c.Sha, c.Author, c.AuthorEmail, c.AuthoredTime.String(), c.Message)
		if len(c.Parents) == 0 {
			break
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(logCmd)
}
