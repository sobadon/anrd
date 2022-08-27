package usecase

import (
	"context"

	"github.com/sobadon/anrd/domain/model/date"
	"github.com/sobadon/anrd/domain/repository"
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
