package secreq

import (
	"fmt"
	"net/http"
	"time"
)

type SECReq struct {
	UserAgent   string
	RequestType string
}

func (sr *SECReq) SendRequest(retryLimit int, rateLimit time.Duration, fullurl string) (*http.Response, error) {
	var resp *http.Response
	var etag string

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

		etag = resp.Header.Get("eTag")
		if etag != "" {
			break
		}
		time.Sleep(rateLimit)
	}

	if currentRetryLimit == 0 && etag == "" {
		return nil, fmt.Errorf("retried %v request %v times and failed", sr.RequestType, retryLimit)
	}

	return resp, nil
}

func NewSECReqHEAD() *SECReq {
	return &SECReq{
		UserAgent:   "Equres LLC wojciech@koszek.com",
		RequestType: http.MethodHead,
	}
}

func NewSECReqGET() *SECReq {
	return &SECReq{
		UserAgent:   "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0)",
		RequestType: http.MethodGet,
	}
}
