package cmd

import (
	"fmt"
	"github.com/richardjennings/mygit/pkg/mygit"
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
		if err := mygit.Restore(args[0], restoreStaged); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	restoreCmd.Flags().BoolVar(&restoreStaged, "staged", true, "--staged")
	rootCmd.AddCommand(restoreCmd)
}
