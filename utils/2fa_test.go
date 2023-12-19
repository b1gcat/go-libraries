package utils

import (
	"encoding/base32"
	"fmt"
	"testing"
	"time"
)

func Test_2fa(t *testing.T) {
	secret := "3w3QVQu9~M&xx9afbdK#"
	key := base32.StdEncoding.EncodeToString([]byte(secret))
	fmt.Println("key:", key)
	//key = "B3UY2M7YIWNHNFNKZD6HWRRLLUCGWXH5W4OPJ6OZ5P5NCX6ZOMIA===="
	fmt.Println(ComputeCode(key, time.Now().Unix()/30))
	otp := OTPConfig{
		Secret:      key,
		WindowSize:  3,
		HotpCounter: 0,
	}
	fmt.Println(otp.ProvisionURIWithIssuer("test", ""))
}
