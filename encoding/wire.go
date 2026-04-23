// Package encoding provides thin VNC wire format marshaling/unmarshaling
// between big-endian byte streams and typed message structures.
//
// All VNC wire format is big-endian. This package reads/writes between
// the typed structs and big-endian byte streams without protoc dependency,
// matching proto/vnc.proto exactly field-for-field.
package encoding

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/kward/go-vnc/encodings"
)

// ---- Message type codes (for WireMarshal on zero-length messages) ----

const (
	msgSetPixelFormat byte = 0
	_
	msgSetEncodings byte = 2
	_
	msgFramebufferUpdateRequest byte = 3
	msgKeyEvent                 byte = 4
	msgPointerEvent             byte = 5
	msgClientCutText            byte = 6

	msgFramebufferUpdate  byte = 0
	msgSetColorMapEntries byte = 1
	msgBell               byte = 2
	msgServerCutText      byte = 3
)

// wireMessage is implemented by any message type that knows its own wire format.
type wireMessage interface {
	WireMarshal() ([]byte, error)
	WireUnmarshal(data []byte) error
}

func writeWire(buf *bytes.Buffer, msg wireMessage) error {
	b, err := msg.WireMarshal()
	if err != nil {
		return err
	}
	_, err = buf.Write(b)
	return err
}

// MarshalToWire serializes a wireMessage into big-endian VNC wire format bytes.
func MarshalToWire(m wireMessage) ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := writeWire(buf, m); err != nil {
		return nil, fmt.Errorf("wire marshal: %w", err)
	}
	return buf.Bytes(), nil
}

// UnmarshalFromWire deserializes big-endian VNC wire format bytes into a wireMessage.
func UnmarshalFromWire(data []byte, m wireMessage) error {
	return m.WireUnmarshal(data)
}

// ---- PixelFormatWire (§7.4) ----
// Wire: bpp(1) depth(1) big_endian(1) true_color(1)
//       red_max(2 le) green_max(2 le) blue_max(2 le)
//       red_shift(1) green_shift(1) blue_shift(1) pad(3)

type PixelFormatWire struct {
	BPP        uint8
	Depth      uint8
	BigEndian  uint8
	TrueColor  uint8
	RedMax     uint16
	GreenMax   uint16
	BlueMax    uint16
	RedShift   uint8
	GreenShift uint8
	BlueShift  uint8
	Padding    [3]byte
}

func (w *PixelFormatWire) marshalFields() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte(w.BPP)
	buf.WriteByte(w.Depth)
	buf.WriteByte(w.BigEndian)
	buf.WriteByte(w.TrueColor)
	binary.Write(buf, binary.LittleEndian, w.RedMax)
	binary.Write(buf, binary.LittleEndian, w.GreenMax)
	binary.Write(buf, binary.LittleEndian, w.BlueMax)
	buf.WriteByte(w.RedShift)
	buf.WriteByte(w.GreenShift)
	buf.WriteByte(w.BlueShift)
	buf.Write(w.Padding[:])
	return buf.Bytes(), nil
}

func (w *PixelFormatWire) WireMarshal() ([]byte, error) {
	return w.marshalFields()
}

func (w *PixelFormatWire) WireUnmarshal(data []byte) error {
	if len(data) < 16 {
		return fmt.Errorf("pixel format: expected 16 bytes, got %d", len(data))
	}
	w.BPP = data[0]
	w.Depth = data[1]
	w.BigEndian = data[2]
	w.TrueColor = data[3]
	w.RedMax = binary.LittleEndian.Uint16(data[4:6])
	w.GreenMax = binary.LittleEndian.Uint16(data[6:8])
	w.BlueMax = binary.LittleEndian.Uint16(data[8:10])
	w.RedShift = data[10]
	w.GreenShift = data[11]
	w.BlueShift = data[12]
	copy(w.Padding[:], data[13:16])
	return nil
}

// ---- SetPixelFormatRequestWire (§7.5.1) ----
// Wire: msg_type(1) + pad(3) + pixel_format(16)

type SetPixelFormatRequestWire struct {
	MsgType byte
	PF      PixelFormatWire
}

func (w *SetPixelFormatRequestWire) WireMarshal() ([]byte, error) {
	b, err := w.PF.marshalFields()
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	buf.WriteByte(w.MsgType)
	buf.Write([]byte{0, 0, 0}) // padding
	buf.Write(b)
	return buf.Bytes(), nil
}

func (w *SetPixelFormatRequestWire) WireUnmarshal(data []byte) error {
	if len(data) < 20 {
		return fmt.Errorf("SetPixelFormatRequest: expected 20 bytes, got %d", len(data))
	}
	w.MsgType = data[0]
	return w.PF.WireUnmarshal(data[4:])
}

// ---- SetEncodingsRequestWire (§7.5.2) ----
// Wire: msg_type(1) + pad(1) + num_encodings(2) + [encoding_type(4)]xN

type SetEncodingsRequestWire struct {
	MsgType   byte
	NumEnc    uint16
	Encodings []encodings.Encoding
}

func (w *SetEncodingsRequestWire) WireMarshal() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte(w.MsgType)
	buf.WriteByte(0) // padding
	binary.Write(buf, binary.BigEndian, w.NumEnc)
	for _, enc := range w.Encodings {
		e := int32(enc)
		binary.Write(buf, binary.BigEndian, e)
	}
	return buf.Bytes(), nil
}

func (w *SetEncodingsRequestWire) WireUnmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("SetEncodingsRequest: expected at least 4 bytes, got %d", len(data))
	}
	w.MsgType = data[0]
	w.NumEnc = binary.BigEndian.Uint16(data[2:4])
	if len(data) < 4+int(w.NumEnc)*4 {
		return fmt.Errorf("SetEncodingsRequest: expected %d bytes, got %d", 4+int(w.NumEnc)*4, len(data))
	}
	w.Encodings = make([]encodings.Encoding, w.NumEnc)
	for i := uint16(0); i < w.NumEnc; i++ {
		e := int32(binary.BigEndian.Uint32(data[4+i*4 : 4+(i+1)*4]))
		w.Encodings[i] = encodings.Encoding(e)
	}
	return nil
}

// ---- FramebufferUpdateRequestWire (§7.5.3) ----
// Wire: msg_type(1) + incremental(1) + x(2) + y(2) + width(2) + height(2)

type FramebufferUpdateRequestWire struct {
	MsgType       byte
	Incremental   byte
	X, Y          uint16
	Width, Height uint16
}

func (w *FramebufferUpdateRequestWire) WireMarshal() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte(w.MsgType)
	buf.WriteByte(w.Incremental)
	binary.Write(buf, binary.BigEndian, w.X)
	binary.Write(buf, binary.BigEndian, w.Y)
	binary.Write(buf, binary.BigEndian, w.Width)
	binary.Write(buf, binary.BigEndian, w.Height)
	return buf.Bytes(), nil
}

func (w *FramebufferUpdateRequestWire) WireUnmarshal(data []byte) error {
	if len(data) < 10 {
		return fmt.Errorf("FramebufferUpdateRequest: expected 10 bytes, got %d", len(data))
	}
	w.MsgType = data[0]
	w.Incremental = data[1]
	w.X = binary.BigEndian.Uint16(data[2:4])
	w.Y = binary.BigEndian.Uint16(data[4:6])
	w.Width = binary.BigEndian.Uint16(data[6:8])
	w.Height = binary.BigEndian.Uint16(data[8:10])
	return nil
}

// ---- KeyEventWire (§7.5.4) ----
// Wire: msg_type(1) + down_flag(1) + pad(2) + key(4)

type KeyEventWire struct {
	MsgType  byte
	DownFlag byte
	Key      int32
}

func (w *KeyEventWire) WireMarshal() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte(w.MsgType)
	buf.WriteByte(w.DownFlag)
	buf.Write([]byte{0, 0}) // padding
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(w.Key))
	buf.Write(b)
	return buf.Bytes(), nil
}

func (w *KeyEventWire) WireUnmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("KeyEvent: expected 8 bytes, got %d", len(data))
	}
	w.MsgType = data[0]
	w.DownFlag = data[1]
	w.Key = int32(binary.BigEndian.Uint32(data[4:8]))
	return nil
}

// ---- PointerEventWire (§7.5.5) ----
// Wire: msg_type(1) + button_mask(1) + x(2) + y(2)

type PointerEventWire struct {
	MsgType    byte
	ButtonMask byte
	X, Y       uint16
}

func (w *PointerEventWire) WireMarshal() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte(w.MsgType)
	buf.WriteByte(w.ButtonMask)
	binary.Write(buf, binary.BigEndian, w.X)
	binary.Write(buf, binary.BigEndian, w.Y)
	return buf.Bytes(), nil
}

func (w *PointerEventWire) WireUnmarshal(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("PointerEvent: expected 6 bytes, got %d", len(data))
	}
	w.MsgType = data[0]
	w.ButtonMask = data[1]
	w.X = binary.BigEndian.Uint16(data[2:4])
	w.Y = binary.BigEndian.Uint16(data[4:6])
	return nil
}

// ---- ClientCutTextWire (§7.5.6) ----
// Wire: msg_type(1) + pad(3) + length(4) + text[N]

type ClientCutTextWire struct {
	MsgType byte
	Length  uint32
	Text    string
}

func (w *ClientCutTextWire) WireMarshal() ([]byte, error) {
	w.Length = uint32(len(w.Text))
	buf := &bytes.Buffer{}
	buf.WriteByte(w.MsgType)
	buf.Write([]byte{0, 0, 0}) // padding
	binary.Write(buf, binary.BigEndian, w.Length)
	buf.WriteString(w.Text)
	return buf.Bytes(), nil
}

func (w *ClientCutTextWire) WireUnmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("ClientCutText: expected 8 bytes, got %d", len(data))
	}
	w.MsgType = data[0]
	w.Length = binary.BigEndian.Uint32(data[4:8])
	if len(data) < int(8+w.Length) {
		return fmt.Errorf("ClientCutText: expected %d bytes, got %d", 8+w.Length, len(data))
	}
	w.Text = string(data[8 : 8+w.Length])
	return nil
}

// ---- RectDataWire (rectangle inside FramebufferUpdate §7.6.1) ----
// Wire: x(2) + y(2) + width(2) + height(2) + encoding(4) + [payload]

type RectDataWire struct {
	X, Y          uint16
	Width, Height uint16
	Encoding      encodings.Encoding
	Payload       []byte
}

func (w *RectDataWire) WireMarshal() ([]byte, error) {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, w.X)
	binary.Write(buf, binary.BigEndian, w.Y)
	binary.Write(buf, binary.BigEndian, w.Width)
	binary.Write(buf, binary.BigEndian, w.Height)
	enc := int32(w.Encoding)
	binary.Write(buf, binary.BigEndian, enc)
	buf.Write(w.Payload)
	return buf.Bytes(), nil
}

func (w *RectDataWire) WireUnmarshal(data []byte) error {
	if len(data) < 12 {
		return fmt.Errorf("rectangle: expected 12 bytes, got %d", len(data))
	}
	w.X = binary.BigEndian.Uint16(data[0:2])
	w.Y = binary.BigEndian.Uint16(data[2:4])
	w.Width = binary.BigEndian.Uint16(data[4:6])
	w.Height = binary.BigEndian.Uint16(data[6:8])
	w.Encoding = encodings.Encoding(int32(binary.BigEndian.Uint32(data[8:12])))
	if len(data) > 12 {
		w.Payload = data[12:]
	}
	return nil
}

// ---- FramebufferUpdateWire (§7.6.1) ----
// Wire: msg_type(1) + pad(1) + num_rectangles(2) + [rect][N]

type FramebufferUpdateWire struct {
	MsgType  byte
	NumRects uint16
	RectData []RectDataWire
}

func (w *FramebufferUpdateWire) WireMarshal() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte(w.MsgType)
	buf.WriteByte(0) // padding
	binary.Write(buf, binary.BigEndian, w.NumRects)
	for _, rect := range w.RectData {
		b, err := rect.WireMarshal()
		if err != nil {
			return nil, err
		}
		buf.Write(b)
	}
	return buf.Bytes(), nil
}

func (w *FramebufferUpdateWire) WireUnmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("FramebufferUpdate: expected at least 4 bytes, got %d", len(data))
	}
	w.NumRects = binary.BigEndian.Uint16(data[2:4])
	rest := data[4:]
	for i := uint16(0); i < w.NumRects; i++ {
		if len(rest) < 12 {
			return fmt.Errorf("FramebufferUpdate: rectangle %d: need 12 bytes, got %d", i, len(rest))
		}
		rect := RectDataWire{}
		if err := rect.WireUnmarshal(rest); err != nil {
			return err
		}
		w.RectData = append(w.RectData, rect)
		payloadLen := 12 + len(rect.Payload)
		rest = rest[payloadLen:]
	}
	return nil
}

// ---- SetColorMapEntriesWire (§7.6.2) ----
// Wire: msg_type(1) + pad(1) + first_color(2) + num_colors(2) + [pad(1) red(2) green(2) blue(2)]xN

type SetColorMapEntriesWire struct {
	MsgType    byte
	FirstColor uint16
	NumColors  uint16
	Colors     []ColorEntryWire
}

type ColorEntryWire struct {
	Red   uint16
	Green uint16
	Blue  uint16
}

func (w *SetColorMapEntriesWire) WireMarshal() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte(w.MsgType)
	buf.WriteByte(0) // padding
	binary.Write(buf, binary.BigEndian, w.FirstColor)
	binary.Write(buf, binary.BigEndian, w.NumColors)
	for i := uint16(0); i < w.NumColors && i < uint16(len(w.Colors)); i++ {
		c := w.Colors[i]
		buf.WriteByte(0) // per-color pad
		binary.Write(buf, binary.BigEndian, c.Red)
		binary.Write(buf, binary.BigEndian, c.Green)
		binary.Write(buf, binary.BigEndian, c.Blue)
	}
	return buf.Bytes(), nil
}

func (w *SetColorMapEntriesWire) WireUnmarshal(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("SetColorMapEntries: expected 6 bytes, got %d", len(data))
	}
	w.MsgType = data[0]
	w.FirstColor = binary.BigEndian.Uint16(data[2:4])
	w.NumColors = binary.BigEndian.Uint16(data[4:6])
	for i := uint16(0); i < w.NumColors; i++ {
		off := 6 + i*6
		if int(off)+6 > len(data) {
			return fmt.Errorf("SetColorMapEntries: color %d: insufficient data", i)
		}
		c := ColorEntryWire{
			Red:   binary.BigEndian.Uint16(data[off+1 : off+3]),
			Green: binary.BigEndian.Uint16(data[off+3 : off+5]),
			Blue:  binary.BigEndian.Uint16(data[off+5 : off+7]),
		}
		w.Colors = append(w.Colors, c)
	}
	return nil
}

// ---- BellWire (§7.6.3) ----
// Wire: msg_type(1) [no body]

type BellWire struct{}

func (w *BellWire) WireMarshal() ([]byte, error) {
	return []byte{msgBell}, nil
}

func (w *BellWire) WireUnmarshal(data []byte) error {
	if len(data) < 1 {
		return fmt.Errorf("Bell: expected at least 1 byte, got %d", len(data))
	}
	return nil
}

// ---- ServerCutTextWire (§7.6.4) ----
// Wire: msg_type(1) + pad(1) + length(4) + text[N]

type ServerCutTextWire struct {
	MsgType byte
	Length  uint32
	Text    string
}

func (w *ServerCutTextWire) WireMarshal() ([]byte, error) {
	w.Length = uint32(len(w.Text))
	buf := &bytes.Buffer{}
	buf.WriteByte(w.MsgType)
	buf.WriteByte(0) // padding
	binary.Write(buf, binary.BigEndian, w.Length)
	buf.WriteString(w.Text)
	return buf.Bytes(), nil
}

func (w *ServerCutTextWire) WireUnmarshal(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("ServerCutText: expected 6 bytes, got %d", len(data))
	}
	w.MsgType = data[0]
	w.Length = binary.BigEndian.Uint32(data[4:8])
	if len(data) < int(8+w.Length) {
		return fmt.Errorf("ServerCutText: expected %d bytes, got %d", 8+w.Length, len(data))
	}
	w.Text = string(data[8 : 8+w.Length])
	return nil
}

// ---- Raw Encoding Wire (§7.7.1) ----
// Raw encoding is raw pixel data. Each pixel is serialized in color map
// format (big-endian bytes per pixel, matching the pixel format).
// The wire format is: [color_bytes][color_bytes]... for each pixel.

type RawEncodingWire struct {
	Bytes []byte
}

func (w *RawEncodingWire) WireMarshal() ([]byte, error) {
	return w.Bytes, nil
}

func (w *RawEncodingWire) WireUnmarshal(data []byte) error {
	w.Bytes = data
	return nil
}

// ---- DesktopSize Pseudo-Encoding Wire (§7.8.2) ----
// Wire: width(2 BE) + height(2 BE)

type DesktopSizeWire struct {
	Width  uint16
	Height uint16
}

func (w *DesktopSizeWire) WireMarshal() ([]byte, error) {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, w.Width)
	binary.Write(buf, binary.BigEndian, w.Height)
	return buf.Bytes(), nil
}

func (w *DesktopSizeWire) WireUnmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("DesktopSize: expected 4 bytes, got %d", len(data))
	}
	w.Width = binary.BigEndian.Uint16(data[0:2])
	w.Height = binary.BigEndian.Uint16(data[2:4])
	return nil
}

// ---- CopyRect Pseudo-Encoding Wire (§7.7.2) ----
// Wire: src_x(2 BE) + src_y(2 BE) + dest_x(2 BE) + dest_y(2 BE)

type CopyRectWire struct {
	SrcX, SrcY uint16
	DestX, DestY uint16
}

func (w *CopyRectWire) WireMarshal() ([]byte, error) {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, w.SrcX)
	binary.Write(buf, binary.BigEndian, w.SrcY)
	binary.Write(buf, binary.BigEndian, w.DestX)
	binary.Write(buf, binary.BigEndian, w.DestY)
	return buf.Bytes(), nil
}

func (w *CopyRectWire) WireUnmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("CopyRect: expected 8 bytes, got %d", len(data))
	}
	w.SrcX = binary.BigEndian.Uint16(data[0:2])
	w.SrcY = binary.BigEndian.Uint16(data[2:4])
	w.DestX = binary.BigEndian.Uint16(data[4:6])
	w.DestY = binary.BigEndian.Uint16(data[6:8])
	return nil
}
