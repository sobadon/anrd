package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sobadon/anrd/domain/model/program"
	"github.com/sobadon/anrd/internal/errutil"
	"github.com/sobadon/anrd/internal/testutil"
	"github.com/sobadon/anrd/internal/timeutil"
)

func tempFilename(t testing.TB) string {
	f, err := os.CreateTemp("", "anrd-")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestSetup(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "エラーなしで終了する",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFilename := tempFilename(t)
			defer os.Remove(tempFilename)
			db, err := sqlx.Open("sqlite3", tempFilename)
			if err != nil {
				t.Fatal(err)
			}

			if err := Setup(db); (err != nil) != tt.wantErr {
				t.Errorf("Setup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_client_Save(t *testing.T) {
	type args struct {
		pgram program.Program
	}
	tests := []struct {
		name    string
		prepare func(db *sqlx.DB) error
		args    args
		// エラーなく終了すればよいものとする
		wantErr bool
	}{
		/*
			今は broadcast は対応していない
			{
				name:    "ok: StreamType = Broadcast",
				prepare: func(db *sqlx.DB) error { return nil },
				args: args{
					pgram: program.Program{
						UUID: uuid.NewString(),
						ID:          514530,
						Title:       "鷲崎健のヨルナイト×ヨルナイト",
						Episode:     "",
						Start:       time.Date(2022, 8, 4, 0, 0, 0, 0, timeutil.LocationJST()),
						End:         time.Date(2022, 8, 4, 0, 30, 0, 0, timeutil.LocationJST()),
						Status:      program.StatusScheduled,
						StreamType:  program.StreamTypeBroadcast,
						PlaylistURL: "",
					},
				},
				wantErr: false,
			},
		*/

		{
			name:    "ok: 空っぽ - StreamType = Ondemand",
			prepare: func(db *sqlx.DB) error { return nil },
			args: args{
				pgram: program.Program{
					UUID:        uuid.NewString(),
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
			},
			wantErr: false,
		},
		{
			// 本来はアップデート処理を実装したいところ
			name: "ok: 既に存在していてもエラーにならない - StreamType = Ondemand",
			prepare: func(db *sqlx.DB) error {
				_, err := db.Exec(`insert into programs (uuid, id, station, title, episode, start, end, status, stream_type, playlist_url) values (
					"89350da4-7f3b-4438-b99f-41ae9aa52bf5", "11134", "onsen", "セブン-イレブン presents 佐倉としたい大西", "第334回", "2022-08-23 00:00:00+09:00", "0001-01-01 00:00:00+00:00", "scheduled", "ondemand", "https://onsen.test/playlist.m3u8"
				)`)
				return err
			},
			args: args{
				pgram: program.Program{
					UUID:        uuid.NewString(),
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
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFilename := tempFilename(t)
			defer os.Remove(tempFilename)
			db, err := sqlx.Open("sqlite3", tempFilename)
			if err != nil {
				t.Fatal(err)
			}

			err = Setup(db)
			if err != nil {
				t.Fatal(err)
			}

			err = tt.prepare(db)
			if err != nil {
				t.Fatal(err)
			}

			c := &client{
				DB: db,
			}
			if err := c.Save(context.Background(), tt.args.pgram); (err != nil) != tt.wantErr {
				t.Errorf("client.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_client_LoadBroadcastStartIn(t *testing.T) {
	type args struct {
		now      time.Time
		duration time.Duration
	}
	tests := []struct {
		name    string
		prepare func(db *sqlx.DB) error
		args    args
		want    []program.Program
		wantErr error
	}{
		{
			name: "番組 1 つ取得できる",
			prepare: func(db *sqlx.DB) error {
				_, err := db.Exec(`insert into programs (uuid, id, station, title, episode, start, end, status, stream_type, playlist_url) values
					("3824be0b-5103-4976-8d4f-212c53ac4999", "514529", "agqr", "テスト番組名", null, "2022-08-09 23:50:00+09:00", "2022-08-10 00:00:00+09:00", "recording", "broadcast", null),
					("b7750840-3407-44a0-b670-2b08cb8e0eb3", "514530", "agqr", "鷲崎健のヨルナイト×ヨルナイト", null, "2022-08-10 00:00:00+09:00", "2022-08-10 00:30:00+09:00", "scheduled", "broadcast", null)
				`)
				return err
			},
			args: args{
				now:      time.Date(2022, 8, 9, 23, 59, 30, 0, timeutil.LocationJST()),
				duration: 1 * time.Minute,
			},
			want: []program.Program{
				{
					ID:         514530,
					Station:    program.StationAgqr,
					Title:      "鷲崎健のヨルナイト×ヨルナイト",
					Start:      time.Date(2022, 8, 10, 0, 0, 0, 0, timeutil.LocationJST()),
					End:        time.Date(2022, 8, 10, 0, 30, 0, 0, timeutil.LocationJST()),
					Status:     program.StatusScheduled,
					StreamType: program.StreamTypeBroadcast,
				},
			},
			wantErr: nil,
		},
		{
			// agqr で 2, 3 分間の番組ってないかも
			name: "番組 2 つ取得できる",
			prepare: func(db *sqlx.DB) error {
				_, err := db.Exec(`insert into programs (uuid, id, station, title, episode, start, end, status, stream_type, playlist_url) values
				("7d4af7b8-666f-48b0-b756-10ce4e3131a6", "514528", "agqr", "テスト番組名1", null, "2022-08-09 23:50:00+09:00", "2022-08-10 23:59:00+09:00", "recording", "broadcast", null),
				("616b518b-063f-4ad0-b6de-a5c257886fe7", "514529", "agqr", "テスト番組名2", null, "2022-08-09 23:59:00+09:00", "2022-08-10 00:00:00+09:00", "scheduled", "broadcast", null),
				("ea705dfe-17d3-412c-af7a-04071c63efdb", "514530", "agqr", "鷲崎健のヨルナイト×ヨルナイト", null, "2022-08-10 00:00:00+09:00", "2022-08-10 00:30:00+09:00", "scheduled", "broadcast", null)
				`)
				return err
			},
			args: args{
				now:      time.Date(2022, 8, 9, 23, 58, 0, 0, timeutil.LocationJST()),
				duration: 5 * time.Minute,
			},
			want: []program.Program{
				{
					ID:         514529,
					Station:    program.StationAgqr,
					Title:      "テスト番組名2",
					Start:      time.Date(2022, 8, 9, 23, 59, 0, 0, timeutil.LocationJST()),
					End:        time.Date(2022, 8, 10, 0, 0, 0, 0, timeutil.LocationJST()),
					Status:     program.StatusScheduled,
					StreamType: program.StreamTypeBroadcast,
				},
				{
					ID:         514530,
					Station:    program.StationAgqr,
					Title:      "鷲崎健のヨルナイト×ヨルナイト",
					Start:      time.Date(2022, 8, 10, 0, 0, 0, 0, timeutil.LocationJST()),
					End:        time.Date(2022, 8, 10, 0, 30, 0, 0, timeutil.LocationJST()),
					Status:     program.StatusScheduled,
					StreamType: program.StreamTypeBroadcast,
				},
			},
			wantErr: nil,
		},
		{
			name: "該当番組がなければ nil を返す",
			prepare: func(db *sqlx.DB) error {
				_, err := db.Exec(`insert into programs (uuid, id, station, title, episode, start, end, status, stream_type, playlist_url) values
					("3824be0b-5103-4976-8d4f-212c53ac4999", "514529", "agqr", "テスト番組名", null, "2022-08-09 23:50:00+09:00", "2022-08-10 00:00:00+09:00", "done", "broadcast", null),
					("b7750840-3407-44a0-b670-2b08cb8e0eb3", "514530", "agqr", "鷲崎健のヨルナイト×ヨルナイト", null, "2022-08-10 00:00:00+09:00", "2022-08-10 00:30:00+09:00", "recording", "broadcast", null)
				`)
				return err
			},
			args: args{
				now:      time.Date(2022, 8, 10, 0, 10, 0, 0, timeutil.LocationJST()),
				duration: 5 * time.Minute,
			},
			want:    nil,
			wantErr: errutil.ErrDatabaseNotFoundProgram,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFilename := tempFilename(t)
			defer os.Remove(tempFilename)
			db, err := sqlx.Open("sqlite3", tempFilename)
			if err != nil {
				t.Fatal(err)
			}

			p := &client{
				DB: db,
			}

			err = Setup(p.DB)
			if err != nil {
				t.Fatal(err)
			}

			err = tt.prepare(p.DB)
			if err != nil {
				t.Fatal(err)
			}

			got, err := p.LoadBroadcastStartIn(context.Background(), tt.args.now, tt.args.duration)
			if !testutil.ErrorsAs(err, tt.wantErr) {
				t.Errorf("client.LoadBroadcastStartIn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreFields(program.Program{}, "UUID")); diff != "" {
				t.Errorf("client.LoadBroadcastStartIn() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_client_ChangeStatus(t *testing.T) {
	type args struct {
		pgram     program.Program
		newStatus program.Status
	}
	tests := []struct {
		name    string
		prepare func(db *sqlx.DB) error
		args    args
		wantErr bool
	}{
		{
			name: "正常に status を変更できる",
			prepare: func(db *sqlx.DB) error {
				_, err := db.Exec(`insert into programs (uuid, id, station, title, episode, start, end, status, stream_type, playlist_url) values (
					"89350da4-7f3b-4438-b99f-41ae9aa52bf5", "11134", "onsen", "セブン-イレブン presents 佐倉としたい大西", "第334回", "2022-08-23 00:00:00+09:00", "0001-01-01 00:00:00+00:00", "scheduled", "ondemand", "https://onsen.test/playlist.m3u8"
				)`)
				return err
			},
			args: args{
				pgram: program.Program{
					UUID:        "89350da4-7f3b-4438-b99f-41ae9aa52bf5", // prepare と同一
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
				newStatus: program.StatusRecording,
			},
			wantErr: false,
		},
		{
			name:    "存在しない program の status を変更しようとしたらエラー",
			prepare: func(db *sqlx.DB) error { return nil },
			args: args{
				// prepare されていない program
				pgram: program.Program{
					UUID:        "89350da4-7f3b-4438-b99f-41ae9aa52bf5",
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
				newStatus: program.StatusRecording,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFilename := tempFilename(t)
			defer os.Remove(tempFilename)
			db, err := sqlx.Open("sqlite3", tempFilename)
			if err != nil {
				t.Fatal(err)
			}

			err = Setup(db)
			if err != nil {
				t.Fatal(err)
			}

			err = tt.prepare(db)
			if err != nil {
				t.Fatal(err)
			}

			c := &client{
				DB: db,
			}
			if err := c.ChangeStatus(context.Background(), tt.args.pgram, tt.args.newStatus); (err != nil) != tt.wantErr {
				t.Errorf("client.ChangeStatus() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			var gotStatus program.Status
			err = c.DB.Get(&gotStatus, `select status from programs where uuid = ?`, tt.args.pgram.UUID)
			if err != nil {
				t.Fatal(err)
			}
			if gotStatus != tt.args.newStatus {
				t.Errorf("client.ChangeStatus() gotStatus = %v, wantStatus %v", gotStatus, tt.args.newStatus)
			}
		})
	}
}

func Test_client_LoadOndemandScheduled(t *testing.T) {
	type args struct {
		limit int
	}
	tests := []struct {
		name    string
		args    args
		prepare func(db *sqlx.DB) error
		want    *[]program.Program
		wantErr error
	}{
		{
			name: "番組が存在しなければ ErrDatabaseNotFoundProgram を返す",
			args: args{
				limit: 100,
			},
			prepare: func(db *sqlx.DB) error { return nil },
			want:    nil,
			wantErr: errutil.ErrDatabaseNotFoundProgram,
		},
		{
			name: "正常に番組を返す",
			args: args{
				limit: 100,
			},
			prepare: func(db *sqlx.DB) error {
				_, err := db.Exec(`insert into programs (uuid, id, station, title, episode, start, end, status, stream_type, playlist_url) values
					("e07df7c6-eae8-40f8-8922-6b7ef0497dc8", "11132", "onsen", "セブン-イレブン presents 佐倉としたい大西", "第332回", "2022-08-21 00:00:00+09:00", "0001-01-01 00:00:00+00:00", "scheduled", "ondemand", null),
					("4ba3b9ff-5e0b-44ae-a99d-6dfb27deac0e", "11133", "onsen", "セブン-イレブン presents 佐倉としたい大西", "第333回", "2022-08-22 00:00:00+09:00", "0001-01-01 00:00:00+00:00", "done", "ondemand", "https://onsen.test/playlist.m3u8"),
					("89350da4-7f3b-4438-b99f-41ae9aa52bf5", "11134", "onsen", "セブン-イレブン presents 佐倉としたい大西", "第334回", "2022-08-23 00:00:00+09:00", "0001-01-01 00:00:00+00:00", "scheduled", "ondemand", "https://onsen.test/playlist.m3u8")
`)
				return err
			},
			want: &[]program.Program{
				{
					UUID:        "89350da4-7f3b-4438-b99f-41ae9aa52bf5",
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
			},
			wantErr: nil,
		},
		{
			name: "limit にて取得件数を指定できる",
			args: args{
				limit: 1,
			},
			prepare: func(db *sqlx.DB) error {
				_, err := db.Exec(`insert into programs (uuid, id, station, title, episode, start, end, status, stream_type, playlist_url) values
					("e07df7c6-eae8-40f8-8922-6b7ef0497dc8", "11132", "onsen", "セブン-イレブン presents 佐倉としたい大西", "第332回", "2022-08-21 00:00:00+09:00", "0001-01-01 00:00:00+00:00", "scheduled", "ondemand", null),
					("4ba3b9ff-5e0b-44ae-a99d-6dfb27deac0e", "11133", "onsen", "セブン-イレブン presents 佐倉としたい大西", "第333回", "2022-08-22 00:00:00+09:00", "0001-01-01 00:00:00+00:00", "done", "ondemand", "https://onsen.test/playlist.m3u8"),
					("89350da4-7f3b-4438-b99f-41ae9aa52bf5", "11134", "onsen", "セブン-イレブン presents 佐倉としたい大西", "第334回", "2022-08-23 00:00:00+09:00", "0001-01-01 00:00:00+00:00", "scheduled", "ondemand", "https://onsen.test/playlist.m3u8"),
					("0dc3a01d-ddaa-4bc0-8714-3262e713940c", "11135", "onsen", "セブン-イレブン presents 佐倉としたい大西", "第335回", "2022-08-24 00:00:00+09:00", "0001-01-01 00:00:00+00:00", "scheduled", "ondemand", "https://onsen.test/playlist.m3u8")
`)
				return err
			},
			want: &[]program.Program{
				{
					UUID:        "89350da4-7f3b-4438-b99f-41ae9aa52bf5",
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
			},
			wantErr: nil,
		},
		// 複数番組取得のテスト
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFilename := tempFilename(t)
			defer os.Remove(tempFilename)
			db, err := sqlx.Open("sqlite3", tempFilename)
			if err != nil {
				t.Fatal(err)
			}

			err = Setup(db)
			if err != nil {
				t.Fatal(err)
			}

			err = tt.prepare(db)
			if err != nil {
				t.Fatal(err)
			}

			c := &client{
				DB: db,
			}
			got, err := c.LoadOndemandScheduled(context.Background(), tt.args.limit)
			if !testutil.ErrorsAs(err, tt.wantErr) {
				t.Errorf("client.LoadOndemandScheduled() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("client.LoadOndemandScheduled() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
