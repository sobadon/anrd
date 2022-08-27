package onsen

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sobadon/anrd/domain/model/date"
	"github.com/sobadon/anrd/domain/model/program"
	"github.com/sobadon/anrd/internal/timeutil"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

func Test_buildDateFromMMDD(t *testing.T) {
	type args struct {
		now  time.Time
		mmdd string
	}
	tests := []struct {
		name    string
		args    args
		want    date.Date
		wantErr bool
	}{
		{
			name: "およそ 1 週間前（同じ年）",
			args: args{
				now:  time.Date(2022, 6, 14, 0, 0, 0, 0, timeutil.LocationJST()),
				mmdd: "6/6",
			},
			want: date.New(2022, 6, 6),
		},
		{
			name: "およそ 1 週間前（昨年）",
			args: args{
				now:  time.Date(2022, 1, 3, 0, 0, 0, 0, timeutil.LocationJST()),
				mmdd: "12/27",
			},
			want: date.New(2021, 12, 27),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildDateFromMMDD(tt.args.now, tt.args.mmdd)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildDateFromMMDD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(time.Time(tt.want), time.Time(got)); diff != "" {
				t.Errorf("buildDateFromMMDD() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_onsenProgramToPrograms(t *testing.T) {
	type args struct {
		onsenPgram onsenProgram
	}
	tests := []struct {
		name    string
		args    args
		want    []program.Program
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				onsenPgram: onsenProgram{
					ID:    17,
					Title: "セブン-イレブン presents 佐倉としたい大西",
					Contents: []Content{
						{
							ID:           11134,
							Title:        "第334回",
							ProgramID:    17,
							OngenID:      11134,
							Premium:      false,
							Free:         true,
							DeliveryDate: "8/23",
							StreamingURL: "https://onsen.test/playlist.m3u8",
							Expiring:     false,
						},
						{
							ID:           11054,
							Title:        "第333回",
							ProgramID:    17,
							OngenID:      11054,
							Premium:      true,
							Free:         false,
							DeliveryDate: "8/16",
							StreamingURL: "",
							Expiring:     false,
						},
					},
				},
			},
			want: []program.Program{
				{
					ID:          11134,
					Station:     program.StationOnsen,
					Title:       "セブン-イレブン presents 佐倉としたい大西",
					Episode:     "第334回",
					Start:       time.Date(2022, 8, 23, 0, 0, 0, 0, timeutil.LocationJST()),
					End:         time.Time{},
					Status:      program.StatusScheduled,
					StreamType:  program.StreamTypeOndemand,
					PlaylistURL: "https://onsen.test/playlist.m3u8",
				},
				{
					ID:          11054,
					Station:     program.StationOnsen,
					Title:       "セブン-イレブン presents 佐倉としたい大西",
					Episode:     "第333回",
					Start:       time.Date(2022, 8, 16, 0, 0, 0, 0, timeutil.LocationJST()),
					End:         time.Time{},
					Status:      program.StatusScheduled,
					StreamType:  program.StreamTypeOndemand,
					PlaylistURL: "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := onsenProgramToPrograms(tt.args.onsenPgram)
			if (err != nil) != tt.wantErr {
				t.Errorf("onsenProgramToPrograms() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreFields(program.Program{}, "UUID")); diff != "" {
				t.Errorf("onsenProgramToPrograms() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_client_GetPrograms(t *testing.T) {
	tests := []struct {
		name    string
		wantLen int
		wantErr bool
	}{
		{
			name: "ok",
			// 雑に
			// $ jq '.[].contents | length' programs.json | paste -sd+ | bc
			// 2477
			wantLen: 2477,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := recorder.New(fmt.Sprintf("../../testdata/infrastructures/onsen/go-vcr/%s", tt.name))
			if err != nil {
				t.Fatal(err)
			}
			defer r.Stop()

			c := &client{
				httpClient: r.GetDefaultClient(),
			}
			got, err := c.GetPrograms(context.Background(), date.Date{})
			if (err != nil) != tt.wantErr {
				t.Errorf("client.GetPrograms() error = %+v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantLen {
				t.Errorf("client.GetPrograms() len(got) = %v, wantLen %v", len(got), tt.wantLen)
			}
		})
	}
}
