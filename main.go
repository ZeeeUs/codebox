package main

import (
	"log"
	"os"

	"github.com/ZeeeUs/codebox/internal/cli"
)

func main() {
	app := cli.New(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
