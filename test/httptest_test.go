// Copyright (c) 2021 Equres LLC. All rights reserved.
package test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/equres/sec/util"
	"golang.org/x/net/html/charset"
)

func TestGetFiles(t *testing.T) {
	var files []string

	config, err := util.LoadConfig("../ci")
	if err != nil {
		t.Errorf(err.Error())
	}
	filePath := fmt.Sprintf(".%v%v", config.Main.CacheDir, "/Archives/edgar/monthly/")
	err = filepath.Walk(filePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Errorf(err.Error())
	}
	for _, v := range files {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, v)
		}))

		defer testServer.Close()
		res, err := http.Get(testServer.URL)
		if err != nil {
			t.Errorf(err.Error())
		}

		data, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Errorf(err.Error())
		}

		var rssFile util.RSSFile

		reader := bytes.NewReader(data)
		decoder := xml.NewDecoder(reader)
		decoder.CharsetReader = charset.NewReaderLabel
		err = decoder.Decode(&rssFile)
		if err != nil {
			t.Errorf(err.Error())
		}

		fmt.Println(rssFile.Channel.Title)
	}

}
