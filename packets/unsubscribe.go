package packets

import (
	"bytes"
	"fmt"
	"io"
)

type UnsubscribePacket struct {
	FixedHeader
	MessageID uint16
	Topics    []string
}

func (u *UnsubscribePacket) String() string {
	return fmt.Sprintf("%s MessageID: %d", u.FixedHeader, u.MessageID)
}

func (u *UnsubscribePacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(encodeUint16(u.MessageID))
	for _, topic := range u.Topics {
		body.Write(encodeString(topic))
	}
	u.FixedHeader.RemainingLength = body.Len()
	packet := u.FixedHeader.pack()
	packet.Write(body.Bytes())
	_, err = packet.WriteTo(w)
	return err
}

func (u *UnsubscribePacket) Read(b io.Reader) error {
	var err error
	u.MessageID, err = decodeUint16(b)
	if err != nil {
		return err
	}

	for topic, err := decodeString(b); err == nil && topic != ""; topic, err = decodeString(b) {
		u.Topics = append(u.Topics, topic)
	}

	return err
}
