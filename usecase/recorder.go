package usecase

import (
	"context"
	"errors"
	"flag"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sobadon/anrd/domain/model/date"
	"github.com/sobadon/anrd/domain/model/program"
	"github.com/sobadon/anrd/domain/model/recorder"
	"github.com/sobadon/anrd/domain/repository"
	"github.com/sobadon/anrd/internal/errutil"
	"github.com/sobadon/anrd/internal/timeutil"
)

type ucRecorder struct {
	programPersistence repository.ProgramPersistence
	onsen              repository.Station
	agqr               repository.Station
}

func NewRecorder(
	programPersistence repository.ProgramPersistence,
	onsen repository.Station,
	agqr repository.Station,
) *ucRecorder {
	return &ucRecorder{
		programPersistence: programPersistence,
		onsen:              onsen,
		agqr:               agqr,
	}
}

func (r *ucRecorder) UpdateProgram(ctx context.Context) error {
	var pgrams []program.Program

	// 音泉
	pgrams_temp, err := r.onsen.GetPrograms(ctx, date.Date{})
	if err != nil {
		return err
	}
	pgrams = append(pgrams, pgrams_temp...)

	// agqr
	now := time.Now().In(timeutil.LocationJST())
	pgrams_temp, err = r.agqr.GetPrograms(ctx, date.NewFromToday(now))
	if err != nil {
		return err
	}
	pgrams = append(pgrams, pgrams_temp...)

	for _, pgram := range pgrams {
		err = r.programPersistence.Save(ctx, pgram)
		if err != nil {
			return err
		}
	}

	log.Ctx(ctx).Info().Msg("successfully update program")
	return nil
}

func (r *ucRecorder) RecBroadcastPrepare(ctx context.Context, config recorder.Config, now time.Time) error {
	targetPgrams, err := r.programPersistence.LoadBroadcastStartIn(ctx, now, config.PrepareAfter)
	if errors.As(err, &errutil.ErrDatabaseNotFoundProgram) {
		log.Ctx(ctx).Debug().Msg("not found program")
		return nil
	}
	if err != nil {
		return err
	}

	for _, targetPgram := range targetPgrams {
		go r.rec(ctx, config, now, targetPgram)
	}

	return nil
}

func (r *ucRecorder) RecOndemandPrepare(ctx context.Context, config recorder.Config, now time.Time) error {
	// あまり多いとリソース割かれちゃうので、ひとまず一気に 2 件
	targetPgrams, err := r.programPersistence.LoadOndemandScheduled(ctx, 2)
	if errors.As(err, &errutil.ErrDatabaseNotFoundProgram) {
		log.Ctx(ctx).Debug().Msg("not found program")
		return nil
	}
	if err != nil {
		return err
	}

	for _, targetPgram := range *targetPgrams {
		go r.rec(ctx, config, now, targetPgram)
		// 一気に録画開始は負荷高そうなので気持ちズラす
		time.Sleep(30 * time.Second)
	}

	return nil
}

// 録画（録音）処理を呼び出す
// 内部でリトライあり
// これは goroutine として呼び出されることを想定
// エラーが発生すればこの関数内でログ出力してしまう
func (r *ucRecorder) rec(ctx context.Context, config recorder.Config, now time.Time, targetPgram program.Program) {
	// retryCount=0, 1, 2, 3 の計 4 回トライする
	const retryMaxCount = 3
	retryCount := 0

	err := r.programPersistence.ChangeStatus(ctx, targetPgram, program.StatusRecording)
	if err != nil {
		return
	}

	if targetPgram.StreamType == program.StreamTypeBroadcast {
		// ffmpeg 叩き前の sleep
		sleepDuration := targetPgram.Start.Sub(now) - config.Margin
		log.Ctx(ctx).Debug().Msgf("sleep ... (duration = %s)", sleepDuration)
		if flag.Lookup("test.v") == nil {
			// テスト実行時に time.Sleep() されると困っちゃうから無理くり無効に
			time.Sleep(sleepDuration)
		}
	}

	for retryCount <= retryMaxCount {
		var err error
		switch targetPgram.Station {
		case program.StationOnsen:
			err = r.onsen.Rec(ctx, config, targetPgram)
		}
		if err == nil {
			log.Ctx(ctx).Info().Msgf("successfully rec program (program = %+v)", targetPgram)
			err := r.programPersistence.ChangeStatus(ctx, targetPgram, program.StatusDone)
			if err != nil {
				log.Ctx(ctx).Error().Msgf("%+v", err)
				return
			}
			return
		}

		log.Ctx(ctx).Warn().Msgf("failed to rec (retryCount = %d)", retryCount)
		retryCount++
	}

	log.Ctx(ctx).Error().Msgf("rec retry count exceeded retryMaxCount (program = %+v)", targetPgram)
	err = r.programPersistence.ChangeStatus(ctx, targetPgram, program.StatusFailed)
	if err != nil {
		log.Printf("%+v", err)
	}
}
