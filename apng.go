package apng2webp

import (
	"bytes"
	"encoding/binary"
)

func readInt32(buffer *bytes.Buffer) int {
	var data [4]byte
	buffer.Read(data[:])
	return int(binary.BigEndian.Uint32(data[:]))
}
