package utils

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	_ "github.com/gisvr/golib/encoding"
)

func GetFuncName(depth int) (funcname string) {
	pc, _, _, ok := runtime.Caller(depth)
	if ok {
		funcname = runtime.FuncForPC(pc).Name()      // main.(*MyStruct).foo
		funcname = filepath.Ext(funcname)            // .foo
		funcname = strings.TrimPrefix(funcname, ".") // foo
	}
	return
}

func JsonFormat(v interface{}) string {
	buf, _ := json.MarshalIndent(v, "", "\t")
	return string(buf)
}

func ContextWithSigterm() context.Context {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		defer cancel()
		sig := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
		c := make(chan os.Signal, 1)
		signal.Notify(c, sig...)
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx
}
