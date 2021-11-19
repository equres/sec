package secreq

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/equres/sec/pkg/config"
	log "github.com/sirupsen/logrus"
)

type SECReq struct {
	UserAgent       string
	RequestType     string
	IsEtag          bool
	IsContentLength bool
	Config          config.Config
}

func (sr *SECReq) SendRequest(retryLimit int, rateLimit time.Duration, fullurl string) (*http.Response, error) {
	var resp *http.Response
	var etag string
	var contentLength string

	currentRetryLimit := retryLimit
	waitIfFail := 2
	for currentRetryLimit > 0 {
		currentRetryLimit--

		client := &http.Client{}
		if len(sr.Config.Proxies.Addresses) > 0 {
			// Choose a random proxy and use in HTTP client
			rand.Seed(time.Now().Unix())
			proxyNum := rand.Intn(len(sr.Config.Proxies.Addresses))
			proxyUrl, err := url.Parse(fmt.Sprintf("http://%v", sr.Config.Proxies.Addresses[proxyNum]))
			if err != nil {
				return nil, err
			}

			client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
		}

		req, err := http.NewRequest(sr.RequestType, fullurl, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", sr.UserAgent)

		resp, err = client.Do(req)
		if err != nil {
			time.Sleep(time.Duration(waitIfFail) * time.Second)
			waitIfFail *= 2
			continue
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

func NewSECReqHEAD(cfg config.Config) *SECReq {
	return &SECReq{
		UserAgent:   "Equres LLC, wojciech@koszek.com",
		RequestType: http.MethodHead,
		Config:      cfg,
	}
}

func NewSECReqGET(cfg config.Config) *SECReq {
	return &SECReq{
		UserAgent:   "Equres LLC, wojciech@koszek.com",
		RequestType: http.MethodGet,
		Config:      cfg,
	}
}
