package main

import (
	"os"

	"github.com/metwurcht/torrent-all-in-one/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
