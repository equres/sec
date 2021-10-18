package secreq

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type SECReq struct {
	UserAgent       string
	RequestType     string
	IsEtag          bool
	IsContentLength bool
}

func (sr *SECReq) SendRequest(retryLimit int, rateLimit time.Duration, fullurl string) (*http.Response, error) {
	var resp *http.Response
	var etag string
	var contentLength string

	currentRetryLimit := retryLimit
	for currentRetryLimit > 0 {
		currentRetryLimit--
		req, err := http.NewRequest(sr.RequestType, fullurl, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", sr.UserAgent)

		resp, err = new(http.Client).Do(req)
		if err != nil {
			return nil, err
		}

		if sr.RequestType == http.MethodGet {
			responseBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			resp.Body = ioutil.NopCloser(bytes.NewBuffer(responseBody))

			fileExtensions, err := mime.ExtensionsByType(http.DetectContentType(responseBody))
			if err != nil || len(fileExtensions) == 0 {
				return nil, err
			}

			if strings.ToLower(fileExtensions[0]) != strings.ToLower(filepath.Ext(fullurl)) {
				continue
			}
		}

		if sr.IsEtag {
			etag = resp.Header.Get("eTag")
			if etag != "" {
				break
			}
		}

		if sr.IsContentLength {
			contentLength = resp.Header.Get("Content-Length")
			if contentLength != "" {
				break
			}
		}
		time.Sleep(rateLimit)
	}

	if currentRetryLimit == 0 && etag == "" && contentLength == "" {
		logrus.Error(fmt.Sprintf("retried %v request %v times and failed", sr.RequestType, retryLimit))
		return nil, fmt.Errorf("retried %v request %v times and failed", sr.RequestType, retryLimit)
	}

	// Sleep before starting any other HTTP request
	time.Sleep(rateLimit)

	return resp, nil
}

func NewSECReqHEAD() *SECReq {
	return &SECReq{
		UserAgent:   "Equres LLC, wojciech@koszek.com",
		RequestType: http.MethodHead,
	}
}

func NewSECReqGET() *SECReq {
	return &SECReq{
		UserAgent:   "Equres LLC, wojciech@koszek.com",
		RequestType: http.MethodGet,
	}
}
