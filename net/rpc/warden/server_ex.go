package warden

import (
	"github/gisvr/wallet/src/lib/log"
	"github/gisvr/wallet/src/lib/net/rpc/warden/cert"
	"google.golang.org/grpc"
)

type GrpcServerConfig struct {
	Server *ServerConfig      `yaml:"server"`
	Cert   *cert.ServerConfig `yaml:"cert"`
}

func NewServerWithCert(c *GrpcServerConfig, opts ...grpc.ServerOption) (s *Server) {
	var (
		cc *ServerConfig
	)
	if c != nil {
		opt, err := cert.NewServerOption(c.Cert)
		if err != nil {
			log.Fatal("%+v", err)
		}
		if opt != nil {
			opts = append(opts, opt)
		}
		cc = c.Server
	}
	s = NewServer(cc, opts...)
	return s
}
