package message

import (
	"encoding/binary"
	"log"
	"testing"
)

func TestEndianNess(t *testing.T) {
	log.Printf("%#x - LE:%d BE:%d", []byte{0x00, 0x01}, binary.LittleEndian.Uint16([]byte{0x00, 0x01}), binary.BigEndian.Uint16([]byte{0x00, 0x01}))
}
