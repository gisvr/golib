package codec

import (
	"bytes"
	"encoding/gob"

	"github.com/gisvr/golib/log"
)

// go binary encoder
func ToGOB64(m interface{}) (string, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(m)
	if err != nil {
		log.Errorf(`failed gob Encode, `, err)
		return ``, err
	}
	result := string(b.Bytes())
	return result, nil
}

// go binary decoder
func FromGOB64(str string, obj interface{}) error {
	b := bytes.Buffer{}
	b.Write([]byte(str))
	d := gob.NewDecoder(&b)
	err := d.Decode(obj)
	if err != nil {
		log.Errorf(`failed gob Decode, `, err)
		return err
	}
	return nil
}
