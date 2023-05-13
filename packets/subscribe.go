package packets

import (
	"bytes"
	"fmt"
	"io"
)

type SubscribePacket struct {
	FixedHeader
	MessageID uint16
	Topics    []string
	Qoss      []byte
}

func (s *SubscribePacket) String() string {
	return fmt.Sprintf("%s MessageID: %d topics: %s", s.FixedHeader, s.MessageID, s.Topics)
}

func (s *SubscribePacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error

	body.Write(encodeUint16(s.MessageID))
	for i, topic := range s.Topics {
		body.Write(encodeString(topic))
		body.WriteByte(s.Qoss[i])
	}
	s.FixedHeader.RemainingLength = body.Len()
	packet := s.FixedHeader.pack()
	packet.Write(body.Bytes())
	_, err = packet.WriteTo(w)

	return err
}

func (s *SubscribePacket) Read(b io.Reader) error {
	var err error
	s.MessageID, err = decodeUint16(b)
	if err != nil {
		return err
	}
	payloadLength := s.FixedHeader.RemainingLength - 2
	for payloadLength > 0 {
		topic, err := decodeString(b)
		if err != nil {
			return err
		}
		s.Topics = append(s.Topics, topic)
		qos, err := decodeByte(b)
		if err != nil {
			return err
		}
		s.Qoss = append(s.Qoss, qos)
		payloadLength -= 2 + len(topic) + 1 // 2 bytes of string length, plus string, plus 1 byte for Qos
	}

	return nil
}
