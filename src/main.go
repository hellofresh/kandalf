package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/urfave/cli"

	"./kandalf/config"
	"./kandalf/logger"
	"./kandalf/pipes"
)

// Instantiates new application and launches it
func main() {
	app := cli.NewApp()

	app.Name = "kandalf"
	app.Usage = "Daemon that reads all messages from RabbitMQ and put them to kafka"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		{
			Name:  "Nikita Vershinin",
			Email: "endeveit@gmail.com",
		},
	}
	app.Action = actionRun
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "/etc/kandalf/config.yml",
			Usage: "Path to the configuration file",
		},
		cli.StringFlag{
			Name:  "pipes, p",
			Value: "/etc/kandalf/pipes.yml",
			Usage: "Path to file with pipes rules",
		},
	}

	app.Run(os.Args)
}

// Runs the application
func actionRun(ctx *cli.Context) error {
	var (
		wg      *sync.WaitGroup = &sync.WaitGroup{}
		die     chan bool       = make(chan bool, 1)
		pConfig string          = ctx.String("config")
		pPipes  string          = ctx.String("pipes")
	)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP)

	go func() {
		for {
			sig := <-ch
			switch sig {
			case os.Interrupt:
				logger.Instance().Info("Got interrupt signal. Will stop the work")
				close(die)
			case syscall.SIGHUP:
				doReload(pConfig, pPipes)
			}
		}
	}()

	doReload(pConfig, pPipes)

	wg.Wait()

	return nil
}

// Reloads configuration and lists of available pipes
func doReload(pConfig, pPipes string) {
	config.Instance(pConfig)
	_ = pipes.All(pPipes)
}
