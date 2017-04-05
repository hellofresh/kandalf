package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

const flagVersion = "version"

var (
	version    string
	configPath string
)

func failOnError(err error, msg string) {
	if err != nil {
		log.WithError(err).Panic(msg)
	}
}

func main() {
	versionString := "Kandalf v" + version
	var RootCmd = &cobra.Command{
		Use:   "kandalf",
		Short: versionString,
		Long: versionString + `. RabbitMQ to Kafka bridge.

Complete documentation is available at https://github.com/hellofresh/kandalf`,
		Run: RunApp,
	}
	RootCmd.Flags().StringVarP(&configPath, "config", "c", "", "Source of a configuration file")
	RootCmd.Flags().BoolP(flagVersion, "v", false, "Print application version")

	err := RootCmd.Execute()
	failOnError(err, "Failed to execute root command")
}
