package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"

	"github.com/pkg/errors"
)

func addBase64PaddingAES(value string) string {
	m := len(value) % 4
	if m != 0 {
		value += strings.Repeat("=", 4-m)
	}

	return value
}

func removeBase64PaddingAES(value string) string {
	return strings.Replace(value, "=", "", -1)
}

func padAES(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func unpadAES(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}

// EncryptAES ...
func EncryptAES(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	msg := padAES([]byte(text))
	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cbe := cipher.NewCBCEncrypter(block, iv)
	cbe.CryptBlocks(ciphertext[aes.BlockSize:], []byte(text))
	finalMsg := removeBase64PaddingAES(base64.URLEncoding.EncodeToString(ciphertext))
	return finalMsg, nil
}

// DecryptAES ...
func DecryptAES(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(addBase64PaddingAES(text))
	if err != nil {
		return "", err
	}

	if (len(decodedMsg) % aes.BlockSize) != 0 {
		return "", errors.New("blocksize must be multipe of decoded message length")
	}

	iv := decodedMsg[:aes.BlockSize]
	msg := decodedMsg[aes.BlockSize:]

	cbe := cipher.NewCBCDecrypter(block, iv)
	cbe.CryptBlocks(msg, msg)

	unpadMsg, err := unpadAES(msg)
	if err != nil {
		return "", err
	}

	return string(unpadMsg), nil
}
