package usecase

import (
	"context"
	"errors"
	"log"

	"github.com/sobadon/anrd/domain/model/date"
	"github.com/sobadon/anrd/domain/model/program"
	"github.com/sobadon/anrd/domain/model/recorder"
	"github.com/sobadon/anrd/domain/repository"
	"github.com/sobadon/anrd/internal/errutil"
)

type ucRecorder struct {
	programPersistence repository.ProgramPersistence
	onsen              repository.Station
}

func NewRecorder(
	programPersistence repository.ProgramPersistence,
	onsen repository.Station,
) *ucRecorder {
	return &ucRecorder{
		programPersistence: programPersistence,
		onsen:              onsen,
	}
}

func (r *ucRecorder) UpdateProgram(ctx context.Context) error {
	pgrams, err := r.onsen.GetPrograms(ctx, date.Date{})
	if err != nil {
		return err
	}

	for _, pgram := range pgrams {
		err = r.programPersistence.Save(ctx, pgram)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ucRecorder) RecPrepare(ctx context.Context, config recorder.Config) error {
	targetPgram, err := r.programPersistence.LoadOndemandScheduled(ctx)
	if errors.As(err, &errutil.ErrDatabaseNotFoundProgram) {
		log.Println("not found program")
		return nil
	}
	if err != nil {
		return err
	}

	log.Printf("RecPrepare: %+v", targetPgram)

	go r.rec(ctx, config, *targetPgram)
	return nil
}

// 録画（録音）処理を呼び出す
// 内部でリトライあり
// これは goroutine として呼び出されることを想定
// エラーが発生すればこの関数内でログ出力してしまう
func (r *ucRecorder) rec(ctx context.Context, config recorder.Config, targetPgram program.Program) {
	// retryCount=0, 1, 2, 3 の計 4 回トライする
	const retryMaxCount = 3
	retryCount := 0

	err := r.programPersistence.ChangeStatus(ctx, targetPgram, program.StatusRecording)
	if err != nil {
		return
	}

	for retryCount <= retryMaxCount {
		var err error
		switch targetPgram.Station {
		case program.StationOnsen:
			err = r.onsen.Rec(ctx, config, targetPgram)
		}
		if err == nil {
			err := r.programPersistence.ChangeStatus(ctx, targetPgram, program.StatusDone)
			if err != nil {
				log.Printf("%+v", err)
				return
			}
			return
		}

		log.Printf("fail to rec")
		retryCount++
	}

	err = r.programPersistence.ChangeStatus(ctx, targetPgram, program.StatusFailed)
	if err != nil {
		log.Printf("%+v", err)
	}
}
