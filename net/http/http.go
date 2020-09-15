package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var httpOnce sync.Once
var netClient *http.Client
var ProxyUrl string
var TlsCheck bool

func GetHTTPClient() *http.Client {
	httpOnce.Do(func() {
		var proxy = func(_ *http.Request) (*url.URL, error) {
			if ProxyUrl == "" {
				return nil, nil
			}
			return url.Parse(ProxyUrl)
		}
		var netTransport = &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			Proxy:                 proxy,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: !TlsCheck},
		}
		netClient = &http.Client{
			Timeout:   time.Second * 5,
			Transport: netTransport,
		}
	})
	return netClient
}
