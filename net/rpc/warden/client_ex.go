package warden

import (
	"github.com/gisvr/golib/log"
	"github.com/gisvr/golib/net/rpc/warden/cert"
	"google.golang.org/grpc"
)

type GrpcClientConfig struct {
	Client *ClientConfig      `yaml:"client"`
	Addr   string             `yaml:"addr"`
	Cert   *cert.ClientConfig `yaml:"cert"`
}

func NewClientWithCert(c *GrpcClientConfig, opts ...grpc.DialOption) *Client {
	var (
		cc *ClientConfig
	)
	if c != nil {
		opt, err := cert.NewClientOption(c.Cert)
		if err != nil {
			log.Fatal("%+v", err)
		}
		if opt != nil {
			opts = append(opts, opt)
		}
		cc = c.Client
	}
	return NewClient(cc, opts...)
}
