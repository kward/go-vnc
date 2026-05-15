// Implementation of RFC 6143 §7.4 Pixel Format Data Structure.

package vnc

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/kward/go-vnc/encoding"
)

var (
	PixelFormat8bit  PixelFormat = NewPixelFormat(8)
	PixelFormat16bit PixelFormat = NewPixelFormat(16)
	PixelFormat32bit PixelFormat = NewPixelFormat(32)
)

// PixelFormat describes the way a pixel is formatted for a VNC connection.
type PixelFormat struct {
	BPP                             uint8  // bits-per-pixel
	Depth                           uint8  // depth
	BigEndian                       bool   // big-endian-flag
	TrueColor                       bool   // true-color-flag
	RedMax, GreenMax, BlueMax       uint16 // red-, green-, blue-max (2^BPP-1)
	RedShift, GreenShift, BlueShift uint8  // red-, green-, blue-shift
	_                               [3]byte // padding
}

const pixelFormatLen = 16

// Verify that interfaces are honored.
var _ fmt.Stringer = (*PixelFormat)(nil)
var _ MarshalerUnmarshaler = (*PixelFormat)(nil)

// NewPixelFormat returns a populated PixelFormat structure.
func NewPixelFormat(bpp uint8) PixelFormat {
	// Avoid float rounding/overflow; cap at 0xFFFF since fields are uint16.
	var rgbMax uint16
	if bpp >= 16 {
		rgbMax = 0xFFFF
	} else {
		rgbMax = uint16((1 << bpp) - 1)
	}
	var (
		bigEndian = true
		tc        = true
		rs, gs, bs uint8
	)
	switch bpp {
	case 8:
		tc = false
		rs, gs, bs = 0, 0, 0
	case 16:
		rs, gs, bs = 0, 4, 8
	case 32:
		rs, gs, bs = 0, 8, 16
	}
	return PixelFormat{bpp, bpp, bigEndian, tc, rgbMax, rgbMax, rgbMax, rs, gs, bs, [3]byte{}}
}

// Marshal implements the Marshaler interface.
func (pf PixelFormat) Marshal() ([]byte, error) {
	// Validation checks.
	switch pf.BPP {
	case 8, 16, 32:
	default:
		return nil, NewVNCError(fmt.Sprintf("Invalid BPP value %v; must be 8, 16, or 32", pf.BPP))
	}

	if pf.Depth < pf.BPP {
		return nil, NewVNCError(fmt.Sprintf("Invalid Depth value %v; cannot be < BPP", pf.Depth))
	}
	switch pf.Depth {
	case 8, 16, 32:
	default:
		return nil, NewVNCError(fmt.Sprintf("Invalid Depth value %v; must be 8, 16, or 32", pf.Depth))
	}

	// Use the Pixel_FormatWire type to serialize.
	wire := encoding.PixelFormatWire{
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
		Padding:    [3]byte{},
	}
	return wire.WireMarshal()
}

// Read reads from an io.Reader, and populates the PixelFormat.
func (pf *PixelFormat) Read(r io.Reader) error {
	buf := make([]byte, pixelFormatLen)
	if _, err := io.ReadAtLeast(r, buf, pixelFormatLen); err != nil {
		return err
	}
	return pf.Unmarshal(buf)
}

// Unmarshal implements the Unmarshaler interface.
func (pf *PixelFormat) Unmarshal(data []byte) error {
	buf := NewBuffer(data)

	var msg PixelFormat
	if err := buf.Read(&msg); err != nil {
		return err
	}
	*pf = msg

	return nil
}

// String implements the fmt.Stringer interface.
func (pf PixelFormat) String() string {
	return fmt.Sprintf("{ bpp: %d depth: %d big-endian: %t true-color: %t red-max: %d green-max: %d blue-max: %d red-shift: %d green-shift: %d blue-shift: %d }",
		pf.BPP, pf.Depth, pf.BigEndian, pf.TrueColor, pf.RedMax, pf.GreenMax, pf.BlueMax, pf.RedShift, pf.GreenShift, pf.BlueShift)
}

func (pf PixelFormat) order() binary.ByteOrder {
	if pf.BigEndian {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

// boolToUint8 converts bool to uint8 (1 or 0) for wire format.
func boolToUint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
