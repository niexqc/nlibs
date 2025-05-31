package ndnen_test

import (
	"log/slog"
	"testing"

	"github.com/niexqc/nlibs/ndnen"
	"github.com/niexqc/nlibs/ntools"
)

func TestKeyGen(t *testing.T) {
	ntools.SlogConf("test", "debug", 1, 2)
	pri, pub := ndnen.Sm2GenKeyPair()
	priHex, pubHex := ndnen.Sm2Key2Hex(pri, pub)
	slog.Info("私钥 Hex:" + priHex)
	slog.Info("公钥 Hex:" + pubHex)
	ntools.TestEq(t, "测试 Sm2Key2Hex,私钥长度不匹配", 64, len(priHex))
	ntools.TestEq(t, "测试 Sm2Key2Hex,公钥长度不匹配", 130, len(pubHex))

	priDerB64, pubDerB64 := ndnen.Sm2Key2Der2B64(pri, pub)
	slog.Info("私钥 DerB64:" + priDerB64)
	slog.Info("公钥 DerB64:" + pubDerB64)
	ntools.TestEq(t, "测试 Sm2Key2Der2B64,私钥长度不匹配", 200, len(priDerB64))
	ntools.TestEq(t, "测试 Sm2Key2Der2B64,公钥长度不匹配", 124, len(pubDerB64))

	priPem, pubPem := ndnen.Sm2Key2Pem(pri, pub)
	slog.Info("私钥 DerPem:\n" + priPem)
	slog.Info("公钥 DerPem:\n" + pubPem)
	ntools.TestEq(t, "测试 Sm2Key2Pem,私钥长度不匹配", 258, len(priPem))
	ntools.TestEq(t, "测试 Sm2Key2Pem,公钥长度不匹配", 178, len(pubPem))
}

func TestLoadFromHex(t *testing.T) {
	keyXText := "9110175885747722819732440574911524584232116321888268287657103114450431652835"
	keyYText := "23902428675165520847631018454485887619039328327888689927080668929818068476405"
	keyDText := "79293913723962675249728469258961058083043805132026114477962318078938045835019"
	priHexStr := "af4ec3c4f851fa875a7064be5bc8b7b42dc0c24f30fdc42036a4aa52f8a9670b"
	pubHexStr := "0414242d444ae79fcf1942bc3e490066100c60bae0f17c47c6199d83bc963433e334d84b929539d07f4d073d75d837b456b6dce34f9488049e632a44905f4361f5"

	priKeyLoadFromHex := ndnen.Sm2LoadPriKeyFromHex(priHexStr)
	if keyXText != priKeyLoadFromHex.X.Text(10) || keyYText != priKeyLoadFromHex.Y.Text(10) || keyDText != priKeyLoadFromHex.D.Text(10) {
		ntools.TestErrPanicMsg(t, "从Hex加载私钥后,X,Y,D不匹配")
	}
	pubKeyLoadFromHex := ndnen.Sm2LoadPubKeyFromHex(pubHexStr)
	if keyXText != pubKeyLoadFromHex.X.Text(10) || keyYText != pubKeyLoadFromHex.Y.Text(10) {
		ntools.TestErrPanicMsg(t, "从Hex加载公钥后,X,Y不匹配")
	}

}

func TestLoadFromDerB64(t *testing.T) {
	keyXText := "45491728670726455077704469908930866236637347913347224320666877715914640567669"
	keyYText := "12122001825702705280758340394046163503492966848176631225078848642982586336973"
	keyDText := "10575470501061192161446575300722225078608851643319307194990556739345271688258"
	priDerB64Str := "MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgF2GBEkocYjmxhaSS28+4KfTfM2h6cx4KDjuxqhBfDEKgCgYIKoEcz1UBgi2hRANCAARkk2ft6Js9wKQwGKEHtrUyb3aMfqBiTNW6l+b4tAOtdRrMz1VOLxrBwWzN3z/upT29N5TVowpPVNiW3XjgR8bN"
	pubDerB64Str := "MFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEZJNn7eibPcCkMBihB7a1Mm92jH6gYkzVupfm+LQDrXUazM9VTi8awcFszd8/7qU9vTeU1aMKT1TYlt144EfGzQ=="
	priKeyLoadFromDerB64 := ndnen.Sm2LoadPriKeyFromDerB64(priDerB64Str)
	if keyXText != priKeyLoadFromDerB64.X.Text(10) || keyYText != priKeyLoadFromDerB64.Y.Text(10) || keyDText != priKeyLoadFromDerB64.D.Text(10) {
		ntools.TestErrPanicMsg(t, "从Der2B64加载私钥后,X,Y,D不匹配")
	}
	pubKeyLoadFromDerB64 := ndnen.Sm2LoadPubKeyFromDerB64(pubDerB64Str)
	if keyXText != pubKeyLoadFromDerB64.X.Text(10) || keyYText != pubKeyLoadFromDerB64.Y.Text(10) {
		ntools.TestErrPanicMsg(t, "从Der2B64加载公钥后,X,Y不匹配")
	}
}

func TestKey2DerPem(t *testing.T) {
	keyXText := "77961794746177998525898831629860805532241655337404864433242088763761881008284"
	keyYText := "78632196359869886911192869185435857914453293536150231342822745897639896700968"
	keyDText := "34191172974912639021300555522713401434497511740003000268687450958359715855216"
	priPemStr := `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgS5eEovpGSYshmL5P
b6B4Na2db3LFJNqDP7CBXiQ3O3CgCgYIKoEcz1UBgi2hRANCAASsXM/l6VqBhVEf
RAWjSh7tD3PflrZqaeY5+xHPWQBInK3YPvfNCBP6oqKkiwDWAJmHvi9wAnCj2TCa
lHq0sFgo
-----END PRIVATE KEY-----`
	pubPemStr := `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAErFzP5elagYVRH0QFo0oe7Q9z35a2
amnmOfsRz1kASJyt2D73zQgT+qKipIsA1gCZh74vcAJwo9kwmpR6tLBYKA==
-----END PUBLIC KEY-----`

	priKeyLoadFromPem := ndnen.Sm2LoadPriKeyFromPem(priPemStr)
	if keyXText != priKeyLoadFromPem.X.Text(10) || keyYText != priKeyLoadFromPem.Y.Text(10) || keyDText != priKeyLoadFromPem.D.Text(10) {
		ntools.TestErrPanicMsg(t, "从DerPem加载私钥后,X,Y,D不匹配")
	}
	pubKeyLoadFromPem := ndnen.Sm2LoadPubKeyFromPem(pubPemStr)
	if keyXText != pubKeyLoadFromPem.X.Text(10) || keyYText != pubKeyLoadFromPem.Y.Text(10) {
		ntools.TestErrPanicMsg(t, "从DerPem加载公钥后,X,Y不匹配")
	}
}

func TestSm2EnDn(t *testing.T) {
	pubHexStr := "04bdedf93932233a806a424efeb13b868bb50c3ba45860334b75342602743e39ee0d5d9f9a189ded81b2772fe229d7e992b1bb19f4fbcb541ff9d150a5e1da7806"
	priHexStr := "626d2ea7d29769e3c89a1944c53b4fa35cff8db32300718a5885f0fa426d56f2"
	pri := ndnen.Sm2LoadPriKeyFromHex(priHexStr)
	pub := ndnen.Sm2LoadPubKeyFromHex(pubHexStr)
	orginText := "123"
	sm2EnedStr := ndnen.Sm2EncryptToBase64(pub, orginText)
	dnok, sm2DnedStr := ndnen.Sm2DecryptBase64(pri, sm2EnedStr)
	if !dnok {
		ntools.TestErrPanicMsg(t, "SM2加解密失败")
	}
	ntools.TestEq(t, "测试 TestSm2EnDn 加密解密", orginText, sm2DnedStr)
	//密文
	sm2EnedStr = "BMKUnLTSPEFjTyYQsdnJetfWcieodAphSsfbrNeWuN3JGLCcBQ9q/w7rr81cA6KL5PMX1Oxw3fmANPDWzCmJtI2nNDY9WIGEHeWLejDIq8gU+jb0TOlda/N6EcEaD7tmAhR5tw=="
	dnok, sm2DnedStr = ndnen.Sm2DecryptBase64(pri, sm2EnedStr)
	if !dnok {
		ntools.TestErrPanicMsg(t, "SM2加解密失败")
	}
	ntools.TestEq(t, "测试 TestSm2EnDn 加密解密", orginText, sm2DnedStr)

}

func TestSm2Verify(t *testing.T) {
	pubHexStr := "04bdedf93932233a806a424efeb13b868bb50c3ba45860334b75342602743e39ee0d5d9f9a189ded81b2772fe229d7e992b1bb19f4fbcb541ff9d150a5e1da7806"
	priHexStr := "626d2ea7d29769e3c89a1944c53b4fa35cff8db32300718a5885f0fa426d56f2"
	pri := ndnen.Sm2LoadPriKeyFromHex(priHexStr)
	pub := ndnen.Sm2LoadPubKeyFromHex(pubHexStr)

	sourceStr := "123"
	jsGenB64DerSignStr := "MEUCIQCLVsjVmGMAjVUAg3tuQEniyEYcIlFeSF0ninfNfl7P5AIgC1zD5WedSJlU/8ZQVQeaEvz9XWrrKxO6kBvw7wZNb/A="
	verfiyResult := ndnen.Sm2VerifyByPubKey(pub, sourceStr, jsGenB64DerSignStr)
	slog.Info("Sm2Verify 签名验证", "原文", sourceStr, "签名", jsGenB64DerSignStr, "结果", verfiyResult)
	if !verfiyResult {
		ntools.TestErrPanicMsg(t, "Sm2Verify验签失败")
	}

	// 自签验证
	selfSignStr := ndnen.Sm2SignByPriKey(pri, sourceStr)
	verfiyResult = ndnen.Sm2VerifyByPubKey(pub, sourceStr, selfSignStr)
	slog.Info("Sm2Verify 自签验证", "原文", sourceStr, "签名", jsGenB64DerSignStr, "结果", verfiyResult)
	if !verfiyResult {
		ntools.TestErrPanicMsg(t, "Sm2Verify自签验证失败")
	}

}

func TestSm4CbcEnDn(t *testing.T) {
	sourceStr := "123"
	sm4Key := "ABCDEF1234567890"
	sm4KeyIv := "ABCDEF1234567890"

	sm4EndStr := "a8Vdi0+57ngGosjiiglIhA=="

	endStr := ndnen.Sm4CbcEnDataToBase64(sm4Key, sm4KeyIv, sourceStr)
	slog.Info("Sm4CbcEnDn 加密", "原文", sourceStr, "秘钥", sm4Key, "向量", sm4KeyIv, "期望结果", sm4EndStr, "加密结果", endStr)
	ntools.TestEq(t, "测试 Sm4CbcEnDn 加密解密", sm4EndStr, endStr)

	dndStr := ndnen.Sm4CbcDnBase64Data(sm4Key, sm4KeyIv, sm4EndStr)
	slog.Info("Sm4CbcEnDn 解密", "原文", sourceStr, "秘钥", sm4Key, "向量", sm4KeyIv, "期望结果", sourceStr, "解密结果", dndStr)
	ntools.TestEq(t, "测试 Sm4CbcEnDn 加密解密", sourceStr, dndStr)

}

func TestSm4EcbEnDn(t *testing.T) {
	sourceStr := "123"
	sm4Key := "3e448cb55fc737f50a0851d5e34c6473"

	sm4EndStr := "f49734fccc350dedcd45b1886adb239c"

	endStr := ndnen.Sm4EcbPkcs5EnData2HexStr(sm4Key, sourceStr)
	slog.Info("Sm4CbcEnDn 加密", "原文", sourceStr, "Hex秘钥", sm4Key, "期望结果", sm4EndStr, "加密结果", endStr)
	ntools.TestEq(t, "测试 Sm4CbcEnDn 加密解密", sm4EndStr, endStr)

	dndStr := ndnen.Sm4EcbPkcs5DnHexStr(sm4Key, sm4EndStr)
	slog.Info("Sm4CbcEnDn 解密", "原文", sourceStr, "Hex秘钥", sm4Key, "期望结果", sourceStr, "解密结果", dndStr)
	ntools.TestEq(t, "测试 Sm4CbcEnDn 加密解密", sourceStr, dndStr)

}
