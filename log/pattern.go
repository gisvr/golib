package log

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

var patternMap = map[string]func(map[string]interface{}) string{
	"T": longTime,
	"t": shortTime,
	"L": keyFactory(_level),
	"f": keyFactory(_bizName),
	"P": keyFactory(_svcName),
	"p": keyFactory(_pid),
	"u": keyFactory(_userId),
	"h": keyFactory(_hostName),
	"S": longSource,
	"s": shortSource,
	"M": message,
}

// newPatternRender new pattern render
func newPatternRender(format string) Render {
	p := &pattern{
		bufPool: sync.Pool{New: func() interface{} { return &bytes.Buffer{} }},
	}
	b := make([]byte, 0, len(format))
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			b = append(b, format[i])
			continue
		}
		if i+1 >= len(format) {
			b = append(b, format[i])
			continue
		}
		f, ok := patternMap[string(format[i+1])]
		if !ok {
			b = append(b, format[i])
			continue
		}
		if len(b) != 0 {
			p.funcs = append(p.funcs, textFactory(string(b)))
			b = b[:0]
		}
		p.funcs = append(p.funcs, f)
		i++
	}
	if len(b) != 0 {
		p.funcs = append(p.funcs, textFactory(string(b)))
	}
	return p
}

type pattern struct {
	funcs   []func(map[string]interface{}) string
	bufPool sync.Pool
}

// Render implemet Formater
func (p *pattern) Render(w io.Writer, d map[string]interface{}) error {
	buf := p.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufPool.Put(buf)
	}()
	for _, f := range p.funcs {
		buf.WriteString(f(d))
	}
	buf.WriteString("\n")
	_, err := buf.WriteTo(w)
	return err
}

// Render implemet Formater as string
func (p *pattern) RenderString(d map[string]interface{}) string {
	// TODO strings.Builder
	buf := p.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufPool.Put(buf)
	}()
	for _, f := range p.funcs {
		buf.WriteString(f(d))
	}

	return buf.String()
}

func textFactory(text string) func(map[string]interface{}) string {
	return func(map[string]interface{}) string {
		return text
	}
}
func keyFactory(key string) func(map[string]interface{}) string {
	return func(d map[string]interface{}) string {
		if v, ok := d[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
			return fmt.Sprint(v)
		}
		return ""
	}
}

func longSource(d map[string]interface{}) string {
	lno, ok := d[_line]
	if !ok {
		return "unknown:0"
	}
	fn, ok := d[_fileName]
	if !ok {
		return "unknown:0"
	}
	return fn.(string) + ":" + strconv.FormatInt(lno.(int64), 10)
}

func shortSource(d map[string]interface{}) string {
	lno, ok := d[_line]
	if !ok {
		return "unknown:0"
	}
	fn, ok := d[_fileName]
	if !ok {
		return "unknown:0"
	}
	return path.Base(fn.(string)) + ":" + strconv.FormatInt(lno.(int64), 10)
}

func longTime(map[string]interface{}) string {
	return time.Now().Format(_timeFormat)
}

func shortTime(map[string]interface{}) string {
	return time.Now().Format("2006-01-02T15:04:05")
}

func longDate(map[string]interface{}) string {
	return time.Now().Format("2006-01-02")
}

func shortDate(map[string]interface{}) string {
	return time.Now().Format("01/02")
}

func isInternalKey(k string) bool {
	switch k {
	case _level, _levelValue, _time,
		_source, _instanceID, _appID,
		_deplyEnv, _zone, _fileName,
		_line, _pid, _threadid,
		_svcName, _bizName, _userId,
		_hostName:
		return true
	}
	return false
}

func message(d map[string]interface{}) string {
	var m string
	var s []string
	for k, v := range d {
		if k == _log {
			m = fmt.Sprint(v)
			continue
		}
		if isInternalKey(k) {
			continue
		}
		s = append(s, fmt.Sprintf("%s=%v", k, v))
	}
	s = append(s, m)
	return strings.Join(s, " ")
}
