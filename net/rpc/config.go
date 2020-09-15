package rpc

// GrpcOption for gRPC
type GrpcOption struct {
	GrpcAddr           string `yaml:"addr"` // 地址
	EnableJaeger       bool   `yaml:"enable_jaeger"`
	EnableLog          bool   `yaml:"enable_log"`
	EnableDumpGrpcMeta bool   `yaml:"enable_dump_grpc_meta"` // log里面包含metadata，默认关闭
	EnableDumpGrpcBody bool   `yaml:"enable_dump_grpc_body"` // log里包含body信息
	EnablePrometheus   bool   `yaml:"enable_prometheus"`
	MaxRecvMsgSize     int    `yaml:"max_recv_msg_size"`  // 如果不写默认10M
	MaxSendMsgSize     int    `yaml:"max_send_msg_size"`  // 如果不写默认10M
	KeepAliveTime      int    `yaml:"keep_alive_time"`    // 单位秒，如果不写默认20秒
	KeepAliveTimeout   int    `yaml:"keep_alive_timeout"` // 单位秒，如果不写默认10秒
}
