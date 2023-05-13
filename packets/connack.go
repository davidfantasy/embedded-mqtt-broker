package packets

import (
	"bytes"
	"fmt"
	"io"
)

//connack包，服务端对connect的回应
type ConnackPacket struct {
	FixedHeader
	SessionPresent bool
	ReturnCode     byte
}

func (ca *ConnackPacket) String() string {
	return fmt.Sprintf("%s sessionpresent: %t returncode: %d", ca.FixedHeader, ca.SessionPresent, ca.ReturnCode)
}

func (ca *ConnackPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.WriteByte(boolToByte(ca.SessionPresent))
	body.WriteByte(ca.ReturnCode)
	ca.FixedHeader.RemainingLength = 2
	packet := ca.FixedHeader.pack()
	packet.Write(body.Bytes())
	_, err = packet.WriteTo(w)
	return err
}

func (ca *ConnackPacket) Read(b io.Reader) error {
	flags, err := readByte(b)
	if err != nil {
		return err
	}
	ca.SessionPresent = 1&flags > 0
	ca.ReturnCode, err = readByte(b)
	return err
}
