package encoding

import (
	"encoding/binary"
	"testing"

	"github.com/kward/go-vnc/encodings"
)

func TestPixelFormatWire_marshal_unmarshal(t *testing.T) {
	original := PixelFormatWire{
		BPP:        32,
		Depth:      24,
		BigEndian:  0,
		TrueColor:  1,
		RedMax:     0xFF,
		GreenMax:   0xFF,
		BlueMax:    0xFF,
		RedShift:   16,
		GreenShift: 8,
		BlueShift:  0,
		Padding:    [3]byte{},
	}
	data, err := original.WireMarshal()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 16 {
		t.Fatalf("expected 16 bytes, got %d", len(data))
	}
	var got PixelFormatWire
	if err := got.WireUnmarshal(data); err != nil {
		t.Fatal(err)
	}
	if got.BPP != original.BPP || got.RedShift != original.RedShift {
		t.Errorf("mismatch: got %+v, want %+v", got, original)
	}
}

func TestSetPixelFormatRequestWire_roundtrip(t *testing.T) {
	original := SetPixelFormatRequestWire{
		MsgType: 0,
		PF: PixelFormatWire{
			BPP: 32, Depth: 24, BigEndian: 0, TrueColor: 1,
			RedMax: 0xFF, GreenMax: 0xFF, BlueMax: 0xFF,
			RedShift: 16, GreenShift: 8, BlueShift: 0,
			Padding: [3]byte{},
		},
	}
	data, err := original.WireMarshal()
	if err != nil {
		t.Fatal(err)
	}
	var got SetPixelFormatRequestWire
	if err := got.WireUnmarshal(data); err != nil {
		t.Fatal(err)
	}
	if got.MsgType != original.MsgType {
		t.Errorf("MsgType: got %d, want %d", got.MsgType, original.MsgType)
	}
	if got.PF.BPP != original.PF.BPP {
		t.Errorf("BPP: got %d, want %d", got.PF.BPP, original.PF.BPP)
	}
}

func TestSetEncodingsRequestWire_roundtrip(t *testing.T) {
	original := SetEncodingsRequestWire{
		MsgType:   2,
		NumEnc:    2,
		Encodings: []encodings.Encoding{0, 1}, // [Raw, CopyRect]
	}
	data, err := original.WireMarshal()
	if err != nil {
		t.Fatal(err)
	}
	var got SetEncodingsRequestWire
	if err := got.WireUnmarshal(data); err != nil {
		t.Fatal(err)
	}
	if got.MsgType != 2 {
		t.Errorf("MsgType: got %d", got.MsgType)
	}
	if got.NumEnc != 2 {
		t.Errorf("NumEnc: got %d", got.NumEnc)
	}
}

func TestFramebufferUpdateRequestWire_roundtrip(t *testing.T) {
	original := FramebufferUpdateRequestWire{
		MsgType:     3,
		Incremental: 1,
		X:           0, Y: 0, Width: 800, Height: 600,
	}
	data, err := original.WireMarshal()
	if err != nil {
		t.Fatal(err)
	}
	var got FramebufferUpdateRequestWire
	if err := got.WireUnmarshal(data); err != nil {
		t.Fatal(err)
	}
	if got.MsgType != 3 || got.Width != 800 || got.Height != 600 {
		t.Errorf("mismatch: got %+v", got)
	}
}

func TestKeyEventWire_roundtrip(t *testing.T) {
	original := KeyEventWire{
		MsgType:  4,
		DownFlag: 1,
		Key:      65,
	}
	data, err := original.WireMarshal()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 8 {
		t.Fatalf("expected 8 bytes, got %d", len(data))
	}
	var got KeyEventWire
	if err := got.WireUnmarshal(data); err != nil {
		t.Fatal(err)
	}
	if got.Key != 65 || got.DownFlag != 1 {
		t.Errorf("mismatch: got %+v", got)
	}
}

func TestPointerEventWire_roundtrip(t *testing.T) {
	original := PointerEventWire{
		MsgType:    5,
		ButtonMask: 1,
		X:          100, Y: 200,
	}
	data, err := original.WireMarshal()
	if err != nil {
		t.Fatal(err)
	}
	var got PointerEventWire
	if err := got.WireUnmarshal(data); err != nil {
		t.Fatal(err)
	}
	if got.X != 100 || got.Y != 200 || got.ButtonMask != 1 {
		t.Errorf("mismatch: got %+v", got)
	}
}

func TestClientCutTextWire_roundtrip(t *testing.T) {
	original := ClientCutTextWire{
		MsgType: 6,
		Text:    "Hello",
	}
	data, err := original.WireMarshal()
	if err != nil {
		t.Fatal(err)
	}
	// The length field (big-endian uint32 at bytes 4-8) should equal the text length.
	length := binary.BigEndian.Uint32(data[4:8])
	if length != 5 {
		t.Errorf("unexpected length: got %d, want 5", length)
	}
	var got ClientCutTextWire
	if err := got.WireUnmarshal(data); err != nil {
		t.Fatal(err)
	}
	if got.Text != "Hello" {
		t.Errorf("Text: got %q, want %q", got.Text, "Hello")
	}
}

func TestRectDataWire_roundtrip(t *testing.T) {
	original := RectDataWire{
		X: 0, Y: 10, Width: 20, Height: 30,
		Encoding: encodings.Encoding(0),
		Payload:  []byte{1, 2, 3},
	}
	data, err := original.WireMarshal()
	if err != nil {
		t.Fatal(err)
	}
	var got RectDataWire
	if err := got.WireUnmarshal(data); err != nil {
		t.Fatal(err)
	}
	if got.X != 0 || got.Y != 10 || got.Width != 20 || got.Height != 30 {
		t.Errorf("mismatch: got %+v", got)
	}
}

func TestBellWire_marshal(t *testing.T) {
	bell := BellWire{}
	data, err := bell.WireMarshal()
	if err != nil {
		t.Fatal(err)
	}
	if data[0] != 2 { // Bell = message type 2
		t.Errorf("expected message type 2, got %d", data[0])
	}
}
