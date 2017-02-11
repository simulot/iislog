package main

import (
	"fmt"
	"os"

	"github.com/simulot/iislog"
)

func main() {
	app := iislog.Application{}
	_, err := app.ParseCommandLine()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	app.Run()
}
