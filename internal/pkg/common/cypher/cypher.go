package cypher

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/Buzzvil/buzzlib-go/core"
)

// EncryptAesBase64Dict func definition
func EncryptAesBase64Dict(data interface{}, key string, iv string, urlSafe bool) string {
	text, err := json.Marshal(data)
	if err != nil {
		core.Logger.WithError(err).Warnf("EncryptAesDict() - Failed to marshal")
	}
	encryptedData, err := EncryptAesWithBase64([]byte(key), []byte(iv), text, urlSafe)
	//core.Logger.Debugf("EncryptAesBase64Dict() - data: %v, text: %s, encrypted: %s", data, text, encryptedData)
	return encryptedData
}

// DecryptAesBase64Dict func definition
func DecryptAesBase64Dict(data string, key string, iv string, urlSafe bool) map[string]interface{} {
	result := make(map[string]interface{})
	decrypted, err := DecryptAesWithBase64([]byte(key), []byte(iv), data, urlSafe)
	if err == nil {
		json.Unmarshal(decrypted, &result)
	}
	return result
}

// EncryptAesWithBase64 func definition
func EncryptAesWithBase64(key, iv, data []byte, urlSafe bool) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		core.Logger.WithError(err).Warnf("EncryptAesWithBase64() - Failed w/ block")
		return "", err
	}

	padText := pad(data)

	cipherText := make([]byte, len(padText))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(cipherText, padText)
	//fmt.Printf("encryptAes() - cipherText: %v\n", cipherText)
	if urlSafe == true {
		return addBase64Padding(base64.RawURLEncoding.EncodeToString(cipherText)), nil
	}
	return addBase64Padding(base64.StdEncoding.EncodeToString(cipherText)), nil
}

// DecryptAesWithBase64 func definition
func DecryptAesWithBase64(key, iv []byte, text string, urlSafe bool) ([]byte, error) {
	var decodedBytes []byte
	var err error
	if urlSafe {
		decodedBytes, err = base64.RawURLEncoding.DecodeString(removeBase64Padding(text))
	} else {
		decodedBytes, err = base64.StdEncoding.DecodeString(removeBase64Padding(text))
	}

	if err != nil {
		core.Logger.WithError(err).Warnf("DecryptAesWithBase64() - Failed w/ base64")
		return decodedBytes, err
	}

	block, err := aes.NewCipher(key)
	decrpytedBytes := make([]byte, len(decodedBytes))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(decrpytedBytes, decodedBytes)
	decrpytedBytes, err = unpad(decrpytedBytes)

	if err != nil {
		core.Logger.WithError(err).Warnf("DecryptAesWithBase64() - Failed w/ block")
		return decodedBytes, err
	}

	return decrpytedBytes, err
}

func pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}

func addBase64Padding(value string) string {
	m := len(value) % 4
	if m != 0 {
		value += strings.Repeat("=", 4-m)
	}

	return value
}

func removeBase64Padding(value string) string {
	return strings.Replace(value, "=", "", -1)
}
