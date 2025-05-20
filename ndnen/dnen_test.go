package ndnen_test

import (
	"fmt"
	"testing"

	"github.com/niexqc/nlibs/ndnen"
)

func TestGenSm2Key(t *testing.T) {
	pri, pub := ndnen.Sm2GenKeyPair()

	priStr, pubStr := ndnen.Sm2Key2Hex(pri, pub)
	fmt.Println("私钥Hex:\n" + priStr)
	fmt.Println("公钥Hex:\n" + pubStr)
	pri = ndnen.Sm2LoadPriKeyFromHex(priStr)
	pub = ndnen.Sm2LoadPubKeyFromHex(pubStr)

	priStr, pubStr = ndnen.Sm2Key2Der2B64(pri, pub)
	fmt.Println("私钥DerB64:\n" + priStr)
	fmt.Println("公钥DerB64:\n" + pubStr)
	pri = ndnen.Sm2LoadPriKeyFromDerB64(priStr)
	pub = ndnen.Sm2LoadPubKeyFromDerB64(pubStr)

	priStr, pubStr = ndnen.Sm2Key2Pem(pri, pub)
	fmt.Println("私钥DerPem:\n" + priStr)
	fmt.Println("公钥DerPem:\n" + pubStr)
	pri = ndnen.Sm2LoadPriKeyFromPem(priStr)
	pub = ndnen.Sm2LoadPubKeyFromPem(pubStr)

	priStr, pubStr = ndnen.Sm2Key2Hex(pri, pub)
	fmt.Println("私钥Hex:\n" + priStr)
	fmt.Println("公钥Hex:\n" + pubStr)

}

func TestSm2EnDn(t *testing.T) {
	pubHexStr := "04bdedf93932233a806a424efeb13b868bb50c3ba45860334b75342602743e39ee0d5d9f9a189ded81b2772fe229d7e992b1bb19f4fbcb541ff9d150a5e1da7806"
	priHexStr := "626d2ea7d29769e3c89a1944c53b4fa35cff8db32300718a5885f0fa426d56f2"
	pri := ndnen.Sm2LoadPriKeyFromHex(priHexStr)
	pub := ndnen.Sm2LoadPubKeyFromHex(pubHexStr)

	sourceStr := "123"

	fmt.Printf("-----------------SM2加密解密--------------------\n")
	sm2EnedStr := ndnen.Sm2EncryptToBase64(pub, sourceStr)
	fmt.Printf("原文:%s,Sm2公钥加密后:%s\n", sourceStr, sm2EnedStr)
	_, sm2DnedStr := ndnen.Sm2DecryptBase64(pri, sm2EnedStr)
	fmt.Printf("密文:%s,Sm2私钥解密后:%s\n", sm2EnedStr, sm2DnedStr)

}

func TestSm2EnDn2(t *testing.T) {
	pubDer := "MFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEQszcg+sXDe4G4bWEpy1AEA520bnjMNSqvugbCuhw3yAMDjpNsuUgwaNApc1BGuN2Ggd35qpXL+IuLNi8n/jOvA=="
	priDer := "MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQglubXvo2TJIu3lJ+wNA/q/IryTemzNoXlDLN7r4dUbeygCgYIKoEcz1UBgi2hRANCAARCzNyD6xcN7gbhtYSnLUAQDnbRueMw1Kq+6BsK6HDfIAwOOk2y5SDBo0ClzUEa43YaB3fmqlcv4i4s2Lyf+M68"
	pri := ndnen.Sm2LoadPriKeyFromDerB64(priDer)
	pub := ndnen.Sm2LoadPubKeyFromDerB64(pubDer)

	sourceStr := "123"

	fmt.Printf("-----------------SM2加密解密--------------------\n")
	sm2EnedStr := ndnen.Sm2EncryptToBase64(pub, sourceStr)
	fmt.Printf("原文:%s,Sm2公钥加密后:%s\n", sourceStr, sm2EnedStr)
	_, sm2DnedStr := ndnen.Sm2DecryptBase64(pri, "BOaUU3yx34ZmIs2FBSEgKheoy2xsS0+RKBlJ04LONntuws7yvrwNedEdfP6xgyfodjKReO+h6Y3TuT39E4IPb4Qt0+XlpBu3uiiKn5rHbXI+QvS0pQ2gIflvy0qI6pbYRiHnSA==")
	fmt.Printf("密文:%s,Sm2私钥解密后:%s\n", sm2EnedStr, sm2DnedStr)

}

func TestSm2Verify(t *testing.T) {
	pubHexStr := "04bdedf93932233a806a424efeb13b868bb50c3ba45860334b75342602743e39ee0d5d9f9a189ded81b2772fe229d7e992b1bb19f4fbcb541ff9d150a5e1da7806"
	priHexStr := "626d2ea7d29769e3c89a1944c53b4fa35cff8db32300718a5885f0fa426d56f2"
	pri := ndnen.Sm2LoadPriKeyFromHex(priHexStr)
	pub := ndnen.Sm2LoadPubKeyFromHex(pubHexStr)

	sourceStr := "123"

	fmt.Printf("-----------------SM2签名验签--------------------\n")
	// JS签名结果
	jsGenB64DerSignStr := "MEUCIQCLVsjVmGMAjVUAg3tuQEniyEYcIlFeSF0ninfNfl7P5AIgC1zD5WedSJlU/8ZQVQeaEvz9XWrrKxO6kBvw7wZNb/A="
	verfiyResult := ndnen.Sm2VerifyByPubKey(pub, sourceStr, jsGenB64DerSignStr)
	fmt.Printf("原文:%s   JSGen签名:%s   Sm2私钥验签:%v\n", sourceStr, jsGenB64DerSignStr, verfiyResult)

	// JS签名结果
	signStr := ndnen.Sm2SignByPriKey(pri, sourceStr)
	fmt.Printf("原文:%s   Sm2私钥签名后:%s\n", sourceStr, signStr)
	verfiyResult = ndnen.Sm2VerifyByPubKey(pub, sourceStr, signStr)
	fmt.Printf("自签验证:%v\n", verfiyResult)

}

func TestSm4EnDn(t *testing.T) {
	sourceStr := "123"
	sm4Key := "ABCDEF1234567890"
	sm4KeyIv := "ABCDEF1234567890"

	fmt.Printf("-----------------SM4加密解密--------------------\n")

	endStr := ndnen.Sm4CbcEnDataToBase64(sm4Key, sm4KeyIv, sourceStr)
	fmt.Printf("原文:%s   秘钥:%s   向量:%s   加密后:%s\n", sourceStr, sm4Key, sm4KeyIv, endStr)
	dndStr := ndnen.Sm4CbcDnBase64Data(sm4Key, sm4KeyIv, endStr)
	fmt.Printf("密文:%s   秘钥:%s   向量:%s   解密后:%s\n", endStr, sm4Key, sm4KeyIv, dndStr)
}

func TestSm4EcbEnDn(t *testing.T) {
	sourceStr := "123"
	sm4Key := "3e448cb55fc737f50a0851d5e34c6473"

	fmt.Printf("-----------------SM4加密解密--------------------\n")
	enedStr := ndnen.Sm4EcbPkcs5EnData2HexStr(sm4Key, sourceStr)
	fmt.Printf("原文:%s   秘钥:%s   加密后:%s\n", sourceStr, sm4Key, enedStr)

	dnStr := ndnen.Sm4EcbPkcs5DnHexStr(sm4Key, "f49734fccc350dedcd45b1886adb239c")
	fmt.Printf("原文:%s   秘钥:%s   解密后:%s\n", sourceStr, sm4Key, dnStr)

	// dndStr := ndnen.Sm4CbcDnBase64Data(sm4Key, sm4KeyIv, endStr)
	// fmt.Printf("密文:%s   秘钥:%s   向量:%s   解密后:%s\n", endStr, sm4Key, sm4KeyIv, dndStr)
}
