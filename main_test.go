// Copyright (c) 2021 Equres LLC. All rights reserved.
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/equres/sec/config"
	"github.com/equres/sec/sec"
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

	s, err := sec.NewSEC(cfg)
	if err != nil {
		t.Errorf(err.Error())
	}

	cfg.Main.CacheDir = "./testdata"
	err = s.DownloadFile(fmt.Sprintf("%v/%v", s.BaseURL, "xbrlrss-2021-04.xml"), cfg)
	if err != nil {
		t.Errorf(err.Error())
	}
}
