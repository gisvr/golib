package log

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gisvr/golib/log/internal/core"
)

func addExtraField(ctx context.Context, fields map[string]interface{}) {
	fields[_svcName] = processName()
	fields[_bizName] = c.Family
	fields[_pid] = os.Getpid()
	fields[_hostName] = c.Host
}

func processName() string {
	app := os.Args[0]
	app = filepath.Base(app)
	app = strings.TrimSuffix(app, ".exe")
	return app
}

// funcName get func name.
func funcFileLine(skip int) (string, int) {
	if _, file, lineNo, ok := runtime.Caller(skip); ok {
		return file, lineNo
	}
	return "unknown", 0
}

// toMap convert D slice to map[string]interface{} for legacy file and stdout.
func toMap(args ...D) map[string]interface{} {
	d := make(map[string]interface{}, 10+len(args))
	for _, arg := range args {
		switch arg.Type {
		case core.UintType, core.Uint64Type, core.IntTpye, core.Int64Type:
			d[arg.Key] = arg.Int64Val
		case core.StringType:
			d[arg.Key] = arg.StringVal
		case core.Float32Type:
			d[arg.Key] = math.Float32frombits(uint32(arg.Int64Val))
		case core.Float64Type:
			d[arg.Key] = math.Float64frombits(uint64(arg.Int64Val))
		case core.DurationType:
			d[arg.Key] = time.Duration(arg.Int64Val)
		default:
			d[arg.Key] = arg.Value
		}
	}
	return d
}
