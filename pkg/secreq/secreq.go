package secreq

import (
	"fmt"
	"net/http"
	"time"
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
		return nil, fmt.Errorf("retried %v request %v times and failed", sr.RequestType, retryLimit)
	}

	// Sleep before starting any other HTTP request
	time.Sleep(rateLimit)

	return resp, nil
}

func NewSECReqHEAD() *SECReq {
	return &SECReq{
		UserAgent:   "Koszek Systems wojciech@koszek.com",
		RequestType: http.MethodHead,
	}
}

func NewSECReqGET() *SECReq {
	return &SECReq{
		UserAgent:   "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0)",
		RequestType: http.MethodGet,
	}
}
