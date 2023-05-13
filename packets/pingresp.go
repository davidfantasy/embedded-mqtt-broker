package packets

import "io"

type PingrespPacket struct {
	FixedHeader
}

func (pr *PingrespPacket) String() string {
	return pr.FixedHeader.String()
}

func (pr *PingrespPacket) Write(w io.Writer) error {
	packet := pr.FixedHeader.pack()
	_, err := packet.WriteTo(w)

	return err
}

func (pr *PingrespPacket) Read(b io.Reader) error {
	return nil
}
