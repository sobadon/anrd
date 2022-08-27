package onsen

import (
	"net/http"

	"github.com/sobadon/anrd/domain/repository"
)

type client struct {
	httpClient *http.Client
}

func New() repository.Station {
	return &client{
		httpClient: &http.Client{},
	}
}
