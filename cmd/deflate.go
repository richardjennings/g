package cmd

import (
	"compress/zlib"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
)

// print out zlib compressed file at absolute path
var deflateCmd = &cobra.Command{
	Use:  "deflate <abspath>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		f, err := os.Open(path)
		if err != nil {
			log.Fatalln(err)
		}
		z, err := zlib.NewReader(f)
		if err != nil {
			log.Fatalln(err)
		}
		_, err = io.Copy(os.Stdout, z)
		if err != nil {
			log.Fatalln(err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deflateCmd)
}
