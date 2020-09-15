package rpc

import (
	"context"
	"errors"
	"flag"
	"github.com/gisvr/golib/utils"
	"net"
	"time"

	"github.com/gisvr/golib/log"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type GrpcServer struct {
	fatalOnError      bool
	config            GrpcOption
	beforeServerStart func(server *grpc.Server)
	opts              []grpc.ServerOption
}

//new grpcserver时grpc.ServerOption已经有log/tracing/prometheus由option指定
func NewGRPCServer(option GrpcOption, beforeServerStart func(server *grpc.Server), opt ...grpc.ServerOption) *GrpcServer {
	s := &GrpcServer{
		fatalOnError:      false,
		config:            option,
		beforeServerStart: beforeServerStart,
		opts:              opt,
	}
	return s
}

func StartGPRCServer(option GrpcOption, beforeServerStart func(server *grpc.Server), opt ...grpc.ServerOption) *utils.ServerAgent {
	s := NewGRPCServer(option, beforeServerStart, opt...)
	h := utils.NewServerAgent()
	h.RunServer(s)
	return h
}

func (s *GrpcServer) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	//glog.V(4).Infof("[Toolset] starting grpc server, config = %+v", s.config)
	server := s.createServer()

	s.beforeServerStart(server)

	l, err := net.Listen("tcp", s.config.GrpcAddr)
	if err != nil {
		if s.fatalOnError {
			log.Fatal("start grpc error:%v", err)
		}
		log.Errorf("start grpc error:%v", err)
		return errors.New("server unable to bind on provided host")
	}

	go func() {
		if err := server.Serve(l); err != nil {
			if s.fatalOnError {
				log.Fatal("starting grpc server failed. err is %v", err)
			}
			log.Errorf("starting grpc server failed. err is %v", err)
		}
		cancel()
	}()

	<-ctx.Done()

	log.Info("grpc server shutting down")

	server.Stop()
	return nil
}

func (s *GrpcServer) createServer() (server *grpc.Server) {
	if s.config.EnableDumpGrpcMeta {
		flag.Set("dump-grpc-meta", "true")
	}
	if s.config.EnableDumpGrpcBody {
		flag.Set("dump-grpc-body", "true")
	}

	opts := GetGrpcServerInterceptorOptions(s.config.EnableLog, s.config.EnableJaeger, s.config.EnablePrometheus)
	s.opts = append(s.opts, opts...)
	s.opts = append(s.opts, s.createSendMsgSizeOption(), s.createRecvMsgSizeOption(), s.createKeepAliveOption())
	server = grpc.NewServer(s.opts...)

	if s.config.EnablePrometheus {
		grpc_prometheus.Register(server)
	}

	return
}

const size10M = 1024 * 1024 * 10

func (s *GrpcServer) createRecvMsgSizeOption() grpc.ServerOption {
	if s.config.MaxRecvMsgSize == 0 {
		s.config.MaxRecvMsgSize = size10M
	}
	return grpc.MaxRecvMsgSize(s.config.MaxRecvMsgSize)
}

func (s *GrpcServer) createSendMsgSizeOption() grpc.ServerOption {
	if s.config.MaxSendMsgSize == 0 {
		s.config.MaxSendMsgSize = size10M
	}
	return grpc.MaxSendMsgSize(s.config.MaxSendMsgSize)
}

func (s *GrpcServer) createKeepAliveOption() grpc.ServerOption {
	if s.config.KeepAliveTime == 0 {
		s.config.KeepAliveTime = 20
	}
	if s.config.KeepAliveTimeout == 0 {
		s.config.KeepAliveTimeout = 10
	}

	return grpc.KeepaliveParams(
		keepalive.ServerParameters{
			Time:    time.Duration(s.config.KeepAliveTime) * time.Second,
			Timeout: time.Duration(s.config.KeepAliveTimeout) * time.Second,
		},
	)
}

type GrpcClient struct {
	fatalOnError bool
	config       GrpcOption
	opts         []grpc.DialOption
}

//new grpcserver时grpc.DialOption已经有log/tracing/prometheus由option指定
func NewGRPClient(option GrpcOption, opt ...grpc.DialOption) (*grpc.ClientConn, error) {
	c := &GrpcClient{
		fatalOnError: false,
		config:       option,
		opts:         opt,
	}
	return c.createClient()
}

func (s *GrpcClient) createClient() (client *grpc.ClientConn, err error) {
	if s.config.EnableDumpGrpcMeta {
		flag.Set("dump-grpc-meta", "true")
	}
	if s.config.EnableDumpGrpcBody {
		flag.Set("dump-grpc-body", "true")
	}

	opts := GetGrpcClientInterceptorOption(s.config.EnableLog, s.config.EnableJaeger, s.config.EnablePrometheus)
	s.opts = append(s.opts, opts...)
	s.opts = append(s.opts, s.createSendMsgSizeOption(), s.createRecvMsgSizeOption(), s.createKeepAliveOption())
	client, err = grpc.Dial(s.config.GrpcAddr, s.opts...)
	return
}

func (s *GrpcClient) createRecvMsgSizeOption() grpc.DialOption {
	if s.config.MaxRecvMsgSize == 0 {
		s.config.MaxRecvMsgSize = size10M
	}
	return grpc.WithReadBufferSize(s.config.MaxRecvMsgSize)
}

func (s *GrpcClient) createSendMsgSizeOption() grpc.DialOption {
	if s.config.MaxSendMsgSize == 0 {
		s.config.MaxSendMsgSize = size10M
	}
	return grpc.WithWriteBufferSize(s.config.MaxSendMsgSize)
}

func (s *GrpcClient) createKeepAliveOption() grpc.DialOption {
	if s.config.KeepAliveTime == 0 {
		s.config.KeepAliveTime = 20
	}
	if s.config.KeepAliveTimeout == 0 {
		s.config.KeepAliveTimeout = 10
	}

	return grpc.WithKeepaliveParams(
		keepalive.ClientParameters{
			Time:    time.Duration(s.config.KeepAliveTime) * time.Second,
			Timeout: time.Duration(s.config.KeepAliveTimeout) * time.Second,
		},
	)
}
