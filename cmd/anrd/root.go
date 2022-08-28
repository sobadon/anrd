package main

import (
	"log"

	"github.com/sobadon/anrd/cmd/anrd/run"
	"github.com/sobadon/anrd/cmd/anrd/version"
	"github.com/spf13/cobra"
)

var (
	flagVersion bool
)

func main() {
	execute()
}

func execute() {
	var rootCmd = &cobra.Command{
		Use:   "anrd",
		Short: "rec aniradi",
	}

	rootCmd.Run = runRoot
	rootCmd.AddCommand(run.Command())
	rootCmd.AddCommand(version.Command())

	rootCmd.Flags().BoolVarP(&flagVersion, "version", "V", false, "Print the version number")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("%+v", err)
	}
}

func runRoot(cmd *cobra.Command, args []string) {
	if flagVersion {
		version.ShowVersion()
		return
	}
}
