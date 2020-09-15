package cert

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
)

type ClientConfig struct {
	CertFile   string `yaml:"certFile"`
	KeyFile    string `yaml:"keyFile"` //有CA时使用
	ServerName string `yaml:"serverName"`
	CaFile     string `yaml:"caFile"`
}

type ServerConfig struct {
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
	CaFile   string `yaml:"caFile"`
}

func loadCaFile(pool *x509.CertPool, caFile string) error {
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		return err
	}
	if ok := pool.AppendCertsFromPEM(ca); !ok {
		return errors.Errorf("certPool.AppendCertsFromPEM err")
	}
	return nil
}

func NewClientOption(c *ClientConfig) (grpc.DialOption, error) {
	if c != nil {
		var cc credentials.TransportCredentials
		if c.CaFile != "" {
			//使用CA模式
			cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
			if err != nil {
				return nil, err
			}
			certPool := x509.NewCertPool()
			err = loadCaFile(certPool, c.CaFile)
			if err != nil {
				return nil, err
			}
			cc = credentials.NewTLS(&tls.Config{
				Certificates: []tls.Certificate{cert},
				ServerName:   c.ServerName,
				RootCAs:      certPool,
			})
		} else {
			cert, err := credentials.NewClientTLSFromFile(c.CertFile, c.ServerName)
			if err != nil {
				return nil, err
			}
			cc = cert
		}
		return grpc.WithTransportCredentials(cc), nil
	} else {
		return grpc.WithInsecure(), nil
	}
}

func NewServerOption(c *ServerConfig) (grpc.ServerOption, error) {
	if c != nil {
		var cc credentials.TransportCredentials
		if c.CaFile != "" {
			//使用CA模式
			cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
			if err != nil {
				return nil, err
			}
			certPool := x509.NewCertPool()
			err = loadCaFile(certPool, c.CaFile)
			if err != nil {
				return nil, err
			}
			cc = credentials.NewTLS(&tls.Config{
				Certificates: []tls.Certificate{cert},
				ClientAuth:   tls.RequireAndVerifyClientCert,
				ClientCAs:    certPool,
			})
		} else {
			cert, err := credentials.NewServerTLSFromFile(c.CertFile, c.KeyFile)
			if err != nil {
				return nil, err
			}
			cc = cert
		}
		return grpc.Creds(cc), nil
	} else {
		return nil, nil
	}
}
