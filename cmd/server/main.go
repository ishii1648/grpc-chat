package main

import (
	"os"

	"github.com/ishii1648/grpc-poc/cmd/server/app"
)

func main() {
	command := app.NewServerCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
