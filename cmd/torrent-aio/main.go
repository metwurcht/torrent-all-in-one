package main

import (
	"os"

	"github.com/metwurcht/torrent-all-in-one/internal/cli"
)

// version est injectée à la compilation via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	if err := cli.Execute(version); err != nil {
		os.Exit(1)
	}
}
