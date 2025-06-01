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
func Pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nerror.NewRunTimeError("空数据")
	}
	// 最后一个字节是填充长度
	padding := int(data[len(data)-1])
	if padding < 1 || padding > len(data) {
		return nil, nerror.NewRunTimeError("Pkcs7填充长度无效")
	}
	return data[:len(data)-padding], nil
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	if length == 0 {
		return origData // 空数据直接返回
	}
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// PKCS5Padding 对数据填充至8字节倍数
func PKCS5Padding(ciphertext []byte) []byte {
	return Pkcs7Pad(ciphertext, 16) // 调用PKCS7并指定块大小8
}
