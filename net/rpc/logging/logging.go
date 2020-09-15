package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/gisvr/golib/utils"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/gisvr/golib/log"
)

const (
	Auto = iota
	Enable
	Disable
)

var (
	DumpGrpcBody tribool
	DumpGrpcMeta tribool
)

type tribool int

func (b *tribool) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		err = errors.New("parse error")
	}
	if v {
		*b = Enable
	} else {
		*b = Disable
	}
	return err
}

func (b *tribool) String() string {
	if *b == Auto {
		return "unset"
	}
	return strconv.FormatBool(*b == Enable)
}

func (b *tribool) IsBoolFlag() bool { return true }

func (b *tribool) Enable() {
	*b = Enable
}

func (b *tribool) Disable() {
	*b = Disable
}

// LogLevel set loglevel ...
var LogLevel = int32(6)
var LogLevelDebug = int32(8)

var rng = rand.New(&lockedSource{src: rand.NewSource(time.Now().UnixNano())})

type requestIDKey struct{}
type requestLogKey struct{}

type lockedSource struct {
	mut sync.Mutex
	src rand.Source
}

func (r *lockedSource) Int63() (n int64) {
	r.mut.Lock()
	n = r.src.Int63()
	r.mut.Unlock()
	return
}

func (r *lockedSource) Seed(seed int64) {
	r.mut.Lock()
	r.src.Seed(seed)
	r.mut.Unlock()
}

func randomNumber() uint64 {
	return uint64(rng.Int63())
}

func getRandomID() string {
	return fmt.Sprintf("%x", randomNumber())
}

// GetRequestIDFromContext 从 Context 中获取 request id
func GetRequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}

	return getRandomID()
}

type streamWrapper struct {
	grpc.ServerStream

	WrapperContext context.Context

	ID string
}

func (s *streamWrapper) Context() context.Context {
	return s.WrapperContext
}

func (s *streamWrapper) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		log.Errorf("[stream] id:%v recv error %v, type:%T",
			s.ID, err, m)
		return err
	}

	_, isProtoMsg := m.(proto.Message)
	if canDumpGrpcBody() && isProtoMsg {
		log.Infof("[stream] id:%v recv msg type:%T data:\n%#v",
			s.ID, m, Marshal(m))
	} else {
		log.Infof("[stream] id:%v recv msg type:%T", s.ID, m)
	}

	return nil
}

func (s *streamWrapper) SendMsg(m interface{}) error {
	err := s.ServerStream.SendMsg(m)
	_, isProtoMsg := m.(proto.Message)
	if canDumpGrpcBody() && isProtoMsg {
		log.Infof("[stream] id:%v send msg type:%T err:%v data:\n%v",
			s.ID, m, err, Marshal(m))
	} else {
		log.Infof("[stream] id:%v send msg type:%T err:%v", s.ID, m, err)
	}

	return err
}

// StreamLogInterceptor log for stream
var StreamLogInterceptor = StreamServerLogInterceptor

func canDumpGrpcBody() bool {
	if DumpGrpcBody == Disable {
		return false
	}
	if DumpGrpcBody == Enable {
		return true
	}
	return log.V(LogLevel).IsEnable()
}

func canDumpGrpcMeta() bool {
	return DumpGrpcMeta == Enable
}

func printWithData(data interface{}, withData bool, format string, args ...interface{}) {
	if withData {
		log.Infof("%s data:\n%v", fmt.Sprintf(format, args...), Marshal(data))
	} else {
		log.Infof(format, args...)
	}
}

// ServerLogInterceptor 调试环境下输入输出的打印
func ServerLogInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var id string

	device := "nil"

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if lv := md.Get("device"); len(lv) == 1 {
			device = lv[0]
		}

		if lv := md.Get("bf-trace-id"); len(lv) == 1 {
			id = lv[0]
		}
	}

	if id == "" {
		id = getRandomID()
	}

	if canDumpGrpcMeta() {
		log.Infof("[req] id:%v method:%v metadata:%+v", id, info.FullMethod, md)
	}

	if canDumpGrpcBody() {
		log.Infof("[req] id:%v method:%v device:%v req:\n%v",
			id, info.FullMethod, device, Marshal(req))
	}

	st := time.Now()

	ctx = context.WithValue(ctx, requestIDKey{}, id)

	res, err := handler(ctx, req)

	printWithData(res, canDumpGrpcBody(),
		"[rsp] id:%v method:%v cost:%v err:%v",
		id, info.FullMethod, time.Since(st), err)

	return res, err
}

// Fields defines map with key string and value interface
type Fields map[string]interface{}

type requestLog struct {
	ID   string
	Data Fields

	cache string

	sync.Mutex
}

func (r *requestLog) CreateRequest() *requestLog {
	nr := &requestLog{
		ID:   fmt.Sprintf("%s-%s", r.ID, getRandomID()),
		Data: make(Fields, len(r.Data)+5),
	}

	for k, v := range r.Data {
		nr.Data[k] = v
	}

	nr.Data["id"] = nr.ID

	return nr
}

func (r *requestLog) Context(ctx context.Context) context.Context {
	return context.WithValue(context.WithValue(ctx, requestIDKey{}, r.ID), requestLogKey{}, r)
}

func (r *requestLog) Fields() string {
	r.Lock()
	defer r.Unlock()

	if r.cache != "" {
		return r.cache
	}

	b := &bytes.Buffer{}

	keys := make([]string, 0, len(r.Data))
	for k := range r.Data {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		r.appendKeyValue(b, key, r.Data[key])
	}

	r.cache = b.String() + " "

	return r.cache
}

func (r *requestLog) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte(':')
	r.appendValue(b, value)
}

func (r *requestLog) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	b.WriteString(stringVal)
}

func (r *requestLog) Set(kv ...interface{}) {
	if len(kv)%2 == 1 {
		panic(fmt.Sprintf("requestLog: Set got the odd number of input: %d", len(kv)))
	}

	r.Lock()

	var key string
	for i, s := range kv {
		if i%2 == 0 {
			stringVal, ok := s.(string)
			if !ok {
				stringVal = fmt.Sprint(s)
			}
			key = stringVal
			continue
		}
		r.Data[key] = s
	}

	r.cache = ""

	r.Unlock()
}

func (r *requestLog) Info(args ...interface{}) {
	log.InfoDepth(1, append([]interface{}{r.Fields()}, args...)...)
}

func (r *requestLog) Infof(format string, args ...interface{}) {
	log.InfoDepth(1, r.Fields(), fmt.Sprintf(format, args...))
}

func (r *requestLog) Infoln(args ...interface{}) {
	log.InfoDepth(1, r.Fields(), fmt.Sprintln(args...))
}

func (r *requestLog) Debug(args ...interface{}) {
	if log.V(int32(LogLevelDebug)).IsEnable() {
		log.InfoDepth(1, append([]interface{}{r.Fields()}, args...)...)
	}
}

func (r *requestLog) Debugf(format string, args ...interface{}) {
	if log.V(int32(LogLevelDebug)).IsEnable() {
		log.InfoDepth(1, r.Fields(), fmt.Sprintf(format, args...))
	}
}

func (r *requestLog) Debugln(args ...interface{}) {
	if log.V(int32(LogLevelDebug)).IsEnable() {
		log.InfoDepth(1, r.Fields(), fmt.Sprintln(args...))
	}
}

func (r *requestLog) Error(args ...interface{}) {
	log.ErrorDepth(1, append([]interface{}{r.Fields()}, args...)...)
}

func (r *requestLog) Errorf(format string, args ...interface{}) {
	log.ErrorDepth(1, r.Fields(), fmt.Sprintf(format, args...))
}

func (r *requestLog) Errorln(args ...interface{}) {
	log.ErrorDepth(1, r.Fields(), fmt.Sprintln(args...))
}

func (r *requestLog) Warn(args ...interface{}) {
	log.WarnDepth(1, append([]interface{}{r.Fields()}, args...)...)
}

func (r *requestLog) Warnf(format string, args ...interface{}) {
	log.WarnDepth(1, r.Fields(), fmt.Sprintf(format, args...))
}

func (r *requestLog) Warnln(args ...interface{}) {
	log.WarnDepth(1, r.Fields(), fmt.Sprintln(args...))
}

func (r *requestLog) DebugEnabled() bool {
	return log.V(int32(LogLevelDebug)).IsEnable()
}

// RequestLogFromContext ...
func RequestLogFromContext(ctx context.Context) *requestLog {
	if rl, ok := ctx.Value(requestIDKey{}).(*requestLog); ok {
		return rl
	}

	rl := &requestLog{
		ID:   GetRequestIDFromContext(ctx),
		Data: make(Fields, 5),
	}

	rl.Data["id"] = rl.ID

	return rl
}

// Marshal ...
func Marshal(v interface{}) string {
	return utils.JsonFormat(v)
}

func JsonMarshal(v interface{}) string {
	dat, err := json.Marshal(v)
	if err == nil {
		return string(dat)
	}

	return fmt.Sprintf("%#v", v)
}

// ClientLogInterceptor ...
func ClientLogInterceptor(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	id := GetRequestIDFromContext(ctx)

	if canDumpGrpcMeta() {
		md, _ := metadata.FromOutgoingContext(ctx)
		log.Infof("[req] *client id:%v method:%v md:%+v", id, method, md)
	}

	if canDumpGrpcBody() {
		log.Infof("[req] *client id:%v method:%v target:%v req:\n%v",
			id, method, cc.Target(), Marshal(req))
	}

	start := time.Now()

	err := invoker(ctx, method, req, reply, cc, opts...)

	printWithData(reply, canDumpGrpcBody(),
		"[rsp] *client id:%v method:%v cost:%v err:%v",
		id, method, time.Since(start), err)

	return err
}

// StreamLogInterceptor log for stream
func StreamServerLogInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if !log.V(int32(LogLevel)).IsEnable() {
		return handler(srv, ss)
	}

	id := getRandomID()

	wrapper := &streamWrapper{
		ServerStream:   ss,
		WrapperContext: context.WithValue(ss.Context(), requestIDKey{}, id),
		ID:             id,
	}

	if canDumpGrpcMeta() {
		md, _ := metadata.FromIncomingContext(ss.Context())
		log.Infof("[stream] [server] id:%v method:%v metadata:md %+v", id, info.FullMethod, md)
	}

	log.Infof("[stream] [server] id:%v method:%v connected",
		wrapper.ID, info.FullMethod)

	st := time.Now()

	err := handler(srv, wrapper)

	log.Infof("[stream] [server] id:%v method:%v closed. cost:%v err:%v",
		wrapper.ID, info.FullMethod, time.Since(st), err)

	return err
}

// StreamLogInterceptor log for stream
func StreamClientLogInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	if !log.V(int32(LogLevel)).IsEnable() {
		return streamer(ctx, desc, cc, method, opts...)
	}

	id := getRandomID()

	if canDumpGrpcMeta() {
		md, _ := metadata.FromOutgoingContext(ctx)
		log.Infof("[stream] [client] id:%v method:%v metadata:md %+v", id, method, md)
	}

	log.V(int32(LogLevel)).Infof("[stream] [client] id:%v method:%v connected",
		id, method)

	return streamer(ctx, desc, cc, method, opts...)
}

func init() {
	flag.Var(&DumpGrpcBody, "dump-grpc-body", "dump grpc request and response body")
	flag.Var(&DumpGrpcMeta, "dump-grpc-meta", "dump grpc meta data")
}
