package cmd

import (
	"encoding/hex"
	"fmt"
	"github.com/richardjennings/mygit/internal/mygit"
	"github.com/spf13/cobra"
	"log"
)

var commitCmd = &cobra.Command{
	Use: "commit",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := configure(); err != nil {
			log.Fatalln(err)
		}
		sha, err := mygit.Commit()
		if err != nil {
			return err
		}
		fmt.Println(hex.EncodeToString(sha))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
}
