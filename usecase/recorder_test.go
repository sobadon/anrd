package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/sobadon/anrd/domain/model/program"
	"github.com/sobadon/anrd/domain/model/recorder"
	"github.com/sobadon/anrd/internal/errutil"
	"github.com/sobadon/anrd/internal/timeutil"
	mock_repository "github.com/sobadon/anrd/testdata/mock/domain/repository"
)

func Test_ucRecorder_rec(t *testing.T) {
	configCommon := recorder.Config{
		ArchiveDir:   "/archive",
		PrepareAfter: 2 * time.Minute,
		Margin:       1 * time.Minute,
	}

	pgramOndemand := program.Program{
		UUID:        "48e582f4-afd8-4a7b-9582-f479f94eff9e",
		ID:          11134,
		Station:     program.StationOnsen,
		Title:       "セブン-イレブン presents 佐倉としたい大西",
		Episode:     "第334回",
		Start:       time.Date(2022, 8, 23, 0, 0, 0, 0, timeutil.LocationJST()),
		End:         time.Time{},
		Status:      program.StatusScheduled,
		StreamType:  program.StreamTypeOndemand,
		PlaylistURL: "https://onsen.test/playlist.m3u8",
	}

	type fields struct {
		programPersistence *mock_repository.MockProgramPersistence
		onsen              *mock_repository.MockStation
	}
	type args struct {
		config      recorder.Config
		targetPgram program.Program
	}
	tests := []struct {
		name    string
		prepare func(f *fields)
		args    args
	}{
		// mock の呼び出し回数だけで、問題ない　or 問題ありを判断するものとする
		// この rec() が error を返さないので（返させたくなかったので）、こうなってしまった
		{
			name: "何ら問題なく録画に成功（onsen）",
			prepare: func(f *fields) {
				f.programPersistence.EXPECT().
					ChangeStatus(gomock.Any(), pgramOndemand, program.StatusRecording).
					Return(nil)
				f.onsen.EXPECT().
					Rec(gomock.Any(), configCommon, pgramOndemand).
					Return(nil)
				f.programPersistence.EXPECT().
					ChangeStatus(gomock.Any(), pgramOndemand, program.StatusDone).
					Return(nil)
			},
			args: args{
				config:      configCommon,
				targetPgram: pgramOndemand,
			},
		},
		{
			name: "一度 ffmpeg が異常終了したとしてもリトライによって録画を継続",
			prepare: func(f *fields) {
				f.programPersistence.EXPECT().
					ChangeStatus(gomock.Any(), pgramOndemand, program.StatusRecording).
					Return(nil)
				f.onsen.EXPECT().
					Rec(gomock.Any(), configCommon, pgramOndemand).
					Return(errors.Wrap(errutil.ErrFfmpeg, "something error")).
					Times(1)
				f.onsen.EXPECT().
					Rec(gomock.Any(), configCommon, pgramOndemand).
					Return(nil).
					Times(1)
				f.programPersistence.EXPECT().
					ChangeStatus(gomock.Any(), pgramOndemand, program.StatusDone).
					Return(nil)
			},
			args: args{
				config:      configCommon,
				targetPgram: pgramOndemand,
			},
		},
		{
			name: "リトライを最大回数実施したが変わらず異常であるので録画を異常終了させる",
			prepare: func(f *fields) {
				f.programPersistence.EXPECT().
					ChangeStatus(gomock.Any(), pgramOndemand, program.StatusRecording).
					Return(nil)
				f.onsen.EXPECT().
					Rec(gomock.Any(), configCommon, pgramOndemand).
					Return(errors.Wrap(errutil.ErrFfmpeg, "something error")).
					Times(4) // retryCount=0, 1, 2, 3 の計 4 回トライする
				f.programPersistence.EXPECT().
					ChangeStatus(gomock.Any(), pgramOndemand, program.StatusFailed).
					Return(nil)
			},
			args: args{
				config:      configCommon,
				targetPgram: pgramOndemand,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProgramPersistence := mock_repository.NewMockProgramPersistence(ctrl)
			mockOnsen := mock_repository.NewMockStation(ctrl)

			r := &ucRecorder{
				programPersistence: mockProgramPersistence,
				onsen:              mockOnsen,
			}
			f := &fields{
				programPersistence: mockProgramPersistence,
				onsen:              mockOnsen,
			}
			tt.prepare(f)
			r.rec(context.Background(), tt.args.config, tt.args.targetPgram)
		})
	}
}
