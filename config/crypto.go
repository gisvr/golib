package config

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"reflect"

	"github.com/gisvr/golib/crypt"
)

const (
	tagName = "encrypted"

	randLen = 12
)

// BasicType .
var BasicType = map[reflect.Kind]bool{
	reflect.Invalid:    true,
	reflect.Bool:       true,
	reflect.Int:        true,
	reflect.Int8:       true,
	reflect.Int16:      true,
	reflect.Int32:      true,
	reflect.Int64:      true,
	reflect.Uint:       true,
	reflect.Uint8:      true,
	reflect.Uint16:     true,
	reflect.Uint32:     true,
	reflect.Uint64:     true,
	reflect.Uintptr:    true,
	reflect.Float32:    true,
	reflect.Float64:    true,
	reflect.Complex64:  true,
	reflect.Complex128: true,
	reflect.Chan:       true,
	reflect.Func:       true,
	reflect.Interface:  true,
	reflect.String:     true,
}

// AdvanceType .
var AdvanceType = map[reflect.Kind]bool{
	reflect.Array:         true,
	reflect.Map:           true,
	reflect.Ptr:           true,
	reflect.Slice:         true,
	reflect.Struct:        true,
	reflect.UnsafePointer: true,
}

var decryptHandler = Decrypt

// DecryptHandlerFunc ..
type DecryptHandlerFunc func(string) (string, error)

var encryptedField = &reflect.StructField{}

func decryptReflectValue(v reflect.Value, field *reflect.StructField, f func(value reflect.Value) error) error {
	encrypted := false

	if field == encryptedField {
		encrypted = true
	} else if field != nil {
		if field.PkgPath != "" {
			return nil
		}

		tag, ok := field.Tag.Lookup(tagName)

		if ok && (tag == "true" || tag == "1") {
			encrypted = true
		}
	}

	switch v.Kind() {
	case reflect.String:
		if encrypted {
			if err := f(v); err != nil {
				return err
			}
		}
		return nil
		// raw string value
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !encrypted {
				fs := v.Type().Field(i)
				field = &fs
			}

			if err := decryptReflectValue(v.Field(i), field, f); err != nil {
				return err
			}
		}

	case reflect.Interface:
		v = v.Elem()
		fallthrough

	case reflect.Ptr:
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			v = v.Elem()

			return decryptReflectValue(v, field, f)
		}

	case reflect.Map:
		if !encrypted {
			if _, ok := BasicType[v.Type().Elem().Kind()]; ok {
				return nil
			}
		}

		for _, k := range v.MapKeys() {
			cv := v.MapIndex(k)

			mapElem := reflect.New(v.Type().Elem()).Elem()
			mapElem.Set(cv)

			nextFiled := encryptedField

			if !encrypted {
				nextFiled = nil
			}

			if err := decryptReflectValue(mapElem, nextFiled, f); err != nil {
				return err
			}
			v.SetMapIndex(k, mapElem)
		}

	case reflect.Slice, reflect.Array:
		if !encrypted {
			if _, ok := BasicType[v.Type().Elem().Kind()]; ok {
				return nil
			}
		}

		for i := 0; i < v.Len(); i++ {
			cv := v.Index(i)

			mapElem := reflect.New(v.Type().Elem()).Elem()
			mapElem.Set(cv)

			nextFiled := encryptedField

			if !encrypted {
				nextFiled = nil
			}

			if err := decryptReflectValue(mapElem, nextFiled, f); err != nil {
				return err
			}
			v.Index(i).Set(mapElem)
		}
	}

	return nil
}

func filterEncryptedWithHandler(v interface{}, h DecryptHandlerFunc) error {
	if h == nil {
		return nil
	}

	f := func(value reflect.Value) error {
		val, err := h(value.String())
		if err == nil {
			value.SetString(val)
		}
		return nil
	}

	return decryptReflectValue(reflect.ValueOf(v), nil, f)
}

func filterEncrypted(v interface{}) error {
	return filterEncryptedWithHandler(v, decryptHandler)
}

// Encrypt 加密
func Encrypt(v string) (string, error) {
	randKey := make([]byte, randLen+2)
	randKey[1] = 18

	if _, err := io.ReadFull(rand.Reader, randKey[2:]); err != nil {
		return "", err
	}

	encryptKey := sha256.Sum256(randKey)
	encrypted, err := crypt.AESGCMEncrypt([]byte(v), encryptKey[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(append(encrypted, randKey[2:]...)), nil
}

// Decrypt 解密
func Decrypt(v string) (string, error) {
	if v == "" {
		return "", nil
	}

	dat, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		log.Printf("decrypt base64 decode error: %v", err)
		return v, nil
	}

	if len(dat) < randLen {
		return v, nil
	}

	randKey := make([]byte, randLen+2)
	randKey[1] = 18
	copy(randKey[2:], dat[len(dat)-randLen:])

	decryptKey := sha256.Sum256(randKey)
	decrypted, err := crypt.AESGCMDecrypt(dat[:len(dat)-randLen], decryptKey[:])
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// KeyGenerator gens key
type KeyGenerator interface {
	Generate([]byte) []byte
}

// KeyGeneratorFunc ..
type KeyGeneratorFunc func([]byte) []byte

// Generate ..
func (f KeyGeneratorFunc) Generate(v []byte) []byte {
	return f(v)
}

// EncryptBytes 通过提供的 generator 加密
func EncryptBytes(v []byte, generator KeyGenerator) ([]byte, error) {
	randKey := make([]byte, randLen)
	if _, err := io.ReadFull(rand.Reader, randKey); err != nil {
		return nil, err
	}

	encryptKey := sha256.Sum256(generator.Generate(randKey))
	encrypted, err := crypt.AESGCMEncrypt([]byte(v), encryptKey[:])
	if err != nil {
		return nil, err
	}

	return append(encrypted, randKey...), nil
}

// DecryptBytes 通过提供的 generator 解密
func DecryptBytes(v []byte, generator KeyGenerator) ([]byte, error) {
	if v == nil {
		return nil, nil
	}

	if len(v) < randLen {
		return nil, errors.New("bad length")
	}

	pos := len(v) - randLen

	decryptKey := sha256.Sum256(generator.Generate(v[pos:]))

	return crypt.AESGCMDecrypt(v[:pos], decryptKey[:])
}
