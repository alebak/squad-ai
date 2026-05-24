package main

import (
	"os"

	"github.com/alebak/squad-ai/internal/cli"
)

func main() {
	cli.AutoUpdate()

	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
