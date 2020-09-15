package prom

import (
	"context"
	"net"
	"net/http"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusServer struct {
	fatalOnError bool
	config       PrometheusOption
}

func NewPrometheusServer(opt PrometheusOption) *PrometheusServer {
	s := &PrometheusServer{
		config: opt,
	}
	return s
}

func (s *PrometheusServer) Run(ctx context.Context) error {
	if s.config.Enable {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		glog.V(4).Info("initializing Prometheus ...")
		http.Handle("/metrics", promhttp.Handler())
		srv := &http.Server{}
		ln, err := net.Listen("tcp", s.config.Addr)
		if err != nil {
			if s.fatalOnError {
				glog.Fatalf("initPrometheus failed. err = %v, config = %+v", err, s.config)
			}
			glog.Errorf("initPrometheus failed. err = %v, config = %+v", err, s.config)
			return err
		}

		go func() {
			err := srv.Serve(ln)
			if err != nil {
				if s.fatalOnError {
					glog.Fatalf("initPrometheus listen failed. err = %v, config = %+v", err, s.config)
				}
				glog.Errorf("initPrometheus listen failed. err = %v, config = %+v", err, s.config)
			}
			cancel()
		}()

		<-ctx.Done()
	}
	return nil
}
