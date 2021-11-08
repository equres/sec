package secreq

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

type SECReq struct {
	UserAgent       string
	RequestType     string
	IsEtag          bool
	IsContentLength bool
}

var Proxies = []string{
	"23.229.92.223:8800",
	"23.229.79.68:8800",
	"23.229.92.238:8800",
	"185.234.6.119:8800",
	"23.229.79.85:8800",
	"23.229.92.213:8800",
	"185.234.6.63:8800",
	"23.229.79.91:8800",
	"23.229.79.105:8800",
	"23.229.92.248:8800",
	"206.214.82.233:8800",
	"192.126.228.11:8800",
	"206.214.82.27:8800",
	"206.214.82.51:8800",
	"192.126.225.28:8800",
	"192.126.228.110:8800",
	"192.126.228.199:8800",
	"192.126.228.25:8800",
	"206.214.82.208:8800",
	"192.126.225.159:8800",
}

func (sr *SECReq) SendRequest(retryLimit int, rateLimit time.Duration, fullurl string) (*http.Response, error) {
	var resp *http.Response
	var etag string
	var contentLength string

	currentRetryLimit := retryLimit
	for currentRetryLimit > 0 {
		currentRetryLimit--

		// Choose a random proxy and use in HTTP client
		rand.Seed(time.Now().Unix())
		proxyNum := rand.Intn(len(Proxies))
		proxyUrl, err := url.Parse(fmt.Sprintf("http://%v", Proxies[proxyNum]))
		if err != nil {
			return nil, err
		}

		client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}

		req, err := http.NewRequest(sr.RequestType, fullurl, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", sr.UserAgent)

		resp, err = client.Do(req)
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
		log.Info(fmt.Sprintf("retried %v request %v times and failed", sr.RequestType, retryLimit))
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
