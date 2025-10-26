// Implementation of RFC 6143 §7.3 Initialization Messages.

package vnc

import (
	"io"

	"github.com/kward/go-vnc/logging"
	"github.com/kward/go-vnc/rfbflags"
)

// clientInit implements §7.3.1 ClientInit.
func (c *ClientConn) clientInit() error {
	if logging.V(logging.FnDeclLevel) {
		logging.Infof("%s", logging.FnName())
	}

	sharedFlag := rfbflags.BoolToRFBFlag(!c.config.Exclusive)
	if logging.V(logging.ResultLevel) {
		logging.Infof("sharedFlag: %d", sharedFlag)
	}
	if err := c.send(sharedFlag); err != nil {
		return err
	}

	// TODO(kward)20170226): VENUE responds with some sort of shared flag
	// response, which includes the VENUE name and IPs. Handle this?

	return nil
}

// ServerInit message sent after server receives a ClientInit message.
// https://tools.ietf.org/html/rfc6143#section-7.3.2
type ServerInit struct {
	FBWidth, FBHeight uint16
	PixelFormat       PixelFormat
	NameLength        uint32
	// Name is of variable length, and must be read separately.
}

const serverInitLen = 24 // Not including Name.

// Verify that interfaces are honored.
var _ Unmarshaler = (*ServerInit)(nil)

// Read implements
func (m *ServerInit) Read(r io.Reader) error {
	buf := make([]byte, serverInitLen)
	if _, err := io.ReadAtLeast(r, buf, serverInitLen); err != nil {
		return err
	}
	return m.Unmarshal(buf)
}

func (m *ServerInit) Unmarshal(data []byte) error {
	buf := NewBuffer(data)
	var msg ServerInit
	if err := buf.Read(&msg); err != nil {
		return err
	}
	*m = msg
	return nil
}

// serverInit implements §7.3.2 ServerInit.
func (c *ClientConn) serverInit() error {
	if logging.V(logging.FnDeclLevel) {
		logging.Infof("%s", logging.FnName())
	}

	var msg ServerInit
	if err := msg.Read(c.c); err != nil {
		return Errorf("failure reading ServerInit message; %v", err)
	}
	if logging.V(logging.ResultLevel) {
		logging.Infof("ServerInit message: %v", msg)
	}

	c.setFramebufferWidth(msg.FBWidth)
	c.setFramebufferHeight(msg.FBHeight)
	c.pixelFormat = msg.PixelFormat

	name := make([]uint8, msg.NameLength)
	if err := c.receive(&name); err != nil {
		return err
	}
	c.setDesktopName(string(name))

	return nil
}
