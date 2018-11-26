package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

var (
	key = []byte("maxiiot@humiture")
)

func AesEncrypt(srcData []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	srcData = PKCS5Padding(srcData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(srcData))
	blockMode.CryptBlocks(crypted, srcData)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

func AesDecrypt(str string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	crypted, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	srcData := make([]byte, len(crypted))
	blockMode.CryptBlocks(srcData, crypted)
	srcData = PKCS5Unpadding(srcData)
	return srcData, nil
}

func PKCS5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func PKCS5Unpadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:length-unpadding]
}
