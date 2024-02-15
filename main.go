// Package main provides the entry point for the application.
package main

import (
	"fmt"
	"os"

	"github.com/mojochao/emacscfg/app"
)

func main() {
	if err := app.New().Run(os.Args); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
