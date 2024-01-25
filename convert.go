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
	"golang.org/x/image/draw"
	"image"
	"image/png"
	"os"
	"unsafe"
)

const (
	width  = 480
	height = 400
)

// Goのimage.Imageをlibwebp.WebPPictureに変換
func imageToWebPPicture(img *image.Image, scale float32, xOffset int, yOffset int) C.WebPPicture {
	bounds := (*img).Bounds()
	fmt.Printf("Dx: %v Dy: %v\n", bounds.Dx(), bounds.Dy())

	var pic C.WebPPicture
	C.WebPPictureInit(&pic)

	pic.width = C.int(width)
	pic.height = C.int(height)
	pic.use_argb = 1

	// RGBAイメージに変換
	rgbaImg := image.NewRGBA(image.Rect(0, 0, width, height))

	fmt.Printf("xOffset: %v yOffset: %v\n", xOffset, yOffset)

	xOffset = int(scale * float32(xOffset))
	yOffset = int(scale * float32(yOffset))

	if scale != 1 {
		newWidth := uint(float32((*img).Bounds().Dx()) * scale)
		newHeight := uint(float32((*img).Bounds().Dy()) * scale)

		fmt.Printf("newWidth: %v newHeight: %v\n", newWidth, newHeight)
		draw.ApproxBiLinear.Scale(rgbaImg, image.Rect(xOffset, yOffset, width, height), *img, (*img).Bounds(), draw.Src, nil)
	} else {
		draw.Draw(rgbaImg, image.Rect(xOffset, yOffset, width, height), *img, bounds.Min, draw.Src)
	}

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

	// Skip chunk type (8 bytes)
	buffer.Next(8)

	originalWidth := readInt32(buffer)
	originalHeight := readInt32(buffer)

	//scale := float32(height) / float32(originalHeight)

	fmt.Printf("originalWidth: %d, originalHeight: %d\n", originalWidth, originalHeight)

	var animConfig C.WebPAnimEncoderOptions
	C.WebPAnimEncoderOptionsInit(&animConfig)
	animEncoder := C.WebPAnimEncoderNew(C.int(width), C.int(height), &animConfig)

	scale := float32(height) / float32(originalHeight)
	fmt.Printf("scale: %v\n", scale)

	i := 0
	_, err := apng.DecodeAll(bytes.NewReader(*imgPtr),
		func(f *apng.FrameHookArgs) error {

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
			pic := imageToWebPPicture(f.Buffer, scale, f.OffsetX, f.OffsetY)

			// リサイズ
			C.WebPPictureRescale(&pic, C.int(width), C.int(height))

			timeStamp := int(float32(i) * f.Delay * 1000)
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
