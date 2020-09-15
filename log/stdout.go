package log

import (
	"context"
	"os"
	"time"
)

const defaultStdoutPattern = "%T %L %s\t] %M"

var _defaultStdout = NewStdout(&StdoutOption{Fmt: defaultStdoutPattern})

// StdoutHandler stdout log handler
type StdoutHandler struct {
	opt    *StdoutOption
	render Render
}

// NewStdout create a stdout log handler
func NewStdout(opt *StdoutOption) *StdoutHandler {
	if opt.Fmt == "" {
		opt.Fmt = defaultStdoutPattern
	}
	h := &StdoutHandler{opt: opt}
	if opt.Json {
		h.render = newJsonPatternRender(opt.Fmt)
	} else {
		h.render = newPatternRender(opt.Fmt)
	}
	return h
}

func (h *StdoutHandler) filter(args ...D) []D {
	isIn := func(v string) bool {
		for _, fin := range h.opt.FilterIn {
			if fin == v {
				return true
			}
		}
		for _, fout := range h.opt.FilterOut {
			if fout == v {
				return false
			}
		}
		//只有filterin
		if h.opt.FilterIn != nil && h.opt.FilterOut == nil {
			return false
		}
		//只有filterout
		if h.opt.FilterIn == nil && h.opt.FilterOut != nil {
			return true
		}
		//filterin和filterout都没设
		if h.opt.FilterIn == nil && h.opt.FilterOut == nil {
			return true
		}
		//filterin和filterout都设置了
		return false
	}
	for i := len(args) - 1; i >= 0; i-- {
		if isInternalKey(args[i].Key) {
			continue
		}
		if !isIn(args[i].Key) {
			args = append(args[:i], args[i+1:]...)
		}
	}
	return args
}

// Log stdout loging, only for developing env.
func (h *StdoutHandler) Log(ctx context.Context, depth int, lv Level, args ...D) {
	args = h.filter(args...)
	if len(args) <= 5 {
		//日志内容都被过滤了
		return
	}
	d := toMap(args...)
	// add extra fields
	addExtraField(ctx, d)
	d[_time] = time.Now().Format(_timeFormat)

	h.render.Render(os.Stderr, d)
}

// Close stdout loging
func (h *StdoutHandler) Close() error {
	return nil
}

// SetFormat set stdout log output format
// %T time format at "15:04:05.999"
// %t time format at "15:04:05"
// %D data format at "2006/01/02"
// %d data format at "01/02"
// %L log level e.g. INFO WARN ERROR
// %f function name and line number e.g. model.Get:121
// %i instance id
// %e deploy env e.g. dev uat fat prod
// %z zone
// %S full file name and line number: /a/b/c/d.go:23
// %s final file name element and line number: d.go:23
// %M log message and additional fields: key=value this is log message
func (h *StdoutHandler) SetFormat(format string) {
	h.render = newPatternRender(format)
}
