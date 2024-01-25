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

	ApngToWebP(&b)
	log.Printf("Done\n")
}
