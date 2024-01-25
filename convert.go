package apng2webp

/*
   #cgo LDFLAGS: -lwebp -lwebpmux
   #include <webp/encode.h>
   #include <webp/mux.h>
*/
import "C"

import (
	"bytes"
	"fmt"
	"github.com/nexryai/apng"
	"image"
	"image/draw"
	"image/png"
	"os"
	"unsafe"
)

const (
	width  = 480
	height = 400
)

// Goのimage.Imageをlibwebp.WebPPictureに変換
func imageToWebPPicture(img *image.Image) C.WebPPicture {
	bounds := (*img).Bounds()
	fmt.Printf("Dx: %v Dy: %v\n", bounds.Dx(), bounds.Dy())

	var pic C.WebPPicture
	C.WebPPictureInit(&pic)

	pic.width = C.int(width)
	pic.height = C.int(height)
	pic.use_argb = 1

	// RGBAイメージに変換
	rgbaImg := image.NewRGBA(image.Rect(0, 0, width, height))

	xOffset := (width - bounds.Dx()) / 2
	yOffset := (height - bounds.Dy()) / 2

	draw.Draw(rgbaImg, image.Rect(xOffset, yOffset, width, height), *img, bounds.Min, draw.Src)

	file, err := os.Create("debug.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = png.Encode(file, rgbaImg)
	if err != nil {
		panic(err)
	}

	// WebPにエンコード
	C.WebPPictureImportRGBA(&pic, (*C.uint8_t)(unsafe.Pointer(&rgbaImg.Pix[0])), C.int(rgbaImg.Stride))

	return pic
}

func ApngToWebP(imgPtr *[]byte) {
	// libwebpの初期化
	buffer := bytes.NewBuffer(*imgPtr)
	// Skip the first 8 bytes (PNG signature)
	buffer.Next(8)

	// Skip chunk type (4 bytes)
	buffer.Next(4)

	originalWidth := readInt32(buffer)
	originalHeight := readInt32(buffer)

	//scale := float32(height) / float32(originalHeight)

	fmt.Printf("originalWidth: %d, originalHeight: %d\n", originalWidth, originalHeight)

	var animConfig C.WebPAnimEncoderOptions
	C.WebPAnimEncoderOptionsInit(&animConfig)
	animEncoder := C.WebPAnimEncoderNew(C.int(width), C.int(height), &animConfig)

	i := 0
	_, err := apng.DecodeAll(bytes.NewReader(*imgPtr),
		func(frame *image.Image, frameNum int, frameDelay float32) error {

			if i == 0 {
				// 最初のフレームはスキップ
				i += 1
				return nil
			}

			println("frame:", i)
			i += 1

			// pngとしてframesディレクトリに保存
			/*
				pngFile, err := os.Create(fmt.Sprintf("frames/%d.png", frameNum))
				if err != nil {
					panic(err)
				}
				defer pngFile.Close()

				err = png.Encode(pngFile, *frame)
				if err != nil {
					panic(err)
				}
			*/

			// webpとしてエンコード
			pic := imageToWebPPicture(frame)

			// リサイズ
			C.WebPPictureRescale(&pic, C.int(width), C.int(height))

			timeStamp := int(float32(i) * frameDelay * 1000)
			fmt.Printf("timeStamp: %d\n", timeStamp)

			// Animated WebPのフレームとして追加
			result := C.int(C.WebPAnimEncoderAdd(animEncoder, &pic, C.int(timeStamp), nil))
			if result == 0 {
				panic("WebPAnimEncoderAdd failed")
			}

			// Cleanup
			C.WebPPictureFree(&pic)

			return nil
		})

	fmt.Printf("i: %v\n", i)

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
