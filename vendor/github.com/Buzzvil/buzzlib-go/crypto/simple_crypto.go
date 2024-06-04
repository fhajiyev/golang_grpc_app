package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

func Base64Encoding(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

func Base64Decoding(text string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(text)
}

func AesEncrypter(key []byte, plainText string) ([]byte, error) {
	plaintext := []byte(plainText)
	block, err := aes.NewCipher(key)
	if err != nil {
		return plaintext, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	cipherBytes := make([]byte, aes.BlockSize+len(plaintext))
	iv := cipherBytes[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return plaintext, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherBytes[aes.BlockSize:], plaintext)
	return cipherBytes, nil
}

func AesDecrpyter(key []byte, cipheredBytes []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(cipheredBytes) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := cipheredBytes[:aes.BlockSize]
	cipheredBytes = cipheredBytes[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipheredBytes, cipheredBytes)
	return fmt.Sprintf("%s", cipheredBytes), nil
}
