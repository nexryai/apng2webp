package apng2webp

import (
	"io"
	"log"
	"os"
	"testing"
)

func TestApngToWebP(t *testing.T) {
	f, err := os.Open("input.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	ConvertAndSaveFile(&b, 480, 400, "output.webp")
	log.Printf("Done\n")
}
