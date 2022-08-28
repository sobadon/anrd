package agqr

import (
	"net/http"
	"net/url"

	"github.com/sobadon/anrd/domain/repository"
)

type client struct {
	httpClient     *http.Client
	programBaseURL *url.URL
	streamURL      *url.URL
}

func New() repository.Station {
	programBaseURL, err := url.Parse("https://www.joqr.co.jp/rss/program/json.php?type=ag")
	if err != nil {
		panic(err)
	}

	// 低画質
	// https://www.uniqueradio.jp/agplayer5/player.php から取得されるもの
	streamURL, err := url.Parse("https://hlsb2.cdnext.stream.ne.jp/agqr1next/aandg1next.m3u8")
	if err != nil {
		panic(err)
	}

	return &client{
		httpClient:     &http.Client{},
		programBaseURL: programBaseURL,
		streamURL:      streamURL,
	}
}
