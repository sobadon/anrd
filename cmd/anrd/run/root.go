package run

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env/v6"
	zlog "github.com/rs/zerolog/log"
	"github.com/sobadon/anrd/domain/model/recorder"
	"github.com/sobadon/anrd/infrastructures/onsen"
	"github.com/sobadon/anrd/infrastructures/sqlite"
	"github.com/sobadon/anrd/internal/logutil"
	"github.com/sobadon/anrd/usecase"
	"github.com/spf13/cobra"
)

var (
	log = logutil.NewLogger()
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
	log.Info().Msg("start")

	var config config
	err := env.Parse(&config, env.Options{
		Prefix: "ATR_",
		OnSet: func(tag string, value interface{}, isDefault bool) {
			log.Info().Msgf("Set %s to %v (default? %v)\n", tag, value, isDefault)
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
	log.Info().Msg("setup done")

	infraProgramPersistence := sqlite.New(db)
	stationOnsen := onsen.New()
	ucRecorder := usecase.NewRecorder(infraProgramPersistence, stationOnsen)

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 1*time.Minute)

	jobUpdate := func(ctx context.Context) {
		ctx = logutil.NewLogger().With().
			Str("job", "update").
			Logger().WithContext(ctx)
		zlog.Ctx(ctx).Info().Msg("job start")
		err := ucRecorder.UpdateProgram(ctx)
		if err != nil {
			zlog.Ctx(ctx).Error().Msgf("%+v", err)
		}
	}

	// TODO: cron job
	jobUpdate(ctx)

	recorderConfig := recorder.Config{
		ArchiveDir: config.ArchiveDir,
		// TODO
	}

	jobRec := func(ctx context.Context) {
		ctx = logutil.NewLogger().With().
			Str("job", "rec").
			Logger().WithContext(ctx)

		err = ucRecorder.RecPrepare(ctx, recorderConfig)
		if err != nil {
			zlog.Ctx(ctx).Error().Msgf("%+v", err)
		}
	}

	// TODO: cron job
	jobRec(ctx)

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Info().Msg("Interrupt")
	defer db.Close()

	return nil
}
