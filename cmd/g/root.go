package main

import (
	"github.com/richardjennings/g"
	"github.com/spf13/cobra"
	"log"
)

var (
	gitDirectoryFlag string
	pathFlag         string
	rootCmd          = &cobra.Command{}
	repo             *g.Repository
)

func init() {
	rootCmd.PersistentFlags().StringVar(&gitDirectoryFlag, "git-directory", g.DefaultGitDirectory, "--git-directory")
	rootCmd.PersistentFlags().StringVar(&pathFlag, "path", g.DefaultPath, "--path")
}

func configure() error {
	var err error
	repo, err = g.Open(
		g.WithGitDirectory(gitDirectoryFlag),
		g.WithPath(pathFlag),
	)
	return err
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	Execute()
}
