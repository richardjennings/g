package cmd

import (
	"github.com/richardjennings/mygit/pkg/mygit/config"
	"github.com/spf13/cobra"
	"log"
)

var (
	gitDirectoryFlag string
	pathFlag         string
	rootCmd          = &cobra.Command{}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&gitDirectoryFlag, "git-directory", config.DefaultGitDirectory, "--git-directory")
	rootCmd.PersistentFlags().StringVar(&pathFlag, "path", config.DefaultPath, "--path")
}

func configure() error {
	opts := []config.Opt{
		config.WithGitDirectory(gitDirectoryFlag),
		config.WithPath(pathFlag),
	}
	return config.Configure(opts...)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
