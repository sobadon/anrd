package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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
	tests := []struct {
		name    string
		prepare func(db *sqlx.DB) error
		want    *program.Program
		wantErr error
	}{
		{
			name:    "番組が存在しなければ ErrDatabaseNotFoundProgram を返す",
			prepare: func(db *sqlx.DB) error { return nil },
			want:    nil,
			wantErr: errutil.ErrDatabaseNotFoundProgram,
		},
		{
			name: "正常に番組を返す",
			prepare: func(db *sqlx.DB) error {
				_, err := db.Exec(`insert into programs (uuid, id, station, title, episode, start, end, status, stream_type, playlist_url) values
					("4ba3b9ff-5e0b-44ae-a99d-6dfb27deac0e", "11133", "onsen", "セブン-イレブン presents 佐倉としたい大西", "第333回", "2022-08-22 00:00:00+09:00", "0001-01-01 00:00:00+00:00", "done", "ondemand", "https://onsen.test/playlist.m3u8"),
					("89350da4-7f3b-4438-b99f-41ae9aa52bf5", "11134", "onsen", "セブン-イレブン presents 佐倉としたい大西", "第334回", "2022-08-23 00:00:00+09:00", "0001-01-01 00:00:00+00:00", "scheduled", "ondemand", "https://onsen.test/playlist.m3u8")
`)
				return err
			},
			want: &program.Program{
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
			wantErr: nil,
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
			got, err := c.LoadOndemandScheduled(context.Background())
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
