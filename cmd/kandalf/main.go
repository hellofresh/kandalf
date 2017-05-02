package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/spf13/cobra"
)

const flagVersion = "version"

var (
	version     string
	configPath  string
	versionFlag bool
)

func failOnError(err error, msg string) {
	if err != nil {
		log.WithError(err).Panic(msg)
	}
}

func main() {
	versionString := "Kandalf v" + version
	cobra.OnInitialize(func() {
		if versionFlag {
			fmt.Println(versionString)
			os.Exit(0)
		}

		log.SetFormatter(&logrus_logstash.LogstashFormatter{
			Type:            "Kandalf",
			TimestampFormat: time.RFC3339Nano,
		})
	})

	var RootCmd = &cobra.Command{
		Use:   "kandalf",
		Short: versionString,
		Long: versionString + `. RabbitMQ to Kafka bridge.

Complete documentation is available at https://github.com/hellofresh/kandalf`,
		Run: RunApp,
	}
	RootCmd.Flags().StringVarP(&configPath, "config", "c", "", "Source of a configuration file")
	RootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print application version")

	err := RootCmd.Execute()
	failOnError(err, "Failed to execute root command")
}
