package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "docker-qa"
	app.Version = "1"
	app.Commands = []cli.Command{
		pushCommand,
		buildCommand,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
