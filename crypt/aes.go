package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

const (
	nonceLen = 12
)

// AESGCMEncryptString AES GCM 加密
func AESGCMEncryptString(text string, key []byte) (string, error) {
	dat, err := AESGCMEncrypt([]byte(text), key)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(dat), nil
}

// AESGCMEncrypt AES GCM 加密
func AESGCMEncrypt(text, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	mode, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	cipherText := mode.Seal(nil, nonce, text, nil)

	return append(cipherText, nonce...), nil
}

// AESGCMDecrypt AES GCM 解密
func AESGCMDecrypt(cipherText, key []byte) ([]byte, error) {
	if len(cipherText) < nonceLen {
		return nil, errors.New("length not satisfied")
	}

	nonce := cipherText[len(cipherText)-nonceLen:]
	cipherText = cipherText[:len(cipherText)-nonceLen]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	mode, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plainText, err := mode.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}

// AESGCMDecryptString AES GCM 解密
func AESGCMDecryptString(cipherString string, key []byte) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(cipherString)
	if err != nil {
		return "", err
	}

	dat, err := AESGCMDecrypt(cipherText, key)
	if err != nil {
		return "", err
	}

	return string(dat), nil
}
