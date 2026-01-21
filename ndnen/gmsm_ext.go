package ndnen

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"math/big"

	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/ntools"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/sm3"
	"github.com/tjfoc/gmsm/sm4"
	"github.com/tjfoc/gmsm/x509"
)

type sm2Signature struct {
	R, S *big.Int
}

// SM2签名默认
var signUserId = []byte{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38}

// 前后端尽量使用 HEX格式的秘钥
func Sm2GenKeyPair() (pri *sm2.PrivateKey, pub *sm2.PublicKey) {
	privateKey, _ := sm2.GenerateKey(rand.Reader)
	return privateKey, &privateKey.PublicKey
}

func Sm2Key2Hex(pri *sm2.PrivateKey, pub *sm2.PublicKey) (priDerb64, pubDerb64 string) {
	privateKeyBytes := pri.D.FillBytes(make([]byte, 32))
	privateKeyHex := hex.EncodeToString(privateKeyBytes)
	// 获取公钥的 X 和 Y 坐标（各 32 字节）
	xBytes := pub.X.FillBytes(make([]byte, 32))
	yBytes := pub.Y.FillBytes(make([]byte, 32))

	publicKeyBytes := append(append([]byte{0x04}, xBytes...), yBytes...)
	publicKeyHex := hex.EncodeToString(publicKeyBytes)
	return privateKeyHex, publicKeyHex
}

func Sm2LoadPriKeyFromHex(priHex string) (*sm2.PrivateKey, error) {
	privateKeyBytes, err := hex.DecodeString(priHex)
	if err != nil {
		return nil, err
	}
	// 将字节转换为 big.Int（私钥的 D 值）
	d := new(big.Int).SetBytes(privateKeyBytes)

	// 构造 PrivateKey 对象
	privateKey := &sm2.PrivateKey{
		PublicKey: sm2.PublicKey{
			Curve: sm2.P256Sm2(), // 使用 SM2 曲线
		},
		D: d,
	}

	// 计算公钥坐标（X,Y）
	privateKey.PublicKey.X, privateKey.PublicKey.Y = sm2.P256Sm2().ScalarBaseMult(d.Bytes())
	return privateKey, nil
}

func Sm2LoadPubKeyFromHex(pubHex string) (*sm2.PublicKey, error) {
	// Hex 解码
	publicKeyBytes, err := hex.DecodeString(pubHex)
	if err != nil {
		return nil, err
	}
	// 检查格式是否为未压缩公钥（0x04 开头）
	if len(publicKeyBytes) != 65 || publicKeyBytes[0] != 0x04 {
		return nil, nerror.NewRunTimeError("invalid public key format")
	}
	// 提取 X 和 Y 坐标（各 32 字节）
	x := new(big.Int).SetBytes(publicKeyBytes[1:33])
	y := new(big.Int).SetBytes(publicKeyBytes[33:65])

	// 构造 PublicKey 对象
	publicKey := &sm2.PublicKey{
		Curve: sm2.P256Sm2(),
		X:     x,
		Y:     y,
	}
	return publicKey, nil
}

func Sm2Key2Der2B64(pri *sm2.PrivateKey, pub *sm2.PublicKey) (priDerb64, pubDerb64 string) {
	// 编码私钥为DER格式
	derPrivate, _ := x509.MarshalSm2UnecryptedPrivateKey(pri)
	priDerb64 = base64.StdEncoding.EncodeToString(derPrivate)
	// 编码公钥为DER格式
	derPublic, _ := x509.MarshalSm2PublicKey(pub)
	pubDerb64 = base64.StdEncoding.EncodeToString(derPublic)
	return priDerb64, pubDerb64
}

func Sm2Key2Pem(pri *sm2.PrivateKey, pub *sm2.PublicKey) (priPemStr, pubPemStr string) {
	// 编码私钥为DER格式
	derPrivate, _ := x509.MarshalSm2UnecryptedPrivateKey(pri)
	block := &pem.Block{Type: "PRIVATE KEY", Bytes: derPrivate}
	privatePEM := pem.EncodeToMemory(block)

	// 编码公钥为DER格式
	derPublic, _ := x509.MarshalSm2PublicKey(pub)
	blockPub := &pem.Block{Type: "PUBLIC KEY", Bytes: derPublic}
	publicPEM := pem.EncodeToMemory(blockPub)

	return string(privatePEM), string(publicPEM)
}

func Sm2LoadPubKeyFromDerB64(pubDerb64 string) (*sm2.PublicKey, error) {
	derPublic, err := base64.StdEncoding.DecodeString(pubDerb64)
	if err != nil {
		return nil, err
	}
	publicKey, err := x509.ParseSm2PublicKey(derPublic)
	if err != nil {
		return nil, err
	}
	return publicKey, nil
}

func Sm2LoadPriKeyFromDerB64(priDerb64 string) (*sm2.PrivateKey, error) {
	derPrivate, err := base64.StdEncoding.DecodeString(priDerb64)
	if err != nil {
		return nil, err
	}
	privateKey, err := x509.ParsePKCS8UnecryptedPrivateKey(derPrivate)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func Sm2LoadPubKeyFromPem(pubPemStr string) (*sm2.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPemStr))
	if block == nil {
		return nil, nerror.NewRunTimeError("PEM 解码失败")
	}
	privateKey, err := x509.ParseSm2PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func Sm2LoadPriKeyFromPem(priPemStr string) (*sm2.PrivateKey, error) {
	block, _ := pem.Decode([]byte(priPemStr))
	if block == nil {
		return nil, nerror.NewRunTimeError("PEM 解码失败")
	}
	privateKey, err := x509.ParsePKCS8UnecryptedPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// SM2 公钥加密
func Sm2Encrypt(pubKey *sm2.PublicKey, plaintext []byte) []byte {
	ciphertext, _ := sm2.Encrypt(pubKey, plaintext, rand.Reader, sm2.C1C3C2)
	return ciphertext
}

func Sm2EncryptToBase64(pubKey *sm2.PublicKey, plaintext string) string {
	encryptData := Sm2Encrypt(pubKey, []byte(plaintext))
	return base64.StdEncoding.EncodeToString(encryptData)
}

// SM2 私钥解密
func Sm2Decrypt(privKey *sm2.PrivateKey, ciphertext []byte) (bool, []byte) {
	plaintext, err := sm2.Decrypt(privKey, ciphertext, sm2.C1C3C2)
	if err != nil {
		return false, nil
	}
	return true, plaintext
}

func Sm2DecryptBase64(privKey *sm2.PrivateKey, base64Str string) (bool, string, error) {
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return false, "", nerror.NewRunTimeError("SM2解密,密文不是base64数据")
	}
	ok, plaintext := Sm2Decrypt(privKey, data)
	return ok, ntools.If3(ok, string(plaintext), ""), nil
}

// SM2 私钥签名 (签名方式 :sm3hash userId=1234567812345678 asn.1 der)
func Sm2SignByPriKey(privKey *sm2.PrivateKey, srcStr string) string {
	r, s, _ := sm2.Sm2Sign(privKey, []byte(srcStr), signUserId, rand.Reader)
	d, _ := asn1.Marshal(sm2Signature{r, s})
	return base64.StdEncoding.EncodeToString(d)
}

// SM2 公钥验签 (签名方式 :sm3hash userId=1234567812345678 asn.1 der)
func Sm2VerifyByPubKey(pubKey *sm2.PublicKey, srcStr, b64DerSign string) bool {
	var sm2Sign sm2Signature
	derBytes, _ := base64.StdEncoding.DecodeString(b64DerSign)
	asn1.Unmarshal(derBytes, &sm2Sign)
	return sm2.Sm2Verify(pubKey, []byte(srcStr), signUserId, sm2Sign.R, sm2Sign.S)
}

// SM2 公钥验签,原文为b64Str (签名方式 :sm3hash userId=1234567812345678 asn.1 der)
func Sm2VerifyB64SrcByPubKey(pubKey *sm2.PublicKey, b64SrcStr, b64DerSign string) bool {
	var sm2Sign sm2Signature
	derBytes, _ := base64.StdEncoding.DecodeString(b64DerSign)
	asn1.Unmarshal(derBytes, &sm2Sign)
	strBytes, _ := base64.StdEncoding.DecodeString(b64SrcStr)
	return sm2.Sm2Verify(pubKey, strBytes, signUserId, sm2Sign.R, sm2Sign.S)
}

func Sm4CbcEnData(key, iv, plaintext string) ([]byte, error) {
	// Pkcs7Pad 明文填充
	pkcs7PadData := Pkcs7Pad([]byte(plaintext), sm4.BlockSize)
	return innerSm4CbcEnData(key, iv, pkcs7PadData)
}

func Sm4CbcEnBytes(key, iv string, data []byte) ([]byte, error) {
	// Pkcs7Pad 明文填充
	pkcs7PadData := PKCS5Padding(data)
	return innerSm4CbcEnData(key, iv, pkcs7PadData)
}

func Sm4CbcEnBytesByHexKey(hexkey, hexIv string, data []byte) ([]byte, error) {
	key, _ := hex.DecodeString(hexkey)
	iv, _ := hex.DecodeString(hexIv)
	return Sm4CbcEnBytes(string(key), string(iv), data)
}

func Sm4CbcEnDataWithPkcs5(key, iv, plaintext string) ([]byte, error) {
	// PKCS5Padding 明文填充
	PKCS5PaddingPadData := PKCS5Padding([]byte(plaintext))
	return innerSm4CbcEnData(key, iv, PKCS5PaddingPadData)
}

func innerSm4CbcEnData(key, iv string, padData []byte) ([]byte, error) {
	block, err := sm4.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	mode := cipher.NewCBCEncrypter(block, []byte(iv))

	resultData := make([]byte, len(padData))
	mode.CryptBlocks(resultData, padData)
	return resultData, err
}

func Sm4CbcEnDataToBase64(key, iv, plaintext string) (string, error) {
	encryptData, err := Sm4CbcEnData(key, iv, plaintext)
	if nil != err {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encryptData), nil
}

func Sm4CbcDnData(key, iv, entryedData string) ([]byte, error) {
	resultData, err := innerSm4CbcDnData(key, iv, entryedData)
	if nil != err {
		return nil, err
	}
	unpadded, err := Pkcs7Unpad(resultData)
	return unpadded, err
}

func Sm4CbcDnBytesByHexKey(hexKey, hexIv string, entryedData []byte) ([]byte, error) {
	key, _ := hex.DecodeString(hexKey)
	iv, _ := hex.DecodeString(hexIv)
	return Sm4CbcDnBytes(string(key), string(iv), entryedData)
}

func Sm4CbcDnBytes(key, iv string, entryedData []byte) ([]byte, error) {
	if len(entryedData)%sm4.BlockSize != 0 {
		return nil, nerror.NewRunTimeError("密文长度无效")
	}
	block, err := sm4.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	resultData := make([]byte, len(entryedData))
	mode.CryptBlocks(resultData, []byte(entryedData))

	unpadded, err := Pkcs7Unpad(resultData)
	return unpadded, err
}

func innerSm4CbcDnData(key, iv, entryedData string) ([]byte, error) {
	if len(entryedData)%sm4.BlockSize != 0 {
		return nil, nerror.NewRunTimeError("密文长度无效")
	}
	block, err := sm4.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	mode := cipher.NewCBCDecrypter(block, []byte(iv))

	resultData := make([]byte, len(entryedData))
	mode.CryptBlocks(resultData, []byte(entryedData))
	return resultData, nil
}

func Sm4CbcDnDataWithPkcs5(key, iv, entryedData string) ([]byte, error) {
	resultData, err := innerSm4CbcDnData(key, iv, entryedData)
	if nil != err {
		return nil, err
	}
	unpadded := PKCS5UnPadding(resultData)
	return unpadded, nil
}

func Sm4CbcDnBase64Data(key, iv, encryptBase64Data string) (string, error) {
	encryedData, err := base64.StdEncoding.DecodeString(encryptBase64Data)
	if err != nil {
		return "", nerror.NewRunTimeError("SM2解密,密文不是base64数据")
	}
	result, err := Sm4CbcDnData(key, iv, string(encryedData))
	return string(result), err
}

func GetSm2PubKeyHexFromPriKey(privateKey *sm2.PrivateKey) string {
	pubKey := privateKey.PublicKey
	data := append([]byte{0x04}, append(pubKey.X.Bytes(), pubKey.Y.Bytes()...)...)
	return hex.EncodeToString(data)
}

func Sm3hash(data []byte) []byte {
	h := sm3.New()
	h.Write(data)
	return h.Sum(nil)
}

func Sm4EcbPkcs5EnData2HexStr(hexKey, plaintext string) (string, error) {
	hexKeyData, _ := hex.DecodeString(hexKey)
	// 检查密钥长度
	if len(hexKeyData) != 16 {
		return "", nerror.NewRunTimeError("SM4加密,key必须为16位")
	}
	padData := PKCS5Padding([]byte(plaintext))
	//
	block, err := sm4.NewCipher(hexKeyData)
	if err != nil {
		return "", err
	}
	// ECB 模式分块加密（无 IV）
	enData := make([]byte, len(padData))
	for start := 0; start < len(padData); start += 16 { // 16 字节分组
		block.Encrypt(enData[start:], padData[start:start+16])
	}
	return hex.EncodeToString(enData), nil
}

func Sm4EcbPkcs5DnHexStr(hexKey, enedStr string) (string, error) {
	hexKeyData, _ := hex.DecodeString(hexKey)
	// 检查密钥长度
	if len(hexKeyData) != 16 {
		return "", nerror.NewRunTimeError("SM4加密,key必须为16位")
	}
	block, err := sm4.NewCipher(hexKeyData)
	if err != nil {
		return "", err
	}
	data, _ := hex.DecodeString(enedStr)
	plaintext := make([]byte, len(data))
	for start := 0; start < len(plaintext); start += 16 {
		block.Decrypt(plaintext[start:], data[start:start+16])
	}
	unPadData := PKCS5UnPadding([]byte(plaintext))
	return string(unPadData), nil
}

func Sm4GenHexKeyIv() (string, string) {
	randomStr := ntools.MD5Str(ntools.UUIDStr(true), true)
	return Sm4GenHexKeyIvFromMd5Str(randomStr)
}

func Sm4GenHexKeyIvFromMd5Str(md5Str string) (string, string) {
	key := []byte(md5Str[:16])
	iv := []byte(md5Str[16:32])
	return hex.EncodeToString(key), hex.EncodeToString(iv)
}
