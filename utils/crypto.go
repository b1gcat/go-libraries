package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base32"
	"fmt"
)

//PKCS7Padding say ...
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

//PKCS7UnPadding 使用PKCS7进行填充 复原
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

//Encode base32( aes(hid) )
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
