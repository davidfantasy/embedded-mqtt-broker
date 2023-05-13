package packets

import (
	"bytes"
	"fmt"
	"io"
)

//MQTT协议的固定头部协议，每条消息都需要携带
type FixedHeader struct {
	MessageType     byte
	Dup             bool
	Qos             byte
	Retain          bool
	RemainingLength int
}

func (fh *FixedHeader) unpack(typeAndFlags byte, r io.Reader) error {
	fh.MessageType = typeAndFlags >> 4
	fh.Dup = (typeAndFlags>>3)&0x01 > 0
	fh.Qos = (typeAndFlags >> 1) & 0x03
	fh.Retain = typeAndFlags&0x01 > 0

	var err error
	fh.RemainingLength, err = decodeLength(r)
	return err
}

func decodeLength(r io.Reader) (int, error) {
	var rLength uint32
	var multiplier uint32
	b := make([]byte, 1)
	for multiplier < 27 { // fix: Infinite '(digit & 128) == 1' will cause the dead loop
		_, err := io.ReadFull(r, b)
		if err != nil {
			return 0, err
		}

		digit := b[0]
		rLength |= uint32(digit&127) << multiplier
		if (digit & 128) == 0 {
			break
		}
		multiplier += 7
	}
	return int(rLength), nil
}

func (fh *FixedHeader) pack() bytes.Buffer {
	var header bytes.Buffer
	header.WriteByte(fh.MessageType<<4 | boolToByte(fh.Dup)<<3 | fh.Qos<<1 | boolToByte(fh.Retain))
	header.Write(encodeLength(fh.RemainingLength))
	return header
}

func (fh FixedHeader) String() string {
	return fmt.Sprintf("%s: dup: %t qos: %d retain: %t rLength: %d", PacketNames[fh.MessageType], fh.Dup, fh.Qos, fh.Retain, fh.RemainingLength)
}
