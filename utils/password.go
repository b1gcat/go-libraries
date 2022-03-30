package utils

import (
	"encoding/base32"
	"golang.org/x/crypto/bcrypt"
)

func GenStorePassword(password string) string {
	ePass, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(ePass)
}

func ComparePassword(ePass string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(ePass), []byte(password))
	return err == nil
}

func GenCryptPassword(password string, key []byte) string {
	return Encode(password, key)
}

func DCryptPassword(ePass string, key []byte) ([]byte, error) {
	return Decode(ePass, key)
}

func HasEncrypt(pass string) bool {
	_, err := base32.StdEncoding.DecodeString(pass)
	if err != nil {
		return false
	}
	return true
}
