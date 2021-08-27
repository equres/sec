package secreq

import "net/http"

type SECReq struct {
	UserAgent string
}

func NewSECReqHEAD(fullurl string) (*http.Request, error) {
	req, err := http.NewRequest("HEAD", fullurl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Equres LLC wojciech@koszek.com")

	return req, nil
}

func NewSECReqGET(fullurl string) (*http.Request, error) {
	req, err := http.NewRequest("GET", fullurl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0)")

	return req, nil
}
