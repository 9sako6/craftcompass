package main

import (
	"os"

	"github.com/9sako6/craftcompass/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr))
}
