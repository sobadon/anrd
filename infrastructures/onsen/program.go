package onsen

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sobadon/anrd/domain/model/date"
	"github.com/sobadon/anrd/domain/model/program"
	"github.com/sobadon/anrd/internal/errutil"
	"github.com/sobadon/anrd/internal/timeutil"
)

type onsenProgram struct {
	// 17
	ID int `json:"id"`

	// DirectoryName     string `json:"directory_name"`
	// Display           bool   `json:"display"`
	// ShowContentsCount int    `json:"show_contents_count"`
	// BrandNew          bool   `json:"brand_new"`
	// BrandNewSp        bool   `json:"brand_new_sp"`

	// セブン-イレブン presents 佐倉としたい大西
	Title string `json:"title"`

	// Image             struct {
	// 	URL string `json:"url"`
	// } `json:"image"`

	// New               bool     `json:"new"`
	// List              bool     `json:"list"`
	// DeliveryInterval  string   `json:"delivery_interval"`
	// DeliveryDayOfWeek []int    `json:"delivery_day_of_week"`
	// CategoryList      []string `json:"category_list"`
	// Copyright         string   `json:"copyright"`
	// SponsorName       string   `json:"sponsor_name"`
	// Updated           string   `json:"updated"`
	// Performers        []struct {
	// 	ID        int    `json:"id"`
	// 	Name      string `json:"name"`
	// 	AllowLike bool   `json:"allow_like"`
	// } `json:"performers"`
	// RelatedLinks []struct {
	// 	LinkURL string `json:"link_url"`
	// 	Image   string `json:"image"`
	// } `json:"related_links"`
	// RelatedInfos []struct {
	// 	Category string `json:"category"`
	// 	LinkURL  string `json:"link_url"`
	// 	Caption  string `json:"caption"`
	// 	Image    string `json:"image"`
	// } `json:"related_infos"`
	// RelatedPrograms []struct {
	// 	Title         string `json:"title"`
	// 	DirectoryName string `json:"directory_name"`
	// 	Category      string `json:"category"`
	// 	Image         string `json:"image"`
	// 	Performers    []struct {
	// 		Name      string `json:"name"`
	// 		ID        int    `json:"id"`
	// 		AllowLike bool   `json:"allow_like"`
	// 	} `json:"performers"`
	// } `json:"related_programs"`
	// GuestInNewContent []struct {
	// 	ID        int    `json:"id"`
	// 	Name      string `json:"name"`
	// 	AllowLike bool   `json:"allow_like"`
	// } `json:"guest_in_new_content"`
	// Guests []struct {
	// 	ID        int    `json:"id"`
	// 	Name      string `json:"name"`
	// 	AllowLike bool   `json:"allow_like"`
	// } `json:"guests"`

	Contents []Content `json:"contents"`
}

type Content struct {
	// 11134
	ID int `json:"id"`

	// 第334回
	Title string `json:"title"`

	// Latest         bool   `json:"latest"`
	// MediaType      string `json:"media_type"`

	// 17
	// .id と同じ？
	ProgramID int `json:"program_id"`

	// New            bool   `json:"new"`
	// Event          bool   `json:"event"`
	// Block          bool   `json:"block"`

	// 11134
	// .contents[].program_id との違いわからず
	OngenID int `json:"ongen_id"`

	// プレミアムであるか否か
	// いわゆる最新回のみ false で、過去回は true になっている
	Premium bool `json:"premium"`

	// !Premium かな？
	Free bool `json:"free"`

	// 8/23
	// ゼロ埋めされない
	// null になることもあり、その場合は Web サイト上では「注目」という意味になる
	DeliveryDate string `json:"delivery_date"`

	// Movie          bool   `json:"movie"`
	// PosterImageURL string `json:"poster_image_url"`

	// m3u8 URL
	// https://onsen-ma3phlsvod.sslcs.cdngc.net/onsen-ma3pvod/_definst_/202208/toshitai220823xv3m-334.mp4/playlist.m3u8
	// premium: false だと m3u8 URL になる？
	// premium: true だと null になる？
	StreamingURL string `json:"streaming_url"`

	// TagImage       struct {
	// 	URL interface{} `json:"url"`
	// } `json:"tag_image"`
	// Guests []struct {
	// 	ID        int    `json:"id"`
	// 	Name      string `json:"name"`
	// 	AllowLike bool   `json:"allow_like"`
	// } `json:"guests"`

	// true を観測できず
	Expiring bool `json:"expiring"`
}

func (c *client) GetPrograms(ctx context.Context, _ date.Date) ([]program.Program, error) {
	const programURL = "https://www.onsen.ag/web_api/programs"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, programURL, nil)
	if err != nil {
		return nil, errors.Wrap(errutil.ErrInternal, err.Error())
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(errutil.ErrHTTPRequest, err.Error())
	}

	if res.StatusCode != 200 {
		return nil, errors.Wrapf(errutil.ErrGetProgramNotOK, "http status code is %d", res.StatusCode)
	}

	defer res.Body.Close()
	onsenPgrams, err := decodeToOnsenPrograms(res.Body)
	if err != nil {
		return nil, err
	}

	var pgrams []program.Program
	for _, onsenPgram := range onsenPgrams {
		pgrams_, err := onsenProgramToPrograms(onsenPgram)
		if err != nil {
			return nil, err
		}
		pgrams = append(pgrams, pgrams_...)
	}

	return pgrams, nil
}

func decodeToOnsenPrograms(input io.Reader) ([]onsenProgram, error) {
	var onsenPgrams []onsenProgram
	decoder := json.NewDecoder(input)
	err := decoder.Decode(&onsenPgrams)
	if err != nil {
		return nil, errors.Wrap(errutil.ErrJSONDecode, err.Error())
	}
	return onsenPgrams, nil
}

// onsenProgram.Contents -> []program.Program
func onsenProgramToPrograms(onsenPgram onsenProgram) ([]program.Program, error) {
	now := time.Now().In(timeutil.LocationJST())
	var pgrams []program.Program
	for _, content := range onsenPgram.Contents {
		// DeliveryDate が「注目」の意味を表わす空文字（null）になることがある
		// 空文字であるならば、とりあえず番組表取得日としてしまう
		var contentDate date.Date
		if content.DeliveryDate == "" {
			contentDate = date.NewFromToday(now)
		} else {
			var err error
			contentDate, err = buildDateFromMMDD(now, content.DeliveryDate)
			if err != nil {
				return nil, err
			}
		}

		pgram := program.NewProgramOndemand(content.ID, program.StationOnsen, onsenPgram.Title, content.Title, contentDate, content.StreamingURL)
		pgrams = append(pgrams, pgram)
	}
	return pgrams, nil
}

// 現在日時と MM/DD から date.Date を生成
// 1 年以上前の放送回が残っていないという前提
// （すなわち 1 年以上前の放送回も残っていたら破綻）
func buildDateFromMMDD(now time.Time, mmdd string) (date.Date, error) {
	mmdds := strings.Split(mmdd, "/")
	month, err := strconv.Atoi(mmdds[0])
	if err != nil {
		return date.Date{}, errors.Wrap(errutil.ErrInternal, err.Error())
	}

	day, err := strconv.Atoi(mmdds[1])
	if err != nil {
		return date.Date{}, errors.Wrap(errutil.ErrInternal, err.Error())
	}

	sameYearDate := date.New(now.Year(), time.Month(month), day)
	yearDuration := 365 * 24 * time.Hour
	diff := now.Sub(time.Time(sameYearDate))
	if 0 < diff && diff < yearDuration {
		return sameYearDate, nil
	}

	oldYearDate := date.New(now.Year()-1, time.Month(month), day)
	return oldYearDate, nil
}
