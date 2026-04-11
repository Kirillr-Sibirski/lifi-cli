package main

import (
	"os"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
