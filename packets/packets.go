package packets

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

// Mqtt消息的包类型
const (
	Connect     = 1
	Connack     = 2
	Publish     = 3
	Puback      = 4
	Pubrec      = 5
	Pubrel      = 6
	Pubcomp     = 7
	Subscribe   = 8
	Suback      = 9
	Unsubscribe = 10
	Unsuback    = 11
	Pingreq     = 12
	Pingresp    = 13
	Disconnect  = 14
)

var PacketNames = map[uint8]string{
	1:  "CONNECT",
	2:  "CONNACK",
	3:  "PUBLISH",
	4:  "PUBACK",
	5:  "PUBREC",
	6:  "PUBREL",
	7:  "PUBCOMP",
	8:  "SUBSCRIBE",
	9:  "SUBACK",
	10: "UNSUBSCRIBE",
	11: "UNSUBACK",
	12: "PINGREQ",
	13: "PINGRESP",
	14: "DISCONNECT",
}

//所有packet的公共接口
type MqttPacket interface {
	Write(io.Writer) error
	Read(io.Reader) error
	String() string
}

const (
	Accepted                        = 0x00
	ErrRefusedBadProtocolVersion    = 0x01
	ErrRefusedIDRejected            = 0x02
	ErrRefusedServerUnavailable     = 0x03
	ErrRefusedBadUsernameOrPassword = 0x04
	ErrRefusedNotAuthorised         = 0x05
	ErrNetworkError                 = 0xFE
	ErrProtocolViolation            = 0xFF
)

func ReadPacket(conn net.Conn) (MqttPacket, error) {
	var fh FixedHeader
	b := make([]byte, 1)

	_, err := io.ReadFull(conn, b)
	if err != nil {
		return nil, err
	}

	err = fh.unpack(b[0], conn)
	if err != nil {
		return nil, err
	}
	packet, err := NewMqttPacketWithHeader(fh)
	if err != nil {
		return nil, err
	}
	packetBytes := make([]byte, fh.RemainingLength)
	n, err := io.ReadFull(conn, packetBytes)
	if err != nil {
		return nil, err
	}
	if n != fh.RemainingLength {
		return nil, fmt.Errorf("failed to read expected data,expect len:%v,actual:%v", fh.RemainingLength, n)
	}
	err = packet.Read(bytes.NewBuffer(packetBytes))
	return packet, err
}

func NewMqttPacket(messageType byte) MqttPacket {
	switch messageType {
	case Connect:
		return &ConnectPacket{FixedHeader: FixedHeader{MessageType: Connect}}
	case Connack:
		return &ConnackPacket{FixedHeader: FixedHeader{MessageType: Connack}}
	case Publish:
		return &PublishPacket{FixedHeader: FixedHeader{MessageType: Publish}}
	case Subscribe:
		return &SubscribePacket{FixedHeader: FixedHeader{MessageType: Subscribe}}
	case Suback:
		return &SubackPacket{FixedHeader: FixedHeader{MessageType: Suback}}
	case Unsubscribe:
		return &UnsubscribePacket{FixedHeader: FixedHeader{MessageType: Unsubscribe}}
	case Unsuback:
		return &UnsubackPacket{FixedHeader: FixedHeader{MessageType: Unsuback}}
	case Disconnect:
		return &DisconnectPacket{FixedHeader: FixedHeader{MessageType: Disconnect}}
	case Pingreq:
		return &DisconnectPacket{FixedHeader: FixedHeader{MessageType: Pingreq}}
	case Pingresp:
		return &DisconnectPacket{FixedHeader: FixedHeader{MessageType: Pingresp}}
	}
	return nil
}

func NewMqttPacketWithHeader(fh FixedHeader) (MqttPacket, error) {
	switch fh.MessageType {
	case Connect:
		return &ConnectPacket{FixedHeader: fh}, nil
	case Connack:
		return &ConnectPacket{FixedHeader: fh}, nil
	case Publish:
		return &PublishPacket{FixedHeader: fh}, nil
	case Subscribe:
		return &SubscribePacket{FixedHeader: fh}, nil
	case Suback:
		return &SubackPacket{FixedHeader: fh}, nil
	case Unsubscribe:
		return &UnsubscribePacket{FixedHeader: fh}, nil
	case Unsuback:
		return &UnsubackPacket{FixedHeader: fh}, nil
	case Disconnect:
		return &DisconnectPacket{FixedHeader: fh}, nil
	case Pingreq:
		return &PingreqPacket{FixedHeader: fh}, nil
	case Pingresp:
		return &PingrespPacket{FixedHeader: fh}, nil
	}

	return nil, fmt.Errorf("unsupported packet type 0x%x", fh.MessageType)
}
