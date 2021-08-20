// Copyright (c) 2021 Equres LLC. All rights reserved.
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/equres/sec/util"
)

// Serve file in HTTP and download to testdata directory
func TestHTTPDownloadFile(t *testing.T) {
	config, err := util.LoadConfig("./ci")
	if err != nil {
		t.Errorf(err.Error())
	}

	// Path for files
	filesPath := "cache/Archives/edgar/monthly"
	testServer := httptest.NewServer(http.FileServer(http.Dir(filesPath)))

	defer testServer.Close()

	config.Main.BaseURL = testServer.URL

	sec, err := util.NewSEC(config)
	if err != nil {
		t.Errorf(err.Error())
	}

	err = sec.DownloadFile(fmt.Sprintf("%v/%v", sec.BaseURL, "xbrlrss-2021-04.xml"), config)
	if err != nil {
		t.Errorf(err.Error())
	}
}
