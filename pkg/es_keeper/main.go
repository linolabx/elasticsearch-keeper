package es_keeper

import (
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"
)

type ESKeeper struct {
	resty *resty.Client
}

func NewESKeeper(dsn string) (*ESKeeper, error) {
	dsnUrl, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid dsn: %s", err)
	}

	if dsnUrl.Scheme == "" {
		dsnUrl.Scheme = "https"
	}

	if dsnUrl.Host == "" {
		return nil, fmt.Errorf("invalid dsn: host is required")
	}

	if dsnUrl.Scheme != "http" && dsnUrl.Scheme != "https" {
		return nil, fmt.Errorf("invalid dsn: scheme must be http or https")
	}

	reqUrl := url.URL{
		Scheme: dsnUrl.Scheme,
		Host:   dsnUrl.Host,
		Path:   dsnUrl.Path,
	}

	_resty := resty.New().SetBaseURL(reqUrl.String())

	apiKey := dsnUrl.Query().Get("api_key")
	if apiKey != "" {
		_resty.SetHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	}

	return &ESKeeper{resty: _resty}, nil
}
