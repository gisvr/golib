package rpc

import (
	"google.golang.org/grpc"
	"sync"
)

var (
	_register *rpcRegister
	_once     sync.Once
)

type RpcRegisterFunc func(s *grpc.Server)

type rpcRegister struct {
	regfuncs []RpcRegisterFunc
}

func (s *rpcRegister) Add(reg ...RpcRegisterFunc) {
	s.regfuncs = append(s.regfuncs, reg...)
}

func (s *rpcRegister) Regist(svr *grpc.Server) {
	for _, rf := range s.regfuncs {
		rf(svr)
	}
}

func Register() *rpcRegister {
	_once.Do(func() {
		_register = &rpcRegister{}
		_register.regfuncs = make([]RpcRegisterFunc, 0)
	})

	return _register
}
