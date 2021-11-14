// Copyright (c) 2021 Koszek Systems. All rights reserved.
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/equres/sec/pkg/config"
	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/download"
	"github.com/equres/sec/pkg/sec"
)

// Serve file in HTTP and download to testdata directory
func TestHTTPDownloadFile(t *testing.T) {
	cfg, err := config.LoadConfig("./ci")
	if err != nil {
		t.Errorf(err.Error())
	}

	// Path for files
	filesPath := "./cache/Archives/edgar/monthly/"
	testServer := httptest.NewServer(http.FileServer(http.Dir(filesPath)))

	defer testServer.Close()

	cfg.Main.BaseURL = testServer.URL
	cfg.Main.CacheDir = "./testdata"

	s, err := sec.NewSEC(cfg)
	if err != nil {
		t.Errorf(err.Error())
	}

	db, err := database.ConnectDB(cfg)
	if err != nil {
		t.Errorf(err.Error())
	}

	downloader := download.NewDownloader(cfg)

	not_download, err := downloader.FileCorrect(db, fmt.Sprintf("%v/%v", s.BaseURL, "xbrlrss-2021-04.xml"), 0, "")
	if err != nil {
		t.Errorf(err.Error())
	}

	if !not_download {
		err = downloader.DownloadFile(db, fmt.Sprintf("%v/%v", s.BaseURL, "xbrlrss-2021-04.xml"))
		if err != nil {
			t.Errorf(err.Error())
		}
	}
}
