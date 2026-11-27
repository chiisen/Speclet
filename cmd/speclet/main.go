package main

import (
	"fmt"
	"io"
	"os"

	"github.com/chiisen/speclet/internal/cli"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, os.Getwd))
}

func run(args []string, stdout, stderr io.Writer, getwd func() (string, error)) int {
	if len(args) > 0 {
		switch args[0] {
		case "--version", "-h", "--help", "help":
			return cli.Run(args, stdout, stderr, "")
		}
	}
	cwd, err := getwd()
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "speclet: get working directory: "+err.Error())
		return 1
	}
	return cli.Run(args, stdout, stderr, cwd)
}
