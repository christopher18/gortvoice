package audioinput

import (
	"bytes"
	"encoding/binary"
)

func int16ToLittleEndian(data []int16) []byte {
	buf := bytes.NewBuffer(nil)
	for _, v := range data {
		binary.Write(buf, binary.LittleEndian, v)
	}
	return buf.Bytes()
}
