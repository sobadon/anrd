package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/sobadon/anrd/domain/model/program"
	"github.com/sobadon/anrd/domain/repository"
	"github.com/sobadon/anrd/internal/errutil"
)

type programSqlite struct {
	UUID        string         `db:"uuid"`
	ID          int            `db:"id"`
	Station     string         `db:"station"`
	Title       string         `db:"title"`
	Episode     sql.NullString `db:"episode"`
	Start       time.Time      `db:"start"`
	End         time.Time      `db:"end"`
	Status      string         `db:"status"`
	StreamType  string         `db:"stream_type"`
	PlaylistURL sql.NullString `db:"playlist_url"`
}

func programSqliteToModelProgram(pgramSqlite programSqlite) program.Program {
	return program.Program{
		UUID:        pgramSqlite.UUID,
		ID:          pgramSqlite.ID,
		Station:     program.Station(pgramSqlite.Station),
		Title:       pgramSqlite.Title,
		Episode:     pgramSqlite.Episode.String, // 空文字になってくれればよい
		Start:       pgramSqlite.Start,
		End:         pgramSqlite.End,
		Status:      program.Status(pgramSqlite.Status),
		StreamType:  program.StreamType(pgramSqlite.StreamType),
		PlaylistURL: pgramSqlite.PlaylistURL.String,
	}
}

func modelProgramToProgramSqlite(pgram program.Program) programSqlite {
	var episode sql.NullString
	if pgram.Episode == "" {
		episode.Valid = false
	} else {
		episode.Valid = true
		episode.String = pgram.Episode
	}

	var playlistURL sql.NullString
	if pgram.PlaylistURL == "" {
		playlistURL.Valid = false
	} else {
		playlistURL.Valid = true
		playlistURL.String = pgram.PlaylistURL
	}

	return programSqlite{
		UUID:        pgram.UUID,
		ID:          pgram.ID,
		Station:     pgram.Station.String(),
		Episode:     episode,
		Title:       pgram.Title,
		Start:       pgram.Start,
		End:         pgram.End,
		Status:      pgram.Status.String(),
		StreamType:  pgram.StreamType.String(),
		PlaylistURL: playlistURL,
	}
}

func NewDB(dbPath string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return nil, errors.Wrap(errutil.ErrDatabaseOpen, err.Error())
	}
	return db, nil
}

// テーブル作成
func Setup(db *sqlx.DB) error {
	_, err := db.Exec(`create table if not exists programs (
		uuid text primary key,
		id integer not null,
		station text not null,
		title text not null,
		episode text,
		start timestamp not null,
		end timestamp not null,
		status text not null,
		stream_type text not null,
		playlist_url text,
		created_at timestamp not null default (datetime('now', 'localtime')),
		updated_at timestamp not null default (datetime('now', 'localtime')),
		unique (station, id)
	);`)
	if err != nil {
		return errors.Wrap(errutil.ErrDatabaseQuery, err.Error())
	}

	_, err = db.Exec(`CREATE TRIGGER if not exists trigger_updated_at AFTER UPDATE ON programs
		BEGIN
			UPDATE programs SET updated_at = DATETIME('now', 'localtime') WHERE rowid == NEW.rowid;
		END;
		`)
	if err != nil {
		return errors.Wrap(errutil.ErrDatabaseQuery, err.Error())
	}

	return nil
}

type client struct {
	DB *sqlx.DB
}

func New(db *sqlx.DB) repository.ProgramPersistence {
	return &client{
		DB: db,
	}
}

func (c *client) Save(ctx context.Context, pgram program.Program) error {
	rows, err := c.DB.QueryxContext(ctx, "select count(*) from programs where id = ?", pgram.ID)
	if err != nil {
		return errors.Wrap(errutil.ErrDatabaseQuery, err.Error())
	}

	var lineCount int
	for rows.Next() {
		err := rows.Scan(&lineCount)
		if err != nil {
			return errors.Wrap(errutil.ErrDatabaseScan, err.Error())
		}
	}

	// 既に番組情報が登録されていれば追加しない
	// TODO: 番組表の変更に対応できない問題がある
	if lineCount != 0 {
		return nil
	}

	pgramSqlite := modelProgramToProgramSqlite(pgram)
	_, err = c.DB.NamedExecContext(ctx,
		`insert into programs (uuid, id, station, title, episode, start, end, status, stream_type, playlist_url)
		values
		(:uuid, :id, :station, :title, :episode, :start, :end, :status, :stream_type, :playlist_url)`,
		pgramSqlite)
	if err != nil {
		return errors.Wrap(errutil.ErrDatabaseQuery, err.Error())
	}

	return nil
}

/*
broadcast のときに使う 後で実装
func (c *client) LoadStartIn(ctx context.Context, now time.Time, duration time.Duration) ([]program.Program, error) {
	afterAbsoluteTime := now.Add(duration)

	stmt, err := c.DB.PrepareNamedContext(ctx, `select id, title, start, end, status from programs where status = 'scheduled' and :now < start and start < :after`)
	if err != nil {
		return nil, errors.Wrap(errutil.ErrDatabasePrepare, err.Error())
	}

	var pgramsSqlite []programSqlite
	err = stmt.SelectContext(ctx, &pgramsSqlite, map[string]interface{}{"now": now, "after": afterAbsoluteTime})
	if err != nil {
		return nil, errors.Wrap(errutil.ErrDatabaseQuery, err.Error())
	}

	var pgrams []program.Program
	for _, pgramSqlite := range pgramsSqlite {
		pgram := programSqliteToModelProgram(pgramSqlite)
		pgrams = append(pgrams, pgram)
	}

	return pgrams, nil
}
*/

func (c *client) ChangeStatus(ctx context.Context, pgram program.Program, newStatus program.Status) error {
	var oldStatus string
	err := c.DB.GetContext(ctx, &oldStatus, `select status from programs where uuid = ?`, pgram.UUID)
	if err != nil {
		return errors.Wrap(errutil.ErrDatabaseQuery, err.Error())
	}

	_, err = c.DB.NamedExecContext(ctx, `update programs set status = :status where uuid = :uuid`, map[string]interface{}{"status": newStatus, "uuid": pgram.UUID})
	if err != nil {
		return errors.Wrap(errutil.ErrDatabaseQuery, err.Error())
	}

	return nil
}

// 返されるエラー
// - errutil.ErrDatabaseNotFoundProgram
func (c *client) LoadOndemandScheduled(ctx context.Context) (*program.Program, error) {
	var pgramsSqlite []programSqlite
	err := c.DB.SelectContext(ctx, &pgramsSqlite, `select uuid, id, station, title, episode, start, end, status, stream_type, playlist_url from programs where status = 'scheduled' limit 1`)
	if err != nil {
		return nil, errors.Wrap(errutil.ErrDatabaseQuery, err.Error())
	}

	if len(pgramsSqlite) == 0 {
		return nil, errors.Wrap(errutil.ErrDatabaseNotFoundProgram, "not found program (scheduled ondemand)")
	}

	pgram := programSqliteToModelProgram(pgramsSqlite[0])
	return &pgram, nil
}
