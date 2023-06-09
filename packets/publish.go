package packets

import (
	"bytes"
	"fmt"
	"io"
)

// PublishPacket is an internal representation of the fields of the
// Publish MQTT packet
type PublishPacket struct {
	FixedHeader
	TopicName string
	MessageID uint16
	Payload   []byte
}

func (p *PublishPacket) String() string {
	return fmt.Sprintf("%s topicName: %s MessageID: %d payload: %s", p.FixedHeader, p.TopicName, p.MessageID, string(p.Payload))
}

func (p *PublishPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error

	body.Write(encodeString(p.TopicName))
	if p.Qos > 0 {
		body.Write(encodeUint16(p.MessageID))
	}
	p.FixedHeader.RemainingLength = body.Len() + len(p.Payload)
	packet := p.FixedHeader.pack()
	packet.Write(body.Bytes())
	packet.Write(p.Payload)
	_, err = w.Write(packet.Bytes())

	return err
}

// Unpack decodes the details of a ControlPacket after the fixed
// header has been read
func (p *PublishPacket) Read(b io.Reader) error {
	var payloadLength = p.FixedHeader.RemainingLength
	var err error
	p.TopicName, err = readString(b)
	if err != nil {
		return err
	}

	if p.Qos > 0 {
		p.MessageID, err = readUint16(b)
		if err != nil {
			return err
		}
		payloadLength -= len(p.TopicName) + 4
	} else {
		payloadLength -= len(p.TopicName) + 2
	}
	if payloadLength < 0 {
		return fmt.Errorf("error unpacking publish, payload length < 0")
	}
	p.Payload = make([]byte, payloadLength)
	_, err = b.Read(p.Payload)

	return err
}

// Copy creates a new PublishPacket with the same topic and payload
// but an empty fixed header, useful for when you want to deliver
// a message with different properties such as Qos but the same
// content
func (p *PublishPacket) Copy() *PublishPacket {
	newP := NewMqttPacket(Publish).(*PublishPacket)
	newP.TopicName = p.TopicName
	newP.Payload = p.Payload
	return newP
}
