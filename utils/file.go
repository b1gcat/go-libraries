package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

//RecoverFiles 导入配置时，生成new配置和bak配置。
func RecoverFiles(newFile, orgFile string) error {
	if _, err := os.Stat(newFile); err != nil {
		return fmt.Errorf("配置更新:无更新")
	}
	bakFile := orgFile + ".bak"
	//先备份
	if err := CopyFiles(orgFile, bakFile); err != nil {
		return fmt.Errorf("备份时发生异常, 停止更新配置:" + err.Error())
	}
	//删掉被更新的文件
	_ = os.Remove(orgFile)
	//还原新文件
	if err := CopyFiles(newFile, orgFile); err != nil {
		//还原失败，则恢复文件
		_ = os.Remove(orgFile)
		if err = CopyFiles(bakFile, orgFile); err != nil {
			return fmt.Errorf("恢复配置失败:" + err.Error())
		}
		return fmt.Errorf("复制新文件时失败:" + err.Error())
	}
	_ = os.Remove(newFile)
	return nil
}

func CopyFiles(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("文件类型不支持")
	}
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func FileEncrypt(src, dst string, key []byte) error {
	inFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inFile.Close()
	outFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer outFile.Close()

	iv := make([]byte, aes.BlockSize)
	_, err = rand.Read(iv)
	if err != nil {
		return err
	}

	var block cipher.Block
	block, _ = aes.NewCipher(key)
	stream := cipher.NewCTR(block, iv[:])

	outFile.Write(iv)
	writer := &cipher.StreamWriter{S: stream, W: outFile}

	if _, err = io.Copy(writer, inFile); err != nil {
		return err
	}
	return nil
}

func FileDEncrypt(src, dst string, key []byte) error {
	inFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inFile.Close()
	outFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}

	defer outFile.Close()

	iv := make([]byte, aes.BlockSize)
	io.ReadFull(inFile, iv[:])

	var block cipher.Block
	block, _ = aes.NewCipher(key)
	stream := cipher.NewCTR(block, iv[:])
	inFile.Seek(aes.BlockSize, 0) // Read after the IV

	reader := &cipher.StreamReader{S: stream, R: inFile}
	if _, err = io.Copy(outFile, reader); err != nil {
		return err
	}
	return nil
}

func LoadListFromFile(filename string) ([]string, error) {
	d0, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("LoadListFromFile.ReadFile:%v", err)
	}
	d1 := strings.ReplaceAll(string(d0), "\r", "")
	return strings.Split(d1, "\n"), nil
}

func WriteToFile(filename string, a ...interface{}) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		logrus.Error("writeTo:", err.Error())
		return
	}
	defer f.Close()
	fmt.Fprintln(f, a...)
}
