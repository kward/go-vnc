//lint:file-ignore S1021 stylistic suggestions not necessary for tests
package vnc

import (
	"testing"

	"github.com/kward/go-vnc/encodings"
	"github.com/kward/go-vnc/go/operators"
)

func TestRectangle_Marshal(t *testing.T) {
	var (
		bytes []byte
		rect  *Rectangle
	)

	mockConn := &MockConn{}
	conn := NewClientConn(mockConn, &ClientConfig{})

	rect = &Rectangle{1, 2, 3, 4, &RawEncoding{}, conn.Encodable}
	bytes, err := rect.Marshal()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got, want := bytes, []byte{0, 1, 0, 2, 0, 3, 0, 4, 0, 0, 0, 0}; !operators.EqualSlicesOfByte(got, want) {
		t.Errorf("incorrect result; got = %v, want = %v", got, want)
	}
}

func TestRectangle_Unmarshal(t *testing.T) {
	var rect *Rectangle

	rect = &Rectangle{}
	if err := rect.Unmarshal([]byte{0, 2, 0, 3, 0, 4, 0, 5, 0, 0, 0, 0}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got, want := rect.X, uint16(2); got != want {
		t.Errorf("incorrect x-position; got = %v, want = %v", got, want)
	}
	if got, want := rect.Y, uint16(3); got != want {
		t.Errorf("incorrect y-position; got = %v, want = %v", got, want)
	}
	if got, want := rect.Width, uint16(4); got != want {
		t.Errorf("incorrect width; got = %v, want = %v", got, want)
	}
	if got, want := rect.Height, uint16(5); got != want {
		t.Errorf("incorrect height; got = %v, want = %v", got, want)
	}
	if rect.Enc == nil {
		t.Errorf("empty encoding")
		return
	}
	if got, want := rect.Enc.Type(), encodings.Raw; got != want {
		t.Errorf("incorrect encoding-type; got = %v, want = %v", got, want)
	}
}

// TODO(kward): need to read encodings in addition to rectangles.
func TestFramebufferUpdate(t *testing.T) {
	mockConn := &MockConn{}
	conn := NewClientConn(mockConn, &ClientConfig{})
	// Use empty PixelFormat so that the BPP is zero, and rects won't be read.
	// TODO(kward): give some real rectangles so this hack isn't necessary.
	conn.pixelFormat = PixelFormat{}

	for _, tt := range []struct {
		desc  string
		rects []Rectangle
		ok    bool
	}{
		{"single raw encoded rect",
			[]Rectangle{{1, 2, 3, 4, &RawEncoding{}, conn.Encodable}}, true},
	} {
		mockConn.Reset()

		// Send the message.
		msg := newFramebufferUpdate(tt.rects)
		bytes, err := msg.Marshal()
		if err != nil {
			t.Errorf("%s: failed to marshal; %s", tt.desc, err)
			continue
		}
		if err := conn.send(bytes); err != nil {
			t.Errorf("%s: failed to send; %s", tt.desc, err)
			continue
		}

		// Validate message handling.
		var messageType uint8
		if err := conn.receive(&messageType); err != nil {
			t.Fatal(err)
		}
		fu := &FramebufferUpdate{}
		parsedFU, err := fu.Read(conn)
		if err != nil {
			t.Fatalf("%s: failed to read; %s", tt.desc, err)
		}
		rects := parsedFU.(*FramebufferUpdate).Rects

		// Validate the message.
		if got, want := len(rects), len(tt.rects); got != want {
			t.Errorf("%s: incorrect number-of-rectangles; got %d, want %d", tt.desc, got, want)
			continue
		}
		for i, r := range tt.rects {
			if got, want := rects[i].X, r.X; got != want {
				t.Errorf("%s: rect[%d] incorrect x-position; got %d, want %d", tt.desc, i, got, want)
			}
			if got, want := rects[i].Y, r.Y; got != want {
				t.Errorf("%s: rect[%d] incorrect y-position; got %d, want %d", tt.desc, i, got, want)
			}
			if got, want := rects[i].Width, r.Width; got != want {
				t.Errorf("%s: rect[%d] incorrect width; got %d, want %d", tt.desc, i, got, want)
			}
			if got, want := rects[i].Height, r.Height; got != want {
				t.Errorf("%s: rect[%d] incorrect height; got %d, want %d", tt.desc, i, got, want)
			}
			if rects[i].Enc == nil {
				t.Errorf("%s: rect[%d] has empty encoding", tt.desc, i)
				continue
			}
			if got, want := rects[i].Enc.Type(), r.Enc.Type(); got != want {
				t.Errorf("%s: rect[%d] incorrect encoding-type; got %d, want %d", tt.desc, i, got, want)
			}
		}
	}
}

func TestColor_Marshal(t *testing.T) {
	cm := ColorMap{}
	for i := 0; i < len(cm); i++ {
		cm[i] = Color{R: uint16(i), G: uint16(i << 4), B: uint16(i << 8)}
	}

	tests := []struct {
		c    *Color
		data []byte
	}{
		// 8 BPP, with ColorMap
		{&Color{&PixelFormat8bit, &cm, 0, 0, 0, 0}, []byte{0}},
		{&Color{&PixelFormat8bit, &cm, 127, 127, 2032, 32512}, []byte{127}},
		{&Color{&PixelFormat8bit, &cm, 255, 255, 4080, 65280}, []byte{255}},
		// 16 BPP
		{&Color{&PixelFormat16bit, &ColorMap{}, 0, 0, 0, 0}, []byte{0, 0}},
		{&Color{&PixelFormat16bit, &ColorMap{}, 0, 127, 7, 0}, []byte{0, 127}},
		{&Color{&PixelFormat16bit, &ColorMap{}, 0, 32767, 2047, 127}, []byte{127, 255}},
		{&Color{&PixelFormat16bit, &ColorMap{}, 0, 65535, 4095, 255}, []byte{255, 255}},
		// 32 BPP
		{&Color{&PixelFormat32bit, &ColorMap{}, 0, 0, 0, 0}, []byte{0, 0, 0, 0}},
		{&Color{&PixelFormat32bit, &ColorMap{}, 0, 127, 0, 0}, []byte{0, 0, 0, 127}},
		{&Color{&PixelFormat32bit, &ColorMap{}, 0, 32767, 127, 0}, []byte{0, 0, 127, 255}},
		{&Color{&PixelFormat32bit, &ColorMap{}, 0, 65535, 32767, 127}, []byte{0, 127, 255, 255}},
		{&Color{&PixelFormat32bit, &ColorMap{}, 0, 65535, 65535, 32767}, []byte{127, 255, 255, 255}},
		{&Color{&PixelFormat32bit, &ColorMap{}, 0, 65535, 65535, 65535}, []byte{255, 255, 255, 255}},
	}

	for i, tt := range tests {
		data, err := tt.c.Marshal()
		if err != nil {
			t.Errorf("%v: unexpected error: %v", i, err)
			continue
		}
		if got, want := data, tt.data; !operators.EqualSlicesOfByte(got, want) {
			t.Errorf("%v: incorrect result; got = %v, want = %v", i, got, want)
		}
	}
}

func TestColor_Unmarshal(t *testing.T) {
	var cm ColorMap
	for i := 0; i < len(cm); i++ {
		cm[i] = Color{R: uint16(i), G: uint16(i << 4), B: uint16(i << 8)}
	}

	tests := []struct {
		data    []byte
		pf      *PixelFormat
		cm      *ColorMap
		cmIndex uint32
		R, G, B uint16
	}{
		// 8 BPP, with ColorMap
		{[]byte{0}, &PixelFormat8bit, &cm, 0, 0, 0, 0},
		{[]byte{127}, &PixelFormat8bit, &cm, 127, 127, 2032, 32512},
		{[]byte{255}, &PixelFormat8bit, &cm, 255, 255, 4080, 65280},
		// 16 BPP
		{[]byte{0, 0}, &PixelFormat16bit, &ColorMap{}, 0, 0, 0, 0},
		{[]byte{0, 127}, &PixelFormat16bit, &ColorMap{}, 0, 127, 7, 0},
		{[]byte{127, 255}, &PixelFormat16bit, &ColorMap{}, 0, 32767, 2047, 127},
		{[]byte{255, 255}, &PixelFormat16bit, &ColorMap{}, 0, 65535, 4095, 255},
		// 32 BPP
		{[]byte{0, 0, 0, 0}, &PixelFormat32bit, &ColorMap{}, 0, 0, 0, 0},
		{[]byte{0, 0, 0, 127}, &PixelFormat32bit, &ColorMap{}, 0, 127, 0, 0},
		{[]byte{0, 0, 127, 255}, &PixelFormat32bit, &ColorMap{}, 0, 32767, 127, 0},
		{[]byte{0, 127, 255, 255}, &PixelFormat32bit, &ColorMap{}, 0, 65535, 32767, 127},
		{[]byte{127, 255, 255, 255}, &PixelFormat32bit, &ColorMap{}, 0, 65535, 65535, 32767},
		{[]byte{255, 255, 255, 255}, &PixelFormat32bit, &ColorMap{}, 0, 65535, 65535, 65535},
	}

	for i, tt := range tests {
		color := NewColor(tt.pf, tt.cm)
		if err := color.Unmarshal(tt.data); err != nil {
			t.Errorf("%v: unexpected error: %v", i, err)
			continue
		}
		if got, want := color.cmIndex, tt.cmIndex; got != want {
			t.Errorf("%v: incorrect cmIndex value; got = %v, want = %v", i, got, want)
		}
		if got, want := color.R, tt.R; got != want {
			t.Errorf("%v: incorrect R value; got = %v, want = %v", i, got, want)
		}
		if got, want := color.G, tt.G; got != want {
			t.Errorf("%v: incorrect G value; got = %v, want = %v", i, got, want)
		}
		if got, want := color.B, tt.B; got != want {
			t.Errorf("%v: incorrect B value; got = %v, want = %v", i, got, want)
		}
	}
}

func TestSetColorMapEntries(t *testing.T) {}

func TestBell(t *testing.T) {}

func TestServerCutText(t *testing.T) {}
