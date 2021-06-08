package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	cs "github.com/webtor-io/common-services"
	s "github.com/webtor-io/download-progress/services"
)

func makeServeCMD() cli.Command {
	serveCmd := cli.Command{
		Name:    "serve",
		Aliases: []string{"s"},
		Usage:   "Serves web server",
		Action:  serve,
	}
	configureServe(&serveCmd)
	return serveCmd
}

func configureServe(c *cli.Command) {
	c.Flags = s.RegisterWebFlags([]cli.Flag{})
	c.Flags = cs.RegisterProbeFlags(c.Flags)
	c.Flags = s.RegisterGRPCFlags(c.Flags)
}

func serve(c *cli.Context) error {
	// Setting Probe
	probe := cs.NewProbe(c)
	defer probe.Close()

	// Setting StatPool
	sp := s.NewStatPool()

	// Setting WriterPool
	wp := s.NewWriterPool(sp)

	// Setting Web
	web := s.NewWeb(c, wp, sp)
	defer web.Close()

	// Setting GRPC
	grpc := s.NewGRPC(c, sp)
	defer grpc.Close()

	// Setting ServeService
	serve := cs.NewServe(probe, web, grpc)

	// And SERVE!
	err := serve.Serve()
	if err != nil {
		log.WithError(err).Error("Got server error")
	}
	return err
}
