package packets

import (
	"bytes"
	"fmt"
	"io"
)

//connect包，客户端发起连接时发送
type ConnectPacket struct {
	FixedHeader
	ProtocolName    string
	ProtocolVersion byte
	CleanSession    bool
	WillFlag        bool
	WillQos         byte
	WillRetain      bool
	UsernameFlag    bool
	PasswordFlag    bool
	ReservedBit     byte
	Keepalive       uint16

	ClientId    string
	WillTopic   string
	WillMessage []byte
	Username    string
	Password    []byte
}

func (cp *ConnectPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(encodeString(cp.ProtocolName))
	body.WriteByte(cp.ProtocolVersion)
	body.WriteByte(boolToByte(cp.CleanSession)<<1 | boolToByte(cp.WillFlag)<<2 | cp.WillQos<<3 | boolToByte(cp.WillRetain)<<5 | boolToByte(cp.PasswordFlag)<<6 | boolToByte(cp.UsernameFlag)<<7)
	body.Write(encodeUint16(cp.Keepalive))
	body.Write(encodeString(cp.ClientId))
	if cp.WillFlag {
		body.Write(encodeString(cp.WillTopic))
		body.Write(encodeBytes(cp.WillMessage))
	}
	if cp.UsernameFlag {
		body.Write(encodeString(cp.Username))
	}
	if cp.PasswordFlag {
		body.Write(encodeBytes(cp.Password))
	}
	cp.FixedHeader.RemainingLength = body.Len()
	packet := cp.FixedHeader.pack()
	packet.Write(body.Bytes())
	_, err = packet.WriteTo(w)
	return err
}

func (cp *ConnectPacket) Read(b io.Reader) error {
	var err error
	cp.ProtocolName, err = readString(b)
	if err != nil {
		return err
	}
	cp.ProtocolVersion, err = readByte(b)
	if err != nil {
		return err
	}
	options, err := readByte(b)
	if err != nil {
		return err
	}
	cp.ReservedBit = 1 & options
	cp.CleanSession = 1&(options>>1) > 0
	cp.WillFlag = 1&(options>>2) > 0
	cp.WillQos = 3 & (options >> 3)
	cp.WillRetain = 1&(options>>5) > 0
	cp.PasswordFlag = 1&(options>>6) > 0
	cp.UsernameFlag = 1&(options>>7) > 0
	cp.Keepalive, err = readUint16(b)
	if err != nil {
		return err
	}
	cp.ClientId, err = readString(b)
	if err != nil {
		return err
	}
	if cp.WillFlag {
		cp.WillTopic, err = readString(b)
		if err != nil {
			return err
		}
		cp.WillMessage, err = readBytes(b)
		if err != nil {
			return err
		}
	}
	if cp.UsernameFlag {
		cp.Username, err = readString(b)
		if err != nil {
			return err
		}
	}
	if cp.PasswordFlag {
		cp.Password, err = readBytes(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cp *ConnectPacket) String() string {
	var password string
	if len(cp.Password) > 0 {
		password = "<redacted>"
	}
	return fmt.Sprintf("ConnectPacket: %s protocolversion: %d protocolname: %s cleansession: %t willflag: %t WillQos: %d WillRetain: %t Usernameflag: %t Passwordflag: %t keepalive: %d clientId: %s willtopic: %s willmessage: %s Username: %s Password: %s", cp.FixedHeader, cp.ProtocolVersion, cp.ProtocolName, cp.CleanSession, cp.WillFlag, cp.WillQos, cp.WillRetain, cp.UsernameFlag, cp.PasswordFlag, cp.Keepalive, cp.ClientId, cp.WillTopic, cp.WillMessage, cp.Username, password)
}
