package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "download-progress"
	app.Usage = "Serves download progress service"
	app.Version = "0.0.1"
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	configure(app)
	err := app.Run(os.Args)
	if err != nil {
		log.WithError(err).Fatal("Failed to serve application")
	}
}
