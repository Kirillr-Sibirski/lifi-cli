package main

import (
	"os"

	"github.com/Kirillr-Sibirski/defi-mullet/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
