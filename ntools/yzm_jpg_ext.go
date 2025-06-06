package ntools

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"image/color"
	"image/jpeg"
	"log/slog"

	"github.com/steambap/captcha"
)

func YzmJpgGenCode(w, h, yzmLen int) (txt string, imgBytes []byte) {
	//字体
	// fontBytes, _ := fileext.ReadFileByte("resources/FiraCode-Retina.ttf")
	// captcha.LoadFont(fontBytes)
	data, _ := captcha.New(w, h, func(options *captcha.Options) {
		options.BackgroundColor = color.Opaque
		options.CharPreset = "1234567890"
		options.CurveNumber = 4
		options.TextLength = yzmLen
		options.FontDPI = float64(float64(w) * float64(0.99))
		options.FontScale = 1.0
		options.Noise = 0.01

		options.Palette = []color.Color{
			color.RGBA{0x00, 0x00, 0x00, 0xFF},
		}
	})

	buffer := bytes.NewBuffer(nil)
	writer := bufio.NewWriter(buffer)
	// data.WriteGIF(writer, &gif.Options{})
	data.WriteJPG(writer, &jpeg.Options{Quality: 75})
	bys := buffer.Bytes()
	slog.Debug(fmt.Sprintf("yzm:%s,jpg_bytes_len:%d", data.Text, len(bys)))
	return data.Text, bys
}

func YzmJpgGenCodeBase64(w, h, yzmLen int) (txt string, imgb64 string) {
	txt, imgBytes := YzmJpgGenCode(w, h, yzmLen)
	// 编码为 Base64 字符串
	base64Str := base64.StdEncoding.EncodeToString(imgBytes)
	// 拼接 Data URI 前缀（适用于网页嵌入）
	return txt, "data:image/jpeg;base64," + base64Str
}
