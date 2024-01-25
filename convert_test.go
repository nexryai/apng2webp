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

	rb, _ := Convert(&b, 480, 400)

	// Write to file
	f, err = os.Create("output.webp")
	if err != nil {
		panic(err)
	}

	_, err = f.Write(*rb)
	log.Printf("Done\n")
}
