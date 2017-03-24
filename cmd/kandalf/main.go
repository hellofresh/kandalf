package main

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/hellofresh/kandalf/pkg/pipes"
	"github.com/hellofresh/kandalf/workers"
	"github.com/hellofresh/stats-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.WithError(err).Panic(msg)
	}
}

func main() {
	globalConfig := config.LoadEnv()

	level, err := log.ParseLevel(strings.ToLower(globalConfig.LogLevel))
	failOnError(err, "Failed to get log level")
	log.SetLevel(level)

	pipesList, err := config.LoadPipes(globalConfig.Kafka.PipesConfig)
	failOnError(err, "Failed to get log level")

	statsClient := stats.NewStatsdStatsClient(globalConfig.Stats.DSN, globalConfig.Stats.Prefix)
	defer statsClient.Close()

	var (
		wg  *sync.WaitGroup = &sync.WaitGroup{}
		die chan bool       = make(chan bool, 1)
	)

	worker := workers.NewWorker(globalConfig, pipesList, statsClient)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP)

	go func() {
		for {
			sig := <-ch
			switch sig {
			case os.Interrupt:
				log.Info("Got interrupt signal. Will stop the work")
				close(die)
			}
		}
	}()

	// Here be dragons
	wg.Add(1)
	go worker.Run(wg, die)
	wg.Wait()
}
