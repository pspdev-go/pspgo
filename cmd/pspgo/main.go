package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pspdev-go/pspgo/internal/app"
)

func main() {
	if err := app.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "pspgo:", err)
		os.Exit(1)
	}
}
