package rpc

import (
	"github.com/gisvr/golib/log"
	"github.com/gisvr/golib/net/rpc/logging"
	"github.com/gisvr/golib/net/rpc/tracing"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

// GetDefaultGrpcServerInterceptorOptions ...
func GetDefaultGrpcServerInterceptorOptions() []grpc.ServerOption {
	return GetGrpcServerInterceptorOptions(true, true, true)
}

// GetGrpcServerInterceptorOptions ...
func GetGrpcServerInterceptorOptions(enableLog, enableJaeger, enablePrometheus bool) []grpc.ServerOption {
	var interceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor
	if enableLog {
		interceptors = append(interceptors, logging.ServerLogInterceptor)
		streamInterceptors = append(streamInterceptors, logging.StreamServerLogInterceptor)
	}

	if enableJaeger {
		tracer := opentracing.GlobalTracer()
		if tracer == nil {
			log.Warn("no global tracer for server option.")
		} else {
			interceptors = append(interceptors, tracing.OpenTracingServerInterceptor(tracer))
			streamInterceptors = append(streamInterceptors, tracing.OpenTracingStreamServerInterceptor(tracer))
		}
	}
	if enablePrometheus {
		interceptors = append(interceptors, grpc_prometheus.UnaryServerInterceptor)
		streamInterceptors = append(streamInterceptors, grpc_prometheus.StreamServerInterceptor)
	}
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(streamInterceptors...)),
	}
}

// GetDefaultGrpcClientInterceptorOption .
func GetDefaultGrpcClientInterceptorOption() []grpc.DialOption {
	return GetGrpcClientInterceptorOption(true, true, true)
}

// GetGrpcClientInterceptorOption .
func GetGrpcClientInterceptorOption(enableLog, enableJaeger, enablePrometheus bool) []grpc.DialOption {
	var interceptors []grpc.UnaryClientInterceptor
	var streamInterceptors []grpc.StreamClientInterceptor

	if enableLog {
		interceptors = append(interceptors, logging.ClientLogInterceptor)
		streamInterceptors = append(streamInterceptors, logging.StreamClientLogInterceptor)
	}
	if enableJaeger {
		tracer := opentracing.GlobalTracer()
		if tracer == nil {
			log.Warn("no global tracer for client dial option.")
		} else {
			interceptors = append(interceptors, tracing.OpenTracingClientInterceptor(tracer))
			streamInterceptors = append(streamInterceptors, tracing.OpenTracingStreamClientInterceptor(tracer))
		}
	}
	if enablePrometheus {
		interceptors = append(interceptors, grpc_prometheus.UnaryClientInterceptor)
		streamInterceptors = append(streamInterceptors, grpc_prometheus.StreamClientInterceptor)
	}
	return []grpc.DialOption{
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(interceptors...)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(streamInterceptors...)),
	}
}
