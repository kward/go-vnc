package vnc

import (
	"fmt"
	"math"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/kward/go-vnc/buttons"
	"github.com/kward/go-vnc/encodings"
	"github.com/kward/go-vnc/go/operators"
	"github.com/kward/go-vnc/keys"
	"github.com/kward/go-vnc/messages"
	"github.com/kward/go-vnc/rfbflags"
	"context"
)

func TestSetPixelFormat(t *testing.T) {
	tests := []struct {
		pf  PixelFormat
		msg SetPixelFormatMessage
	}{
		{
			PixelFormat{},
			SetPixelFormatMessage{
				Msg: messages.SetPixelFormat,
			}},
		{
			NewPixelFormat(16),
			SetPixelFormatMessage{
				Msg: messages.SetPixelFormat,
				PF: PixelFormat{
					BPP:        16,
					Depth:      16,
					BigEndian:  rfbflags.RFBTrue,
					TrueColor:  rfbflags.RFBTrue,
					RedMax:     uint16(math.Exp2(16)) - 1,
					GreenMax:   uint16(math.Exp2(16)) - 1,
					BlueMax:    uint16(math.Exp2(16)) - 1,
					RedShift:   0,
					GreenShift: 4,
					BlueShift:  8,
				},
			}},
	}

	mockConn := &MockConn{}
	conn := NewClientConn(mockConn, &ClientConfig{})

	for _, tt := range tests {
		mockConn.Reset()

		// Send request.
		if err := conn.SetPixelFormat(tt.pf); err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}

		// Read back in.
		req := SetPixelFormatMessage{}
		if err := conn.receive(&req); err != nil {
			t.Error(err)
			continue
		}

		// Validate the request.
		if got, want := req.Msg, messages.SetPixelFormat; got != want {
			t.Errorf("incorrect message-type; got = %v, want = %v", got, want)
		}
		if got, want := req.PF.BPP, tt.msg.PF.BPP; got != want {
			t.Errorf("incorrect pixel-format bits-per-pixel; got = %v, want = %v", got, want)
		}
		if got, want := req.PF.BlueShift, tt.msg.PF.BlueShift; got != want {
			t.Errorf("incorrect pixel-format blue-shift; got = %v, want = %v", got, want)
		}
	}
}

func TestSetEncodings(t *testing.T) {
	tests := []struct {
		encs     Encodings
		encTypes []encodings.Encoding
	}{
		{Encodings{&RawEncoding{}}, []encodings.Encoding{0}},
	}

	mockConn := &MockConn{}
	conn := NewClientConn(mockConn, &ClientConfig{})

	for _, tt := range tests {
		mockConn.Reset()

		// Send request.
		if err := conn.SetEncodings(tt.encs); err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}

		// Read back in.
		req := SetEncodingsMessage{}
		if err := conn.receive(&req); err != nil {
			t.Error(err)
			continue
		}
		var encs []int32 // Can't use the request struct.
		if err := conn.receiveN(&encs, int(req.NumEncs)); err != nil {
			t.Error(err)
			continue
		}

		// Validate the request.
		if got, want := req.Msg, messages.SetEncodings; got != want {
			t.Errorf("incorrect message-type; got = %v, want = %v", got, want)
			continue
		}
		if got, want := req.NumEncs, uint16(len(tt.encs)); got != want {
			t.Errorf("incorrect number-of-encodings; got = %v, want = %v", got, want)
			continue
		}
		if got, want := len(encs), len(tt.encs); got != want {
			t.Errorf("lengths of encodings don't match; got = %v, want = %v", got, want)
			continue
		}
		for i := 0; i < len(tt.encs); i++ {
			if got, want := encodings.Encoding(encs[i]), tt.encs[i].Type(); got != want {
				t.Errorf("incorrect encoding-type [%v]; got = %v, want = %v", i, got, want)
			}
		}
	}
}

func TestFramebufferUpdateRequest(t *testing.T) {
	tests := []struct {
		inc        rfbflags.RFBFlag
		x, y, w, h uint16
	}{
		{rfbflags.RFBFalse, 10, 20, 30, 40},
		{rfbflags.RFBTrue, 11, 21, 31, 41},
	}

	mockConn := &MockConn{}
	conn := NewClientConn(mockConn, &ClientConfig{})

	for _, tt := range tests {
		mockConn.Reset()

		// Send request.
		err := conn.FramebufferUpdateRequest(tt.inc, tt.x, tt.y, tt.w, tt.h)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}

		// Read back in.
		req := FramebufferUpdateRequestMessage{}
		if err := conn.receive(&req); err != nil {
			t.Error(err)
			continue
		}

		// Validate the request.
		if req.Msg != messages.FramebufferUpdateRequest {
			t.Errorf("incorrect message-type; got = %v, want = %v", req.Msg, messages.FramebufferUpdateRequest)
		}
		if req.Inc != tt.inc {
			t.Errorf("incorrect incremental; got = %v, want = %v", req.Inc, tt.inc)
		}
		if req.X != tt.x {
			t.Errorf("incorrect x-position; got = %v, want = %v", req.X, tt.x)
		}
		if req.Y != tt.y {
			t.Errorf("incorrect y-position; got = %v, want = %v", req.Y, tt.y)
		}
		if req.Width != tt.w {
			t.Errorf("incorrect width; got = %v, want = %v", req.Width, tt.w)
		}
		if req.Height != tt.h {
			t.Errorf("incorrect height; got = %v, want = %v", req.Height, tt.h)
		}
	}
}

func ExampleClientConn_KeyEvent() {
	// Establish TCP connection.
	nc, err := net.DialTimeout("tcp", "127.0.0.1:5900", 10*time.Second)
	if err != nil {
		panic(fmt.Sprintf("Error connecting to host: %v\n", err))
	}

	// Negotiate VNC connection.
	vc, err := Connect(context.Background(), nc, NewClientConfig("somepass"))
	if err != nil {
		panic(fmt.Sprintf("Could not negotiate a VNC connection: %v\n", err))
	}

	// Press and release the return key.
	vc.KeyEvent(keys.Return, true)
	vc.KeyEvent(keys.Return, false)

	// Close VNC connection.
	vc.Close()
}

func TestKeyEvent(t *testing.T) {
	tests := []struct {
		key  keys.Key
		down bool
	}{
		{keys.Digit0, PressKey},
		{keys.Digit1, ReleaseKey},
	}

	mockConn := &MockConn{}
	conn := NewClientConn(mockConn, &ClientConfig{})

	SetSettle(0) // Disable UI settling for tests.
	for _, tt := range tests {
		mockConn.Reset()

		// Send request.
		err := conn.KeyEvent(tt.key, tt.down)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}

		// Read back in.
		var req KeyEventMessage
		if err := conn.receive(&req); err != nil {
			t.Error(err)
			continue
		}

		// Validate the request.
		if got, want := req.Msg, messages.KeyEvent; got != want {
			t.Errorf("incorrect message-type; got = %v, want = %v", got, want)
		}
		down := rfbflags.ToBool(req.DownFlag)
		if got, want := down, tt.down; got != want {
			t.Errorf("incorrect down-flag; got = %v, want = %v", got, want)
		}
		if got, want := req.Key, tt.key; got != want {
			t.Errorf("incorrect key; got = %v, want = %v", got, want)
		}
	}
}

func ExampleClientConn_PointerEvent() {
	// Establish TCP connection.
	nc, err := net.DialTimeout("tcp", "127.0.0.1:5900", 10*time.Second)
	if err != nil {
		panic(fmt.Sprintf("Error connecting to host: %v\n", err))
	}

	// Negotiate VNC connection.
	vc, err := Connect(context.Background(), nc, NewClientConfig("somepass"))
	if err != nil {
		panic(fmt.Sprintf("Could not negotiate a VNC connection: %v\n", err))
	}

	// Move mouse to x=100, y=200.
	x, y := uint16(100), uint16(200)
	vc.PointerEvent(buttons.None, x, y)

	// Click and release the left mouse button.
	vc.PointerEvent(buttons.Left, x, y)
	vc.PointerEvent(buttons.None, x, y)

	// Close connection.
	vc.Close()
}

func TestPointerEvent(t *testing.T) {
	tests := []struct {
		button buttons.Button
		x, y   uint16
	}{
		{buttons.None, 0, 0},
		{buttons.Left | buttons.Right, 123, 456},
	}

	mockConn := &MockConn{}
	conn := NewClientConn(mockConn, &ClientConfig{})

	SetSettle(0) // Disable UI settling for tests.
	for _, tt := range tests {
		mockConn.Reset()

		// Send request.
		err := conn.PointerEvent(tt.button, tt.x, tt.y)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}

		// Read back in.
		req := PointerEventMessage{}
		if err := conn.receive(&req); err != nil {
			t.Error(err)
			continue
		}

		// Validate the request.
		if got, want := req.Msg, messages.PointerEvent; got != want {
			t.Errorf("incorrect message-type; got = %v, want = %v", got, want)
		}
		if got, want := req.Mask, buttons.Mask(tt.button); got != want {
			t.Errorf("incorrect button-mask; got = %v, want = %v", got, want)
		}
		if got, want := req.X, tt.x; got != want {
			t.Errorf("incorrect x-position; got = %v, want = %v", got, want)
		}
		if got, want := req.Y, tt.y; got != want {
			t.Errorf("incorrect y-position; got = %v, want = %v", got, want)
		}
	}
}

func TestClientCutText(t *testing.T) {
	tests := []struct {
		text string
		sent []byte
		ok   bool
	}{
		{"abc123", []byte("abc123"), true},
		{"foo\r\nbar", []byte("foo\nbar"), true},
		{"", []byte{}, true},
		{"ɹɐqooɟ", []byte{}, false},
	}

	mockConn := &MockConn{}
	conn := NewClientConn(mockConn, &ClientConfig{})

	SetSettle(0) // Disable UI settling for tests.
	for _, tt := range tests {
		mockConn.Reset()

		// Send request.
		err := conn.ClientCutText(tt.text)
		if err == nil && !tt.ok {
			t.Errorf("expected error")
		}
		if err != nil {
			if verr, ok := err.(*VNCError); !ok {
				t.Errorf("unexpected %v error: %v", reflect.TypeOf(err), verr)
			}
		}
		if !tt.ok {
			continue
		}

		// Read back in.
		req := ClientCutTextMessage{}
		if err := conn.receive(&req); err != nil {
			t.Error(err)
			continue
		}
		var text []byte
		if err := conn.receiveN(&text, int(req.Length)); err != nil {
			t.Error(err)
			continue
		}

		// Validate the request.
		if got, want := req.Msg, messages.ClientCutText; got != want {
			t.Errorf("incorrect message-type; got = %v, want = %v", got, want)
		}
		if got, want := req.Length, uint32(len(tt.sent)); got != want {
			t.Errorf("incorrect length; got = %v, want = %v", got, want)
		}
		if got, want := text, tt.sent; !operators.EqualSlicesOfByte(got, want) {
			t.Errorf("incorrect text; got = %v, want = %v", got, want)
		}
	}
}
