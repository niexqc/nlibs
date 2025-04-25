package ndnen

import (
	"bytes"

	"github.com/niexqc/nlibs/nerror"
)

// PKCS#7 填充
func Pkcs7Pad(data []byte, blockSize int) []byte {
	// 计算需要填充的字节数
	padding := blockSize - (len(data) % blockSize)
	// 每个填充字节的值等于填充长度
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// PKCS#7 去除填充
func Pkcs7Unpad(data []byte) []byte {
	if len(data) == 0 {
		panic(nerror.NewRunTimeError("空数据"))
	}
	// 最后一个字节是填充长度
	padding := int(data[len(data)-1])
	if padding < 1 || padding > len(data) {
		panic(nerror.NewRunTimeError("Pkcs7填充长度无效"))
	}
	return data[:len(data)-padding]
}
