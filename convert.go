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
	"unsafe"
)

// Goのimage.Imageをlibwebp.WebPPictureに変換
func imageToWebPPicture(img *image.Image, scale float32, width int, height int, xOffset int, yOffset int) C.WebPPicture {
	bounds := (*img).Bounds()
	fmt.Printf("Dx: %v Dy: %v\n", bounds.Dx(), bounds.Dy())

	var pic C.WebPPicture
	C.WebPPictureInit(&pic)

	pic.width = C.int(width)
	pic.height = C.int(height)
	pic.use_argb = 1

	// RGBAイメージに変換
	rgbaImg := image.NewRGBA(image.Rect(0, 0, width, height))

	if scale != 1 {
		newWidth := int(float32((*img).Bounds().Dx()) * scale)
		newHeight := int(float32((*img).Bounds().Dy()) * scale)

		//xOffset = int(float32(xOffset) * scale)
		//yOffset = int(float32(yOffset) * scale)

		fmt.Printf("newWidth: %v newHeight: %v\n", newWidth, newHeight)
		draw.ApproxBiLinear.Scale(rgbaImg, image.Rect(xOffset, yOffset, newWidth, newHeight), *img, bounds, draw.Src, nil)
	} else {
		draw.Draw(rgbaImg, image.Rect(xOffset, yOffset, width, height), *img, bounds.Min, draw.Src)
	}

	fmt.Printf("xOffset: %v yOffset: %v\n", xOffset, yOffset)

	// WebPにエンコード
	C.WebPPictureImportRGBA(&pic, (*C.uint8_t)(unsafe.Pointer(&rgbaImg.Pix[0])), C.int(rgbaImg.Stride))

	return pic
}

func Convert(imgPtr *[]byte, width int, height int) (*[]byte, error) {
	// libwebpの初期化
	buffer := bytes.NewBuffer(*imgPtr)
	// Skip the first 8 bytes (PNG signature)
	buffer.Next(8)

	// Skip chunk type (8 bytes)
	buffer.Next(8)

	originalWidth := readInt32(buffer)
	originalHeight := readInt32(buffer)

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

			// webpとしてエンコード
			pic := imageToWebPPicture(f.Buffer, scale, width, height, f.OffsetX, f.OffsetY)

			// リサイズ
			C.WebPPictureRescale(&pic, C.int(width), C.int(height))

			timeStamp := int(float32(i) * f.Delay * 1000)
			fmt.Printf("timeStamp: %d\n", timeStamp)

			// Animated WebPのフレームとして追加
			result := C.int(C.WebPAnimEncoderAdd(animEncoder, &pic, C.int(timeStamp), nil))
			if result == 0 {
				// animEncoderの解放
				C.WebPPictureFree(&pic)
				return fmt.Errorf("WebPAnimEncoderAdd failed")
			}

			// Cleanup
			C.WebPPictureFree(&pic)

			return nil
		})

	if err != nil {
		C.WebPAnimEncoderDelete(animEncoder)
		return nil, err
	}

	fmt.Printf("i: %v\n", i)

	// 書き込み
	var webpData C.WebPData
	C.WebPDataInit(&webpData)
	C.WebPAnimEncoderAssemble(animEncoder, &webpData)
	webpBytes := C.GoBytes(unsafe.Pointer(webpData.bytes), C.int(webpData.size))

	// animEncoderの解放
	C.WebPDataClear(&webpData)
	C.WebPAnimEncoderDelete(animEncoder)

	return &webpBytes, nil
}
