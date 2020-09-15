package tracing

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	llog "github.com/gisvr/golib/log"
)

var (
	// Morally a const:
	gRPCComponentTag = opentracing.Tag{
		Key:   string(ext.Component),
		Value: "gRPC",
	}
)

// NewJaegerTracer ..
func NewJaegerTracer(serviceName string, flushInterval time.Duration, hostPort string) (tracer opentracing.Tracer, closer io.Closer, err error) {
	cfg := config.Configuration{
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},

		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: flushInterval,
			LocalAgentHostPort:  hostPort,
		},
	}

	tracer, closer, err = cfg.NewTracer(
		config.Logger(jaeger.StdLogger),
	)
	if err != nil {
		return
	}
	opentracing.SetGlobalTracer(tracer)
	return
}

//MDReaderWriter metadata Reader and Writer
type MDReaderWriter struct {
	MD metadata.MD
}

// Set ..
func (rw MDReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	rw.MD[key] = append(rw.MD[key], url.QueryEscape(val))
}

// ForeachKey conforms to the TextMapReader interface.
func (rw MDReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range rw.MD {
		for _, v := range vals {
			unescapeV, err := url.QueryUnescape(v)
			if err == nil {
				if err := handler(k, unescapeV); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// TransferContext ..
func TransferContext(server, client context.Context) (newClient context.Context, err error) {
	if parent := opentracing.SpanFromContext(server); parent != nil {
		newClient = opentracing.ContextWithSpan(client, parent)
	} else {
		err = errors.New("server context hasn't span")
	}
	return
}

// OpenTracingServerInterceptor rewrite server's interceptor with open tracing
func OpenTracingServerInterceptor(tracer opentracing.Tracer) grpc.UnaryServerInterceptor {
	grpcOpts := newOptions()
	grpcOpts.apply(LogPayloads())
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		metaData := fmt.Sprintf("metadata %+v", md)
		rw := MDReaderWriter{md}
		spanContext, err := tracer.Extract(opentracing.TextMap, rw)
		if err != nil {
			llog.Error("extract from metadata err %s\n", err)
		}

		if grpcOpts.inclusionFunc != nil &&
			!grpcOpts.inclusionFunc(spanContext, info.FullMethod, req, nil) {
			return handler(ctx, req)
		}

		serverSpan := tracer.StartSpan(
			"server:"+info.FullMethod,
			opentracing.ChildOf(spanContext), // ext.RPCServerOption(spanContext),
			opentracing.Tag{
				Key:   "Metadata",
				Value: metaData,
			},
			gRPCComponentTag,
			ext.SpanKindRPCServer,
		)

		ctx = opentracing.ContextWithSpan(ctx, serverSpan)
		if grpcOpts.logPayloads {
			serverSpan.LogFields(log.Object("gRPC request", req))
		}
		resp, err = handler(ctx, req)
		if err == nil {
			if grpcOpts.logPayloads {
				serverSpan.LogFields(log.Object("gRPC response", resp))
			}
		} else {
			SetSpanTags(serverSpan, err, false)
			serverSpan.LogFields(log.String("event", "error"), log.String("message", err.Error()))
		}
		if grpcOpts.decorator != nil {
			grpcOpts.decorator(serverSpan, info.FullMethod, req, resp, err)
		}
		serverSpan.Finish()
		return resp, err
	}
}

// OpenTracingClientInterceptor ..
func OpenTracingClientInterceptor(tracer opentracing.Tracer) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, resp interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		cliSpan, clientCtx := opentracing.StartSpanFromContext(
			ctx,
			"client:"+method,
			gRPCComponentTag,
			ext.SpanKindRPCClient,
		)

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}
		mdWriter := MDReaderWriter{md}
		err := tracer.Inject(cliSpan.Context(), opentracing.TextMap, mdWriter)
		if err != nil {
			cliSpan.LogFields(log.String("event", "trace inject error"), log.Error(err))
		}
		ctx = metadata.NewOutgoingContext(clientCtx, md)
		err = invoker(ctx, method, req, resp, cc, opts...)
		if err == nil {
			cliSpan.LogFields(log.Object("gRPC response", resp))
		} else {
			cliSpan.LogFields(log.String("event", "error"), log.Error(err))
			//cliSpan.LogFields(log.String("err", err.Error()))
		}
		cliSpan.Finish()
		return err
	}
}

// OpenTracingStreamServerInterceptor returns a grpc.StreamServerInterceptor suitable
// for use in a grpc.NewServer call. The interceptor instruments streaming RPCs by
// creating a single span to correspond to the lifetime of the RPC's stream.
//
// For example:
//
//     s := grpc.NewServer(
//         ...,  // (existing ServerOptions)
//         grpc.StreamInterceptor(otgrpc.OpenTracingStreamServerInterceptor(tracer)))
//
// All gRPC server spans will look for an OpenTracing SpanContext in the gRPC
// metadata; if found, the server span will act as the ChildOf that RPC
// SpanContext.
//
// Root or not, the server Span will be embedded in the context.Context for the
// application-specific gRPC handler(s) to access.
func OpenTracingStreamServerInterceptor(tracer opentracing.Tracer, optFuncs ...Option) grpc.StreamServerInterceptor {
	otgrpcOpts := newOptions()
	otgrpcOpts.apply(optFuncs...)
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			md = metadata.New(nil)
		}

		metaData := fmt.Sprintf("metadata %+v", md)
		rw := MDReaderWriter{md}
		spanContext, err := tracer.Extract(opentracing.TextMap, rw)
		if err != nil {
			llog.Error("extract from metadata err %s\n", err)
		}

		if otgrpcOpts.inclusionFunc != nil &&
			!otgrpcOpts.inclusionFunc(spanContext, info.FullMethod, nil, nil) {
			return handler(srv, ss)
		}

		serverSpan := tracer.StartSpan(
			"server:"+info.FullMethod,
			opentracing.ChildOf(spanContext), // ext.RPCServerOption(spanContext),
			opentracing.Tag{
				Key:   "Metadata",
				Value: metaData,
			},
			gRPCComponentTag,
			ext.SpanKindRPCServer,
		)
		defer serverSpan.Finish()
		ss = &openTracingServerStream{
			ServerStream: ss,
			ctx:          opentracing.ContextWithSpan(ss.Context(), serverSpan),
		}
		err = handler(srv, ss)
		if err != nil {
			SetSpanTags(serverSpan, err, false)
			serverSpan.LogFields(log.String("event", "error"), log.String("message", err.Error()))
		}
		if otgrpcOpts.decorator != nil {
			otgrpcOpts.decorator(serverSpan, info.FullMethod, nil, nil, err)
		}
		return err
	}
}

type openTracingServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss *openTracingServerStream) Context() context.Context {
	return ss.ctx
}

// OpenTracingStreamClientInterceptor returns a grpc.StreamClientInterceptor suitable
// for use in a grpc.Dial call. The interceptor instruments streaming RPCs by creating
// a single span to correspond to the lifetime of the RPC's stream.
//
// For example:
//
//     conn, err := grpc.Dial(
//         address,
//         ...,  // (existing DialOptions)
//         grpc.WithStreamInterceptor(otgrpc.OpenTracingStreamClientInterceptor(tracer)))
//
// All gRPC client spans will inject the OpenTracing SpanContext into the gRPC
// metadata; they will also look in the context.Context for an active
// in-process parent Span and establish a ChildOf reference if such a parent
// Span could be found.
func OpenTracingStreamClientInterceptor(tracer opentracing.Tracer, optFuncs ...Option) grpc.StreamClientInterceptor {
	otgrpcOpts := newOptions()
	otgrpcOpts.apply(optFuncs...)
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		var err error
		var parentCtx opentracing.SpanContext
		if parent := opentracing.SpanFromContext(ctx); parent != nil {
			parentCtx = parent.Context()
		}
		if otgrpcOpts.inclusionFunc != nil &&
			!otgrpcOpts.inclusionFunc(parentCtx, method, nil, nil) {
			return streamer(ctx, desc, cc, method, opts...)
		}

		clientSpan := tracer.StartSpan(
			"client:"+method,
			opentracing.ChildOf(parentCtx),
			ext.SpanKindRPCClient,
			gRPCComponentTag,
		)
		ctx = injectSpanContext(ctx, tracer, clientSpan)
		cs, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			clientSpan.LogFields(log.String("event", "error"), log.String("message", err.Error()))
			SetSpanTags(clientSpan, err, true)
			clientSpan.Finish()
			return cs, err
		}
		return newOpenTracingClientStream(cs, method, desc, clientSpan, otgrpcOpts), nil
	}
}

func newOpenTracingClientStream(cs grpc.ClientStream, method string, desc *grpc.StreamDesc, clientSpan opentracing.Span, otgrpcOpts *options) grpc.ClientStream {
	finishChan := make(chan struct{})

	isFinished := new(int32)
	*isFinished = 0
	finishFunc := func(err error) {
		// The current OpenTracing specification forbids finishing a span more than
		// once. Since we have multiple code paths that could concurrently call
		// `finishFunc`, we need to add some sort of synchronization to guard against
		// multiple finishing.
		if !atomic.CompareAndSwapInt32(isFinished, 0, 1) {
			return
		}
		close(finishChan)
		defer clientSpan.Finish()
		if err != nil {
			clientSpan.LogFields(log.String("event", "error"), log.String("message", err.Error()))
			SetSpanTags(clientSpan, err, true)
		}
		if otgrpcOpts.decorator != nil {
			otgrpcOpts.decorator(clientSpan, method, nil, nil, err)
		}
	}
	go func() {
		select {
		case <-finishChan:
			// The client span is being finished by another code path; hence, no
			// action is necessary.
		case <-cs.Context().Done():
			finishFunc(cs.Context().Err())
		}
	}()
	otcs := &openTracingClientStream{
		ClientStream: cs,
		desc:         desc,
		finishFunc:   finishFunc,
	}

	// The `ClientStream` interface allows one to omit calling `Recv` if it's
	// known that the result will be `io.EOF`. See
	// http://stackoverflow.com/q/42915337
	// In such cases, there's nothing that triggers the span to finish. We,
	// therefore, set a finalizer so that the span and the context goroutine will
	// at least be cleaned up when the garbage collector is run.
	runtime.SetFinalizer(otcs, func(otcs *openTracingClientStream) {
		otcs.finishFunc(nil)
	})
	return otcs
}

type openTracingClientStream struct {
	grpc.ClientStream
	desc       *grpc.StreamDesc
	finishFunc func(error)
}

func (cs *openTracingClientStream) Header() (metadata.MD, error) {
	md, err := cs.ClientStream.Header()
	if err != nil {
		cs.finishFunc(err)
	}
	return md, err
}

func (cs *openTracingClientStream) SendMsg(m interface{}) error {
	err := cs.ClientStream.SendMsg(m)
	if err != nil {
		cs.finishFunc(err)
	}
	return err
}

func (cs *openTracingClientStream) RecvMsg(m interface{}) error {
	err := cs.ClientStream.RecvMsg(m)
	if err == io.EOF {
		cs.finishFunc(nil)
		return err
	} else if err != nil {
		cs.finishFunc(err)
		return err
	}
	if !cs.desc.ServerStreams {
		cs.finishFunc(nil)
	}
	return err
}

func (cs *openTracingClientStream) CloseSend() error {
	err := cs.ClientStream.CloseSend()
	if err != nil {
		cs.finishFunc(err)
	}
	return err
}

func injectSpanContext(ctx context.Context, tracer opentracing.Tracer, clientSpan opentracing.Span) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}
	mdWriter := MDReaderWriter{md}
	err := tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, mdWriter)
	// We have no better place to record an error than the Span itself :-/
	if err != nil {
		clientSpan.LogFields(log.String("event", "Tracer.Inject() failed"), log.Error(err))
	}
	return metadata.NewOutgoingContext(ctx, md)
}
