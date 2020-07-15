package main

import (
	"os"

	"github.com/ishii1648/grpc-poc/cmd/client/app"
)

func main() {
	command := app.NewClientCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
