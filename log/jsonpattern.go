package log

import (
	"fmt"
	"github.com/gisvr/golib/log/internal/core"
	"io"
	"path"
	"time"
)

type jsonFuncMap struct {
	pattern  string
	function func(core.Encoder, map[string]interface{})
}

var jsonPatternMap = map[string]jsonFuncMap{
	"T": {_time, longTime2Codec},
	"t": {_time, shortTime2Codec},
	"L": {_level, keyFactory2Codec(_level)},
	"f": {_bizName, keyFactory2Codec(_bizName)},
	"P": {_svcName, keyFactory2Codec(_svcName)},
	"p": {_pid, keyFactory2Codec(_pid)},
	"u": {_userId, keyFactory2Codec(_userId)},
	"h": {_hostName, keyFactory2Codec(_hostName)},
	"S": {_source, longSource2Codec},
	"s": {_source, shortSource2Codec},
	"M": {_log, message2Codec},
}

// newPatternRender new pattern render
func newJsonPatternRender(format string) Render {
	p := &jsonpattern{
		funcs: make([]jsonFuncMap, 0),
	}
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			continue
		}
		if i+1 >= len(format) {
			continue
		}
		f, ok := jsonPatternMap[string(format[i+1])]
		if !ok {
			continue
		}
		p.funcs = append(p.funcs, f)
		i++
	}
	return p
}

type jsonpattern struct {
	funcs []jsonFuncMap
}

func (p *jsonpattern) format(d map[string]interface{}) []byte {
	enc := core.NewJSONEncoder(
		core.EncoderConfig{
			EncodeTime:     core.EpochTimeEncoder,
			EncodeDuration: core.SecondsDurationEncoder,
		},
		core.GetPool(),
	)

	for _, f := range p.funcs {
		f.function(enc, d)
	}
	buf := core.GetPool()
	enc.Encode(buf)
	defer buf.Free()
	return buf.Bytes()
}

// Render implemet Formater
func (p *jsonpattern) Render(w io.Writer, d map[string]interface{}) error {
	buf := p.format(d)
	_, err := w.Write(buf)
	return err
}

// Render implemet Formater as string
func (p *jsonpattern) RenderString(d map[string]interface{}) string {
	buf := p.format(d)
	return string(buf)
}

func message2Codec(enc core.Encoder, md map[string]interface{}) {
	for key, val := range md {
		if isInternalKey(key) {
			continue
		}
		KV(key, val).AddTo(enc)
	}
}
func keyFactory2Codec(key string) func(core.Encoder, map[string]interface{}) {
	return func(enc core.Encoder, md map[string]interface{}) {
		if v, ok := md[key]; ok {
			if s, ok := v.(string); ok {
				KVString(key, s).AddTo(enc)
			} else {
				KVString(key, fmt.Sprint(v)).AddTo(enc)
			}
		} else {
			KVString(key, "").AddTo(enc)
		}
	}
}

func longSource2Codec(enc core.Encoder, d map[string]interface{}) {
	fn, ok := d[_fileName]
	if ok {
		KVString(_fileName, fn.(string)).AddTo(enc)
	} else {
		KVString(_fileName, "unknown").AddTo(enc)
	}
	lno, ok := d[_line]
	if ok {
		KVInt64(_line, lno.(int64)).AddTo(enc)
	} else {
		KVInt64(_line, 0).AddTo(enc)
	}
}

func shortSource2Codec(enc core.Encoder, d map[string]interface{}) {
	fn, ok := d[_fileName]
	if ok {
		KVString(_fileName, path.Base(fn.(string))).AddTo(enc)
	} else {
		KVString(_fileName, "unknown").AddTo(enc)
	}
	lno, ok := d[_line]
	if ok {
		KVInt64(_line, lno.(int64)).AddTo(enc)
	} else {
		KVInt64(_line, 0).AddTo(enc)
	}
}

func longTime2Codec(enc core.Encoder, d map[string]interface{}) {
	KVString(_time, time.Now().Format(_timeFormat)).AddTo(enc)
}

func shortTime2Codec(enc core.Encoder, d map[string]interface{}) {
	KVString(_time, time.Now().Format("2006-01-02T15:04:05")).AddTo(enc)
}

func longDate2Codec(enc core.Encoder, d map[string]interface{}) {
	KVString(_time, time.Now().Format("2006-01-02")).AddTo(enc)
}

func shortDate2Codec(enc core.Encoder, d map[string]interface{}) {
	KVString(_time, time.Now().Format("01/02")).AddTo(enc)
}
