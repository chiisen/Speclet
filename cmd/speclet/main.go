package main

import (
	"os"

	"github.com/chiisen/speclet/internal/cli"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		_, _ = os.Stderr.WriteString("speclet: get working directory: " + err.Error() + "\n")
		os.Exit(1)
	}
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr, cwd))
}
