// Implementation of §7.5 Client-to-Server Messages.

package vnc

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/kward/go-vnc/buttons"
	"github.com/kward/go-vnc/encoding"
	"github.com/kward/go-vnc/encodings"
	"github.com/kward/go-vnc/keys"
	"github.com/kward/go-vnc/logging"
	"github.com/kward/go-vnc/messages"
)

// SetPixelFormatMessage holds the wire format message.
type SetPixelFormatMessage struct {
	Msg messages.ClientMessage // message-type
	_   [3]byte                // padding
	PF  PixelFormat            // pixel-format
}

// SetPixelFormat sets the format in which pixel values should be sent
// in FramebufferUpdate messages from the server.
//
// See RFC 6143 Section 7.5.1
func (c *ClientConn) SetPixelFormat(pf PixelFormat) error {
	if logging.V(logging.FnDeclLevel) {
		logging.Infof("ClientConn.%s", logging.FnNameWithArgs("%s", pf))
	}

	pfWire := encoding.PixelFormatWire{
		BPP:        pf.BPP,
		Depth:      pf.Depth,
		BigEndian:  boolToUint8(pf.BigEndian),
		TrueColor:  boolToUint8(pf.TrueColor),
		RedMax:     pf.RedMax,
		GreenMax:   pf.GreenMax,
		BlueMax:    pf.BlueMax,
		RedShift:   pf.RedShift,
		GreenShift: pf.GreenShift,
		BlueShift:  pf.BlueShift,
	}
	wmsg := encoding.SetPixelFormatRequestWire{
		MsgType: byte(messages.SetPixelFormat),
		PF:      pfWire,
	}
	encBytes, err := wmsg.WireMarshal()
	if err != nil {
		return err
	}
	if err := c.send(encBytes); err != nil {
		return err
	}

	// Invalidate the color map.
	if !pf.TrueColor {
		c.colorMap = [256]Color{}
	}

	c.pixelFormat = pf
	return nil
}

// SetEncodingsMessage holds the wire format message, sans encoding-type field.
type SetEncodingsMessage struct {
	Msg     messages.ClientMessage // message-type
	_       [1]byte                // padding
	NumEncs uint16                 // number-of-encodings
}

// SetEncodings sets the encoding types in which the pixel data can be sent
// from the server. After calling this method, the encs slice given should not
// be modified.
//
// TODO(kward:20170306) Fix bad practice of mixing of protocol and internal
// state here.
//
// See RFC 6143 Section 7.5.2
func (c *ClientConn) SetEncodings(encs Encodings) error {
	if logging.V(logging.FnDeclLevel) {
		logging.Infof("ClientConn.%s", logging.FnNameWithArgs("%s", encs))
	}

	// Make sure RawEncoding is supported.
	haveRaw := false
	for _, v := range encs {
		if v.Type() == encodings.Raw {
			haveRaw = true
			break
		}
	}
	if !haveRaw {
		encs = append(encs, &RawEncoding{})
	}

	// Prepare message using encoding package.
	wmsg := encoding.SetEncodingsRequestWire{
		MsgType:   byte(messages.SetEncodings),
		NumEnc:    uint16(len(encs)),
		Encodings: encsToEncodingTypes(encs),
	}
	encBytes, err := wmsg.WireMarshal()
	if err != nil {
		return err
	}
	if err := c.send(encBytes); err != nil {
		return err
	}

	c.encodings = encs
	return nil
}

// FramebufferUpdateRequestMessage holds the wire format message.
type FramebufferUpdateRequestMessage struct {
	Msg           messages.ClientMessage // message-type
	Inc           bool                   // incremental
	X, Y          uint16                 // x-, y-position
	Width, Height uint16                 // width, height
}

// Requests a framebuffer update from the server. There may be an indefinite
// time between the request and the actual framebuffer update being received.
//
// See RFC 6143 Section 7.5.3
func (c *ClientConn) FramebufferUpdateRequest(inc bool, x, y, w, h uint16) error {
	wmsg := encoding.FramebufferUpdateRequestWire{
		MsgType:     byte(messages.FramebufferUpdateRequest),
		Incremental: boolToUint8(inc),
		X:           x,
		Y:           y,
		Width:       w,
		Height:      h,
	}
	b, err := wmsg.WireMarshal()
	if err != nil {
		return err
	}
	return c.send(b)
}

// KeyEventMessage holds the wire format message.
type KeyEventMessage struct {
	Msg      messages.ClientMessage // message-type
	DownFlag bool                   // down-flag
	_        [2]byte                // padding
	Key      keys.Key               // key
}

const (
	PressKey   = true
	ReleaseKey = false
)

// KeyEvent indicates a key press or release and sends it to the server.
// The key is indicated using the X Window System "keysym" value. Constants are
// provided in `keys/keys.go`. To simulate a key press, you must send a key with
// both a true and false down event.
//
// See RFC 6143 Section 7.5.4.
func (c *ClientConn) KeyEvent(key keys.Key, down bool) error {
	if logging.V(logging.FnDeclLevel) {
		logging.Infof("ClientConnt.%s", logging.FnNameWithArgs("%s, %t", key, down))
	}

	wmsg := encoding.KeyEventWire{
		MsgType:  byte(messages.KeyEvent),
		DownFlag: boolToUint8(down),
		Key:      int32(key),
	}
	encPayload, err := wmsg.WireMarshal()
	if err != nil {
		return err
	}
	if err := c.send(encPayload); err != nil {
		return err
	}

	settleUI()
	return nil
}

// PointerEventMessage holds the wire format message.
type PointerEventMessage struct {
	Msg  messages.ClientMessage // message-type
	Mask uint8                  // button-mask
	X, Y uint16                 // x-, y-position
}

// PointerEvent indicates that pointer movement or a pointer button
// press or release.
//
// The `button` is a bitwise mask of various Button values. When a button
// is set, it is pressed, when it is unset, it is released.
//
// See RFC 6143 Section 7.5.5
func (c *ClientConn) PointerEvent(button buttons.Button, x, y uint16) error {
	if logging.V(logging.FnDeclLevel) {
		logging.Infof("%s", logging.FnNameWithArgs("%s, %d, %d", button, x, y))
	}

	wmsg := encoding.PointerEventWire{
		MsgType:    byte(messages.PointerEvent),
		ButtonMask: uint8(button),
		X:          x,
		Y:          y,
	}
	encPayload, err := wmsg.WireMarshal()
	if err != nil {
		return err
	}
	if err := c.send(encPayload); err != nil {
		return err
	}

	settleUI()
	return nil
}

// ClientCutTextMessage holds the wire format message, sans the text field.
type ClientCutTextMessage struct {
	Msg    messages.ClientMessage // message-type
	_      [3]byte                // padding
	Length uint32                 // length
}

// ClientCutText tells the server that the client has new text in its cut buffer.
// The text string MUST only contain Latin-1 characters. This encoding
// is compatible with Go's native string format, but can only use up to
// unicode.MaxLatin1 values.
//
// See RFC 6143 Section 7.5.6
func (c *ClientConn) ClientCutText(text string) error {
	if logging.V(logging.FnDeclLevel) {
		logging.Infof("%s", logging.FnNameWithArgs("%s", text))
	}

	for _, char := range text {
		if char > unicode.MaxLatin1 {
			return NewVNCError(fmt.Sprintf("Character %q is not valid Latin-1", char))
		}
	}

	// Strip carriage-return (0x0d) chars.
	// From RFC: "Ends of lines are represented by the newline character (0x0a)
	// alone. No carriage-return (0x0d) is used."
	text = strings.Join(strings.Split(text, "\r"), "")

	wmsg := encoding.ClientCutTextWire{
		MsgType: byte(messages.ClientCutText),
		Text:    text,
	}
	data, err := wmsg.WireMarshal()
	if err != nil {
		return err
	}
	if err := c.send(data); err != nil {
		return err
	}

	settleUI()
	return nil
}

func encsToEncodingTypes(encs Encodings) []encodings.Encoding {
	types := make([]encodings.Encoding, 0, len(encs))
	for _, e := range encs {
		types = append(types, e.Type())
	}
	return types
}
