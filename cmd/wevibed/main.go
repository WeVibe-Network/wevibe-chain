package main

import (
	"os"

	"github.com/wevibe-network/wevibe-chain/cmd/wevibed/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
