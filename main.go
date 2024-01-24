package apng2webp

/*
   #cgo LDFLAGS: -lwebp -lwebpmux
   #include <webp/encode.h>
   #include <webp/mux.h>
*/
import "C"

import (
	"github.com/nexryai/apng"
	"image"
	"image/draw"
	"os"
	"unsafe"
)

// Goのimage.Imageをlibwebp.WebPPictureに変換
func imageToWebPPicture(img *image.Image) C.WebPPicture {
	bounds := (*img).Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	var pic C.WebPPicture
	C.WebPPictureInit(&pic)
	pic.width = C.int(width)
	pic.height = C.int(height)
	pic.use_argb = 1

	// RGBAイメージに変換
	rgbaImg := image.NewRGBA(bounds)
	draw.Draw(rgbaImg, bounds, *img, bounds.Min, draw.Src)

	// WebPにエンコード
	C.WebPPictureImportRGBA(&pic, (*C.uint8_t)(unsafe.Pointer(&rgbaImg.Pix[0])), C.int(rgbaImg.Stride))

	return pic
}

func main() {
	// libwebpの初期化
	println("Using libwebp")
	width := 240
	height := 200

	var animConfig C.WebPAnimEncoderOptions
	C.WebPAnimEncoderOptionsInit(&animConfig)
	animEncoder := C.WebPAnimEncoderNew(C.int(width), C.int(height), &animConfig)

	f, err := os.Open("test.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = apng.DecodeAll(f, func(frame *image.Image, frameNum int) error {
		println("frame:", frame)
		// webpとしてエンコード
		pic := imageToWebPPicture(frame)

		// リサイズ
		C.WebPPictureRescale(&pic, C.int(width), C.int(height))

		// Animated WebPのフレームとして追加
		C.WebPAnimEncoderAdd(animEncoder, &pic, C.int(frameNum*100), nil)

		// Cleanup
		C.WebPPictureFree(&pic)

		return nil
	})

	// ファイルに書き込み
	var webpData C.WebPData
	C.WebPDataInit(&webpData)
	C.WebPAnimEncoderAssemble(animEncoder, &webpData)

	webpFile, err := os.Create("output.webp")
	if err != nil {
		panic(err)
	}
	defer webpFile.Close()

	webpBytes := C.GoBytes(unsafe.Pointer(webpData.bytes), C.int(webpData.size))
	webpFile.Write(webpBytes)

	// animEncoderの解放
	C.WebPDataClear(&webpData)
	C.WebPAnimEncoderDelete(animEncoder)
}
