package log

import (
	"context"
	"io"
	"path/filepath"
	"time"

	"github.com/gisvr/golib/log/internal/filewriter"
)

// level idx
const (
	_infoIdx = iota
	_warnIdx
	_errorIdx
	_totalIdx
)
const defaultFilePattern = "%T %L %s\t] %M"

var _fileNames = map[int]string{
	_infoIdx:  "info.log",
	_warnIdx:  "warning.log",
	_errorIdx: "error.log",
}

// FileHandler .
type FileHandler struct {
	opt    *FileRotateOption
	render Render
	fws    [_totalIdx]*filewriter.FileWriter
}

// NewFile crete a file logger.
func NewFile(opt *FileRotateOption) *FileHandler {

	// new info writer
	newWriter := func(name string, opt *FileRotateOption) *filewriter.FileWriter {
		var options []filewriter.Option
		if opt.RotateSize > 0 {
			options = append(options, filewriter.MaxSize(opt.RotateSize))
		}
		if opt.MaxLogFile > 0 {
			options = append(options, filewriter.MaxFile(opt.MaxLogFile))
		}
		w, err := filewriter.New(filepath.Join(opt.Dir, name), options...)
		if err != nil {
			panic(err)
		}
		return w
	}
	if opt.Fmt == "" {
		opt.Fmt = defaultFilePattern
	}
	h := &FileHandler{opt: opt}
	if opt.Json {
		h.render = newJsonPatternRender(opt.Fmt)
	} else {
		h.render = newPatternRender(opt.Fmt)
	}
	filename := processName()
	if c.Family != "" {
		filename = filename + "_" + c.Family
	}
	if opt.SplitLevelOutput {
		for idx, name := range _fileNames {
			h.fws[idx] = newWriter(filename+"."+name, opt)
		}
	} else {
		h.fws[0] = newWriter(filename+".log", opt)
	}
	return h
}

// Log loggint to file .
func (h *FileHandler) Log(ctx context.Context, depth int, lv Level, args ...D) {
	d := toMap(args...)
	// add extra fields
	addExtraField(ctx, d)
	d[_time] = time.Now().Format(_timeFormat)
	var w io.Writer
	if h.opt.SplitLevelOutput {
		switch lv {
		case _warnLevel:
			w = h.fws[_warnIdx]
		case _errorLevel:
			w = h.fws[_errorIdx]
		default:
			w = h.fws[_infoIdx]
		}
	} else {
		w = h.fws[0]
	}
	h.render.Render(w, d)
}

// Close log handler
func (h *FileHandler) Close() error {
	for _, fw := range h.fws {
		// ignored error
		fw.Close()
	}
	return nil
}

// SetFormat set log format
func (h *FileHandler) SetFormat(format string) {
	h.render = newPatternRender(format)
}
