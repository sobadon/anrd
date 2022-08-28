package version

import (
	"github.com/sobadon/anrd/internal/logutil"
	"github.com/spf13/cobra"
)

var (
	log = logutil.NewLogger()
)

func Command() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "version",
		Short: "show version",
		RunE: func(cmd *cobra.Command, args []string) error {
			ShowVersion()
			return nil
		},
	}
	return rootCmd
}

func ShowVersion() {
	log.Info().Msgf("version: %s", version)
}
