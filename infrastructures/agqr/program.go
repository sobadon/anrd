package agqr

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/sobadon/anrd/domain/model/date"
	"github.com/sobadon/anrd/domain/model/program"
	"github.com/sobadon/anrd/internal/errutil"
	"github.com/sobadon/anrd/internal/timeutil"
)

type agqrProgram struct {
	// "514530"
	ScheduleProgramID string `json:"schedule_program_id"`

	// "2022-08-02"
	ScheduleDate string `json:"schedule_date"`

	// "1791"
	ProgramID string `json:"program_id"`

	// 5:00, 24:00
	ProgramStartTime string `json:"program_start_time"`

	// 5, 24
	ProgramStartTimeHour string `json:"program_start_time_hour"`

	// 00, 30
	ProgramStartTimeMinute string `json:"program_start_time_minute"`

	// 6:00, 24:30
	ProgramEndTime string `json:"program_end_time"`

	// 6, 24
	ProgramEndTimeHour string `json:"program_end_time_hour"`

	// 00, 30
	ProgramEndTimeMinute string `json:"program_end_time_minute"`

	// アニメ・声優・ゲーム業界のイベントの司会を多数担当するミュージシャンの鷲崎健による月曜日～木曜日　２４時～２５時の生放送。火曜日はアイドルの沢口けいこが登場！
	ProgramInformation string `json:"program_information"`

	// 鷲崎健のヨルナイト×ヨルナイト
	ProgramTitle string `json:"program_title"`

	// 鷲崎健, 沢口けいこ
	// カンマ（,）+ スペース区切り？
	ProgramPersonality string `json:"program_personality"`
}

func (c *client) GetPrograms(ctx context.Context, date date.Date) ([]program.Program, error) {
	programURL := buildURL(c.programBaseURL, date)
	log.Ctx(ctx).Debug().Msgf("http get target url: %s", programURL.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, programURL.String(), nil)
	if err != nil {
		return nil, errors.Wrap(errutil.ErrInternal, err.Error())
	}

	log.Ctx(ctx).Debug().Msgf("fetch program .... (day = %s)", time.Time(date).Format("2006-01-02"))
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(errutil.ErrHTTPRequest, err.Error())
	}

	if res.StatusCode != 200 {
		return nil, errors.Wrapf(errutil.ErrGetProgramNotOK, "http status code is %d", res.StatusCode)
	}

	defer res.Body.Close()
	agqrPgrams, err := decodeToAgqrProgram(res.Body)
	if err != nil {
		return nil, err
	}

	var pgrams []program.Program
	for _, agqrPgram := range agqrPgrams {
		pgram, err := agqrProgramToProgram(agqrPgram)
		if err != nil {
			return nil, err
		}
		pgrams = append(pgrams, pgram)
	}

	log.Ctx(ctx).Info().Msgf("successfully fetched program (day = %s)", time.Time(date).Format("2006-01-02"))
	log.Ctx(ctx).Debug().Msgf("fetched program len: %d", len(pgrams))
	return pgrams, nil
}

func buildURL(baseURL *url.URL, date date.Date) *url.URL {
	dateStr := time.Time(date).Format("2006-01-02")
	queries := baseURL.Query()
	queries.Set("date", dateStr)
	baseURL.RawQuery = queries.Encode()
	return baseURL
}

func decodeToAgqrProgram(input io.Reader) ([]agqrProgram, error) {
	var agqrPgrams []agqrProgram
	decoder := json.NewDecoder(input)
	err := decoder.Decode(&agqrPgrams)
	if err != nil {
		return nil, errors.Wrap(errutil.ErrJSONDecode, err.Error())
	}
	return agqrPgrams, nil
}

func agqrProgramToProgram(agqrPgram agqrProgram) (program.Program, error) {
	year, err := strconv.Atoi(agqrPgram.ScheduleDate[0:4])
	if err != nil {
		return program.Program{}, errors.Wrap(errutil.ErrInternal, err.Error())
	}

	month, err := strconv.Atoi(agqrPgram.ScheduleDate[5:7])
	if err != nil {
		return program.Program{}, errors.Wrap(errutil.ErrInternal, err.Error())
	}

	day, err := strconv.Atoi(agqrPgram.ScheduleDate[8:10])
	if err != nil {
		return program.Program{}, errors.Wrap(errutil.ErrInternal, err.Error())
	}

	programDate := date.New(year, time.Month(month), day)

	programStartTimeHour, err := strconv.Atoi(agqrPgram.ProgramStartTimeHour)
	if err != nil {
		return program.Program{}, errors.Wrap(errutil.ErrInternal, err.Error())
	}
	programStartTimeMinute, err := strconv.Atoi(agqrPgram.ProgramStartTimeMinute)
	if err != nil {
		return program.Program{}, errors.Wrap(errutil.ErrInternal, err.Error())
	}
	start, err := converToTime(programDate, programStartTimeHour, programStartTimeMinute)
	if err != nil {
		return program.Program{}, err
	}

	programEndTimeHour, err := strconv.Atoi(agqrPgram.ProgramEndTimeHour)
	if err != nil {
		return program.Program{}, errors.Wrap(errutil.ErrInternal, err.Error())
	}
	programEndTimeMinute, err := strconv.Atoi(agqrPgram.ProgramEndTimeMinute)
	if err != nil {
		return program.Program{}, errors.Wrap(errutil.ErrInternal, err.Error())
	}
	end, err := converToTime(programDate, programEndTimeHour, programEndTimeMinute)
	if err != nil {
		return program.Program{}, err
	}

	id, err := strconv.Atoi(agqrPgram.ScheduleProgramID)
	if err != nil {
		return program.Program{}, errors.Wrap(errutil.ErrInternal, err.Error())
	}

	pgram := program.Program{
		UUID:        uuid.NewString(),
		ID:          id,
		Station:     program.StationAgqr,
		Title:       agqrPgram.ProgramTitle,
		Episode:     "",
		Start:       start,
		End:         end,
		Status:      program.StatusScheduled,
		StreamType:  program.StreamTypeBroadcast,
		PlaylistURL: "",
		// TODO: Personality すぐ必要なわけではないならいいや
	}
	return pgram, nil
}

// 24 時を超える表記もひっくるめて time.Time にパースする
func converToTime(date date.Date, hour int, minute int) (time.Time, error) {
	year, err := strconv.Atoi(time.Time(date).Format("2006"))
	if err != nil {
		return time.Time{}, errors.Wrap(errutil.ErrInternal, err.Error())
	}

	month, err := strconv.Atoi(time.Time(date).Format("01"))
	if err != nil {
		return time.Time{}, errors.Wrap(errutil.ErrInternal, err.Error())
	}

	day, err := strconv.Atoi(time.Time(date).Format("02"))
	if err != nil {
		return time.Time{}, errors.Wrap(errutil.ErrInternal, err.Error())
	}

	// time.Parse では 24, 25, ... 時を扱えないが、time.Date だとそれらを扱える
	retTime := time.Date(year, time.Month(month), day, hour, minute, 0, 0, timeutil.LocationJST())
	if retTime.IsZero() {
		return time.Time{}, errors.Wrap(errutil.ErrTimeParse, "time is zero")
	}

	return retTime, nil
}
