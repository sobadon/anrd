package run

import (
	"context"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/sobadon/anrd/infrastructures/onsen"
	"github.com/sobadon/anrd/infrastructures/sqlite"
	"github.com/sobadon/anrd/usecase"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "run",
		Short: "run components",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
	return rootCmd
}

func run() error {
	var config config
	err := env.Parse(&config, env.Options{
		Prefix: "ATR_",
		OnSet: func(tag string, value interface{}, isDefault bool) {
			log.Printf("Set %s to %v (default? %v)\n", tag, value, isDefault)
		},
	})
	if err != nil {
		return err
	}

	db, err := sqlite.NewDB(config.SqlitePath)
	if err != nil {
		return err
	}

	err = sqlite.Setup(db)
	if err != nil {
		return err
	}

	infraProgramPersistence := sqlite.New(db)
	stationOnsen := onsen.New()
	ucRecorder := usecase.NewRecorder(infraProgramPersistence, stationOnsen)

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 1*time.Minute)

	// とりあえず番組情報取得のみ
	err = ucRecorder.UpdateProgram(ctx)
	if err != nil {
		return err
	}

	return nil
}
