package cmd

import (
	"github.com/richardjennings/mygit/internal/mygit"
	"github.com/spf13/cobra"
	"log"
)

var (
	gitDirectoryFlag string
	rootCmd          = &cobra.Command{}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&gitDirectoryFlag, "git-directory", mygit.DefaultGitDirectory, "--git-directory")
}

func myGit() *mygit.MyGit {
	opts := []mygit.Opt{
		mygit.WithGitDirectory(gitDirectoryFlag),
	}
	return mygit.NewMyGit(opts...)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
