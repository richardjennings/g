package cmd

import (
	"github.com/richardjennings/mygit/internal/mygit"
	"github.com/spf13/cobra"
	"log"
)

var (
	gitDirectoryFlag string
	pathFlag         string
	rootCmd          = &cobra.Command{}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&gitDirectoryFlag, "git-directory", mygit.DefaultGitDirectory, "--git-directory")
	rootCmd.PersistentFlags().StringVar(&pathFlag, "path", mygit.DefaultPath, "--path")
}

func myGit() (*mygit.MyGit, error) {
	opts := []mygit.Opt{
		mygit.WithGitDirectory(gitDirectoryFlag),
		mygit.WithPath(pathFlag),
	}
	return mygit.NewMyGit(opts...)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
