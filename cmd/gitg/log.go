package main

import (
	"github.com/richardjennings/g"
	"github.com/richardjennings/g/git"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
)

var logCmd = &cobra.Command{
	Use:  "log",
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			log.Fatalln(err)
		}
		cmdPath, cmdArgs := g.Pager()
		c := exec.Command(cmdPath, cmdArgs...)
		w, err := c.StdinPipe()
		if err != nil {
			return err
		}
		c.Stdout = os.Stdout
		err = git.Log(w)
		if err != nil {
			return err
		}
		w.Close()
		return c.Run()
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}
