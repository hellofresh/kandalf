package main

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/hellofresh/kandalf/cmd"
)

var (
	version    string
	configPath string
)

func main() {
	var RootCmd = &cobra.Command{
		Use:     "kandalf",
		Version: version,
		Short:   `RabbitMQ to Kafka bridge.`,
		Long: `RabbitMQ to Kafka bridge.

Complete documentation is available at https://github.com/hellofresh/kandalf`,
		RunE: func(c *cobra.Command, args []string) error {
			return cmd.RunApp(c.Context(), version, configPath)
		},
	}
	RootCmd.Flags().StringVarP(&configPath, "config", "c", "", "Source of a configuration file")

	if err := RootCmd.ExecuteContext(context.Background()); err != nil {
		log.Fatal(err)
	}
}
