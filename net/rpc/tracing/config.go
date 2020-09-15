package tracing

import (
	"io"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
)

// Config is used for Tracer creation
type Config struct {
	ServerAddr    string        `yaml:"server_addr"`
	ServiceName   string        `yaml:"service_name"`
	FlushInterval time.Duration `yaml:"flush_interval"`
}

// TracerFromConfig ..
func TracerFromConfig(opt *Config) (opentracing.Tracer, io.Closer, error) {
	if opt.ServiceName == "" {
		opt.ServiceName = "wallet"
	}

	if opt.FlushInterval == 0 {
		opt.FlushInterval = 1 * time.Second
	}

	return NewJaegerTracer(opt.ServiceName, opt.FlushInterval, opt.ServerAddr)
}
