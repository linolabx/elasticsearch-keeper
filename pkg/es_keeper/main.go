package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/go-resty/resty/v2"
)

type EsKeeper struct {
	resty *resty.Client
}

func NewEsKeeper(dsn string) *EsKeeper {
	dsnUrl, err := url.Parse(dsn)
	if err != nil {
		log.Fatalf("invalid dsn: %s", err)
	}

	if dsnUrl.Scheme == "" {
		dsnUrl.Scheme = "https"
	}

	if dsnUrl.Host == "" {
		log.Fatalf("invalid dsn: %s", "host is required")
	}

	if dsnUrl.Scheme != "http" && dsnUrl.Scheme != "https" {
		log.Fatalf("invalid dsn: %s", "scheme must be http or https")
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

	return &EsKeeper{resty: _resty}
}
