package main

import (
	"fmt"
	"os"

	"github.com/simulot/iislogs"
)

func main() {
	app := iislogs.Application{}
	_, err := app.ParseCommandLine()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	app.Run()
}
