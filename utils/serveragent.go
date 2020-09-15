package utils

import (
	"context"
	"github.com/gisvr/golib/log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

//SignalContext ..
func SignalContext(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Infof("listening for shutdown signal")
		<-sigs
		log.Infof("shutdown signal received")
		signal.Stop(sigs)
		close(sigs)
		cancel()
	}()

	return ctx
}

type AbstractServer interface {
	Run(ctx context.Context) error
}

type ServerAgent struct {
	ctx context.Context
	wg  *sync.WaitGroup
}

func NewServerAgent() *ServerAgent {
	return &ServerAgent{
		ctx: SignalContext(context.Background()),
		wg:  &sync.WaitGroup{},
	}
}

func (sh *ServerAgent) RunServer(s AbstractServer) {
	sh.wg.Add(1)
	go func() {
		defer sh.wg.Done()
		if err := s.Run(sh.ctx); err != nil {
			log.Fatalf("runServer error:%v", err)
		}
	}()
}

func (sh *ServerAgent) Wait() {
	sh.wg.Wait()
}
