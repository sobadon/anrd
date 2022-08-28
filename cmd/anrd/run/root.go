package run

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-co-op/gocron"
	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"
	"github.com/sobadon/anrd/domain/model/recorder"
	"github.com/sobadon/anrd/infrastructures/agqr"
	"github.com/sobadon/anrd/infrastructures/onsen"
	"github.com/sobadon/anrd/infrastructures/sqlite"
	"github.com/sobadon/anrd/internal/errutil"
	"github.com/sobadon/anrd/internal/logutil"
	"github.com/sobadon/anrd/internal/timeutil"
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
	stationAgqr := agqr.New()
	ucRecorder := usecase.NewRecorder(infraProgramPersistence, stationOnsen, stationAgqr)

	ctx := context.Background()
	scheduler := gocron.NewScheduler(timeutil.LocationJST())

	jobUpdate := func(ctx context.Context, job gocron.Job) {
		ctx = logutil.NewLogger().With().
			Int("job_count", job.RunCount()).
			Str("job", "update").
			Logger().WithContext(ctx)
		zlog.Ctx(ctx).Info().Msg("job start")
		err := ucRecorder.UpdateProgram(ctx)
		if err != nil {
			zlog.Ctx(ctx).Error().Msgf("%+v", err)
		}
	}
	_, err = scheduler.Every(29*time.Minute).DoWithJobDetails(jobUpdate, ctx)
	if err != nil {
		return errors.Wrap(errutil.ErrScheduler, err.Error())
	}

	recorderConfig := recorder.Config{
		ArchiveDir:   config.ArchiveDir,
		PrepareAfter: config.PrepareAfter,
		Margin:       config.Margin,
	}

	jobRecOndemand := func(ctx context.Context, job gocron.Job) {
		ctx = logutil.NewLogger().With().
			Int("job_count", job.RunCount()).
			Str("job", "rec_ondemand").
			Logger().WithContext(ctx)

		err = ucRecorder.RecOndemandPrepare(ctx, recorderConfig, time.Now().In(timeutil.LocationJST()))
		if err != nil {
			zlog.Ctx(ctx).Error().Msgf("%+v", err)
		}
	}
	_, err = scheduler.Every(5*time.Minute).DoWithJobDetails(jobRecOndemand, ctx)
	if err != nil {
		return errors.Wrap(errutil.ErrScheduler, err.Error())
	}

	jobRecBroadcast := func(ctx context.Context, job gocron.Job) {
		ctx = logutil.NewLogger().With().
			Int("job_count", job.RunCount()).
			Str("job", "rec_broadcast").
			Logger().WithContext(ctx)

		err = ucRecorder.RecBroadcastPrepare(ctx, recorderConfig, time.Now().In(timeutil.LocationJST()))
		if err != nil {
			zlog.Ctx(ctx).Error().Msgf("%+v", err)
		}
	}
	_, err = scheduler.Every(30*time.Second).DoWithJobDetails(jobRecBroadcast, ctx)
	if err != nil {
		return errors.Wrap(errutil.ErrScheduler, err.Error())
	}

	scheduler.StartAsync()
	scheduler.RunAllWithDelay(10 * time.Second)

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Info().Msg("Interrupt")
	defer db.Close()

	return nil
}
