package program

import (
	"time"

	"github.com/sobadon/anrd/domain/model/date"
)

type Program struct {
	// ユニークであろう ID
	ID int

	// 番組タイトル
	Title string

	// 第N回
	// StreamType が ondemand のときに存在する
	Episode string

	// 番組の開始日時
	Start time.Time

	// 番組の終了日時
	// StreamType が ondemand のときは invalid が time.Time が存在する
	End time.Time

	Status     Status
	StreamType StreamType

	// StreamType が ondemand のとき PlaylistURL に m3u8 の URL が存在する
	PlaylistURL string

	// すぐは必要ないならあとで
	// Personality []string
}

/*
func Dummies(now time.Time) []Program {
	return []Program{
		{
			ID:    1,
			Title: "ダミー1",
			Start: now.Add(20 * time.Second),
			End:   now.Add(20 * time.Second).Add(1 * time.Minute),
		},
		// {
		// 	ID: 2,
		// 	Title: "ダミー2",
		// 	Start: now.Add(20 * time.Second),
		// 	End: now.Add(20 * time.Second).Add(1 * time.Minute),
		// },
	}
}
*/

func NewProgramOndemand(id int, title string, episode string, date date.Date, playlistURL string) Program {
	return Program{
		ID:          id,
		Title:       title,
		Episode:     episode,
		Start:       time.Time(date),
		Status:      StatusScheduled,
		StreamType:  StreamTypeOndemand,
		PlaylistURL: playlistURL,
	}
}
