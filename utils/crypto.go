package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base32"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

// PKCS7Padding say ...
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

// PKCS7UnPadding 使用PKCS7进行填充 复原
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	if length == 0 {
		return origData
	}
	unPadding := int(origData[length-1])
	if unPadding > length {
		return origData
	}
	return origData[:(length - unPadding)]
}

func Decode(eid string, key []byte) ([]byte, error) {
	c, _ := aes.NewCipher(key)
	//去掉编码
	pEid, err := base32.StdEncoding.DecodeString(eid)
	if err != nil {
		return nil, err
	}
	if len(pEid)%c.BlockSize() != 0 {
		return nil, fmt.Errorf("bad data length %v %v", len(pEid), len(pEid)/c.BlockSize())
	}
	id := make([]byte, len(pEid))
	//解密得到hid
	mode := cipher.NewCBCDecrypter(c, key[:c.BlockSize()])
	mode.CryptBlocks(id, pEid)
	id = PKCS7UnPadding(id)
	return id, nil
}

// Encode base32( aes(hid) )
func Encode(id string, key []byte) string {
	c, _ := aes.NewCipher(key)
	pHid := PKCS7Padding([]byte(id), c.BlockSize())
	eid := make([]byte, len(pHid))
	//加密hid
	mode := cipher.NewCBCEncrypter(c, key[:c.BlockSize()])
	mode.CryptBlocks(eid, pHid)
	//编码eid
	return base32.StdEncoding.EncodeToString(eid)
}

func PublicEncrypt(publicKey []byte, data []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)

	partLen := pub.N.BitLen()/8 - 11
	chunks := split(data, partLen)
	buffer := bytes.NewBufferString("")
	for _, chunk := range chunks {
		bytes, err := rsa.EncryptPKCS1v15(rand.Reader, pub, chunk)
		if err != nil {
			return nil, err
		}
		buffer.Write(bytes)
	}
	return buffer.Bytes(), nil
}

// 私钥解密
func PrivateDecrypt(privateKey []byte, encrypted string) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, fmt.Errorf("private key error")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	partLen := priv.PublicKey.N.BitLen() / 8
	raw, err := base64.StdEncoding.DecodeString(encrypted)
	chunks := split([]byte(raw), partLen)
	buffer := bytes.NewBufferString("")
	for _, chunk := range chunks {
		decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, priv, chunk)
		if err != nil {
			return nil, err
		}
		buffer.Write(decrypted)
	}
	return buffer.Bytes(), err
}

func split(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:])
	}
	return chunks
}
