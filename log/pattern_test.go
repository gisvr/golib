package log

import (
	"bytes"
	"github.com/gisvr/golib/log2"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPatternDefault(t *testing.T) {
	buf := &bytes.Buffer{}
	p := newPatternRender("%L %T %f %M")
	p.Render(buf, map[string]interface{}{_level: _infoLevel.String(), _log: "hello", _time: time.Now().Format(_timeFormat), _source: "xxx:123"})

	fields := strings.Fields(buf.String())
	assert.Equal(t, 4, len(fields))
	assert.Equal(t, "INFO", fields[0])
	assert.Equal(t, "hello", fields[3])
}

func TestKV(t *testing.T) {
	buf := &bytes.Buffer{}
	p := newPatternRender("%M")
	p.Render(buf, map[string]interface{}{_level: _infoLevel.String(), _log: "2233", "hello": "test"})
	assert.Equal(t, "hello=test 2233", buf.String())
}

func TestBadSymbol(t *testing.T) {
	buf := &bytes.Buffer{}
	p := newPatternRender("%12 %% %xd %M")
	p.Render(buf, map[string]interface{}{_level: _infoLevel.String(), _log: "2233"})
	assert.Equal(t, "%12 %% %xd 2233", buf.String())
}

func TestSliceRemove(t *testing.T) {
	arr := []string{"A", "B", "C", "D", "E", "F", "G"}

	isIn := func(s string) bool {
		rev := []string{"A", "G"}
		for _, r := range rev {
			if r == s {
				return true
			}
		}
		return false
	}
	length := len(arr)
	for i := length - 1; i >= 0; i-- {
		if isIn(arr[i]) {
			arr = append(arr[:i], arr[i+1:]...)
		}
	}
	log.Infof("%+v", strings.Join(arr, ","))
}
