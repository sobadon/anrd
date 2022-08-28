package agqr

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sobadon/anrd/domain/model/date"
	"github.com/sobadon/anrd/domain/model/program"
	"github.com/sobadon/anrd/internal/timeutil"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

func Test_conver24Hour(t *testing.T) {
	type args struct {
		date   date.Date
		hour   int
		minute int
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "普通の午前時間帯を扱える",
			args: args{
				date: date.New(2022, 4, 1), hour: 9, minute: 0,
			},
			want: time.Date(2022, 4, 1, 9, 0, 0, 0, timeutil.LocationJST()),
		},
		{
			name: "普通の午後時間帯を扱える",
			args: args{
				date: date.New(2022, 4, 1), hour: 21, minute: 30,
			},
			want: time.Date(2022, 4, 1, 21, 30, 0, 0, timeutil.LocationJST()),
		},
		{
			name: "24 時ぴったりは翌日の 0 時として扱える",
			args: args{
				date: date.New(2022, 4, 1), hour: 24, minute: 0,
			},
			want: time.Date(2022, 4, 2, 0, 0, 0, 0, timeutil.LocationJST()),
		},
		{
			name: "25 時を扱える",
			args: args{
				date: date.New(2022, 4, 1), hour: 25, minute: 30,
			},
			want: time.Date(2022, 4, 2, 1, 30, 0, 0, timeutil.LocationJST()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := converToTime(tt.args.date, tt.args.hour, tt.args.minute)
			if (err != nil) != tt.wantErr {
				t.Errorf("converToTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("converToTime() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_client_GetPrograms(t *testing.T) {
	type args struct {
		date date.Date
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
	}{
		{
			name: "正常に番組表を取得できる",
			args: args{
				date: date.New(2022, 8, 2),
			},
			wantLen: 200,
			wantErr: false,
		},
		{
			name: "未来すぎる番組表を取得したらエラーにならず len は 0",
			args: args{
				date: date.New(2030, 1, 1),
			},
			wantLen: 0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL, err := url.Parse("https://www.joqr.co.jp/rss/program/json.php?type=ag")
			if err != nil {
				t.Fatal(err)
			}

			rec, err := recorder.New(fmt.Sprintf("../../testdata/infrastructure/agqr/GetPrograms/%s", tt.name))
			if err != nil {
				t.Fatal(err)
			}
			defer rec.Stop()

			rec.SetReplayableInteractions(true)

			c := &client{
				httpClient:     rec.GetDefaultClient(),
				programBaseURL: baseURL,
			}
			got, err := c.GetPrograms(context.Background(), tt.args.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.GetPrograms() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("client.GetPrograms() programLen = %v, wantLen %v", len(got), tt.wantLen)
				return
			}
		})
	}
}

func Test_buildURL(t *testing.T) {
	const baseURL = "https://www.joqr.co.jp/rss/program/json.php?type=ag"
	type args struct {
		baseURL string
		date    date.Date
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "1 ケタ・1 ケタ（0 埋め）",
			args: args{
				baseURL: baseURL,
				date:    date.New(2022, 8, 1),
			},
			want: "https://www.joqr.co.jp/rss/program/json.php?date=2022-08-01&type=ag",
		},
		{
			name: "2 ケタ・2 ケタ",
			args: args{
				baseURL: baseURL,
				date:    date.New(2022, 12, 31),
			},
			want: "https://www.joqr.co.jp/rss/program/json.php?date=2022-12-31&type=ag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL, err := url.Parse(tt.args.baseURL)
			if err != nil {
				t.Fatal(err)
			}

			if got := buildURL(baseURL, tt.args.date); got.String() != tt.want {
				t.Errorf("buildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_agqrProgramToProgram(t *testing.T) {
	type args struct {
		agqrPgram agqrProgram
	}
	tests := []struct {
		name    string
		args    args
		want    program.Program
		wantErr bool
	}{
		{
			name: "Start, End がどちらとも X < 24（日中）",
			args: args{
				agqrPgram: agqrProgram{
					ScheduleProgramID:      "514579",
					ScheduleDate:           "2022-08-03",
					ProgramID:              "1920",
					ProgramStartTime:       "11:30",
					ProgramStartTimeHour:   "11",
					ProgramStartTimeMinute: "30",
					ProgramEndTime:         "12:00",
					ProgramEndTimeHour:     "12",
					ProgramEndTimeMinute:   "0",
					ProgramInformation:     "この番組は、文化放送と＜音泉＞が一緒にラジオを「したい」。佐倉さんが大西さんとラジオが「したい」。大西さんも佐倉さんと繋がりたい！と始まった今までにない実験的なプロジェクトです！！文化放送では超A&amp;G＋にて毎週火曜23時半より動画付きで放送中！＜音泉＞では毎週火曜24時よりおまけコーナーをつけてアーカイブで配信中！",
					ProgramTitle:           "セブン-イレブンpresents 佐倉としたい大西",
					ProgramPersonality:     "佐倉綾音, 大西沙織",
				},
			},
			want: program.Program{
				// UUID: ***
				ID:          514579,
				Station:     program.StationAgqr,
				Title:       "セブン-イレブンpresents 佐倉としたい大西",
				Episode:     "",
				Start:       time.Date(2022, 8, 3, 11, 30, 0, 0, timeutil.LocationJST()),
				End:         time.Date(2022, 8, 3, 12, 0, 0, 0, timeutil.LocationJST()),
				Status:      program.StatusScheduled,
				StreamType:  program.StreamTypeBroadcast,
				PlaylistURL: "",
			},
		},
		{
			name: "Start, End がどちらとも X >= 24（深夜）",
			args: args{
				agqrPgram: agqrProgram{
					ScheduleProgramID:      "514569",
					ScheduleDate:           "2022-08-03",
					ProgramID:              "1791",
					ProgramStartTime:       "24:00",
					ProgramStartTimeHour:   "24",
					ProgramStartTimeMinute: "0",
					ProgramEndTime:         "24:30",
					ProgramEndTimeHour:     "24",
					ProgramEndTimeMinute:   "30",
					ProgramInformation:     "アニメ・声優・ゲーム業界のイベントの司会を多数担当するミュージシャンの鷲崎健による月曜日～木曜日　２４時～２５時の生放送。水曜日はミュージシャンの青木佑磨が登場！",
					ProgramTitle:           "鷲崎健のヨルナイト×ヨルナイト",
					ProgramPersonality:     "鷲崎健, 青木佑磨",
				},
			},
			want: program.Program{
				// UUID: ***
				ID:          514569,
				Station:     program.StationAgqr,
				Title:       "鷲崎健のヨルナイト×ヨルナイト",
				Episode:     "",
				Start:       time.Date(2022, 8, 4, 0, 0, 0, 0, timeutil.LocationJST()),
				End:         time.Date(2022, 8, 4, 0, 30, 0, 0, timeutil.LocationJST()),
				Status:      program.StatusScheduled,
				StreamType:  program.StreamTypeBroadcast,
				PlaylistURL: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := agqrProgramToProgram(tt.args.agqrPgram)
			if (err != nil) != tt.wantErr {
				t.Errorf("agqrProgramToProgram() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreFields(program.Program{}, "UUID")); diff != "" {
				t.Errorf("agqrProgramToProgram() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
