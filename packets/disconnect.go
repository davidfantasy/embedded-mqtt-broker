package packets

import "io"

type DisconnectPacket struct {
	FixedHeader
}

func (d *DisconnectPacket) String() string {
	return d.FixedHeader.String()
}

func (d *DisconnectPacket) Write(w io.Writer) error {
	packet := d.FixedHeader.pack()
	_, err := packet.WriteTo(w)

	return err
}

func (d *DisconnectPacket) Read(b io.Reader) error {
	return nil
}
