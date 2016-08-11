package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/urfave/cli"

	"kandalf/cluster"
	"kandalf/config"
	"kandalf/logger"
	"kandalf/pipes"
	"kandalf/workers"
)

// Instantiates new application and launches it
func main() {
	app := cli.NewApp()

	app.Name = "kandalf"
	app.Usage = "Daemon that reads all messages from RabbitMQ and puts them to kafka"
	// This will be replaced by `build/codeship/publish-release.sh`
	app.Version = "%app.version%"
	app.Authors = []cli.Author{
		{
			Name:  "Nikita Vershinin",
			Email: "endeveit@gmail.com",
		},
	}
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
	app.Action = actionRun

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unhandled error occurred while running application: %v\n", err)
	}
}

// Runs the application
func actionRun(ctx *cli.Context) (err error) {
	var (
		cl      *cluster.Cluster
		clNodes []string = []string{}

		wg      *sync.WaitGroup = &sync.WaitGroup{}
		die     chan bool       = make(chan bool, 1)
		pConfig string          = ctx.String("config")
		pPipes  string          = ctx.String("pipes")
		worker  *workers.Worker
	)

	doReload(pConfig, pPipes)

	clEnabled := config.Instance().UBool("cluster.enabled", false)
	if clEnabled {
		for _, n := range config.Instance().UList("cluster.nodes") {
			clNodes = append(clNodes, n.(string))
		}
		clEnabled = len(clNodes) > 0
	}

	worker = workers.NewWorker()

	if clEnabled {
		cl = cluster.NewCluster(worker, clNodes)
	}

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
				logger.Instance().Info("Got SIGHUP. Will reload config and pipes")
				doReload(pConfig, pPipes)

				if cl != nil {
					cl.Reload()
				} else {
					worker.Reload()
				}
			}
		}
	}()

	wg.Add(1)

	// Here be dragons
	if cl != nil {
		go cl.Run(wg, die)
	} else {
		go worker.Run(wg, die)
	}

	wg.Wait()

	return nil
}

// Reloads configuration and lists of available pipes
func doReload(pConfig, pPipes string) {
	config.Instance(pConfig)
	_ = pipes.All(pPipes)
}
