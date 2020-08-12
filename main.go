package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/hellofresh/kandalf/cmd"
)

var (
	version     string
	configPath  string
	versionFlag bool
)

func main() {
	versionString := "Kandalf v" + version
	cobra.OnInitialize(func() {
		if versionFlag {
			fmt.Println(versionString)
			os.Exit(0)
		}
	})

	var RootCmd = &cobra.Command{
		Use:   "kandalf",
		Short: versionString,
		Long: versionString + `. RabbitMQ to Kafka bridge.

Complete documentation is available at https://github.com/hellofresh/kandalf`,
		RunE: func(c *cobra.Command, args []string) error {
			return cmd.RunApp(version, configPath)
		},
	}
	RootCmd.Flags().StringVarP(&configPath, "config", "c", "", "Source of a configuration file")
	RootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print application version")

	err := RootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
