package utils

import (
	"encoding/base32"
	"fmt"
	"testing"
	"time"
)

func Test_2fa(t *testing.T) {
	secret := "3t3AVQu7~M&ZiPafbhK#"
	key := base32.StdEncoding.EncodeToString([]byte(secret))
	fmt.Println("key:", key)
	fmt.Println(ComputeCode(key, time.Now().Unix()/30))
	otp := OTPConfig{
		Secret:      key,
		WindowSize:  3,
		HotpCounter: 0,
	}
	fmt.Println(otp.ProvisionURIWithIssuer("test", ""))
}
