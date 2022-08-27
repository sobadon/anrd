package main

import (
	"log"

	"github.com/sobadon/anrd/cmd/anrd/run"
	"github.com/spf13/cobra"
)

func main() {
	execute()
}

func execute() {
	var rootCmd = &cobra.Command{
		Use:   "anrd",
		Short: "rec aniradi",
	}

	rootCmd.AddCommand(run.Command())

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("%+v", err)
	}
}
