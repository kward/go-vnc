package keys

import (
	"reflect"
	"testing"
)

func TestIntToKeys(t *testing.T) {
	for _, tt := range []struct {
		val  int
		keys Keys
	}{
		{-1234, Keys{Minus, Digit1, Digit2, Digit3, Digit4}},
		{0, Keys{Digit0}},
		{5678, Keys{Digit5, Digit6, Digit7, Digit8}},
	} {
		if got, want := IntToKeys(tt.val), tt.keys; !reflect.DeepEqual(got, want) {
			t.Errorf("IntToKeys(%d) = %v, want %v", tt.val, got, want)
			continue
		}
	}
}

func TestASCIIMappingToKeyConstants(t *testing.T) {
	// Digits '0'..'9'
	for r, k := range map[rune]Key{
		'0': Digit0, '1': Digit1, '2': Digit2, '3': Digit3, '4': Digit4,
		'5': Digit5, '6': Digit6, '7': Digit7, '8': Digit8, '9': Digit9,
		'-': Minus, ' ': Space,
		'A': A, 'Z': Z,
		'a': SmallA, 'z': SmallZ,
	} {
		if got := Key(r); got != k {
			t.Fatalf("Key(%q) = %v, want %v", r, got, k)
		}
	}
}

func TestControlKeyValues(t *testing.T) {
	// Anchor a few well-known control keys to expected codes.
	// These mirror X11 KeySym values used by RFB.
	tests := []struct {
		name string
		key  Key
		want Key
	}{
		{"BackSpace", BackSpace, 0xff08},
		{"Tab", Tab, 0xff09},
		{"Return", Return, 0xff0d},
		{"Escape", Escape, 0xff1b},
		{"Delete", Delete, 0xffff},
	}
	for _, tt := range tests {
		if tt.key != tt.want {
			t.Errorf("%s = 0x%x, want 0x%x", tt.name, uint32(tt.key), uint32(tt.want))
		}
	}
}

func TestFromRune(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want Key
		ok   bool
	}{
		// Printable ASCII
		{"space", ' ', Space, true},
		{"digit0", '0', Digit0, true},
		{"digit9", '9', Digit9, true},
		{"minus", '-', Minus, true},
		{"A", 'A', A, true},
		{"Z", 'Z', Z, true},
		{"a", 'a', SmallA, true},
		{"z", 'z', SmallZ, true},
		{"tilde", '~', AsciiTilde, true},
		// Control characters
		{"newline", '\n', Linefeed, true},
		{"tab", '\t', Tab, true},
		{"backspace", '\b', BackSpace, true},
		{"return", '\r', Return, true},
		// Extended Latin-1
		{"non-breaking space", '\u00A0', Key(0xA0), true},
		{"yen", '¥', Key(0xA5), true},
		// Unsupported
		{"emoji", '😀', 0, false},
		{"high unicode", '\u2013', 0, false},
		{"control below 0x20", '\x01', 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := FromRune(tt.r)
			if ok != tt.ok {
				t.Errorf("FromRune(%q) ok = %v, want %v", tt.r, ok, tt.ok)
			}
			if ok && got != tt.want {
				t.Errorf("FromRune(%q) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}

func TestTextToKeys(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    Keys
		wantErr bool
	}{
		{"empty", "", Keys{}, false},
		{"digits", "123", Keys{Digit1, Digit2, Digit3}, false},
		{"mixed", "A0-z", Keys{A, Digit0, Minus, SmallZ}, false},
		{"with newline", "hi\n", Keys{SmallH, SmallI, Linefeed}, false},
		{"with tab", "a\tb", Keys{SmallA, Tab, SmallB}, false},
		{"emoji fail", "test😀", nil, true},
		{"high unicode fail", "test—more", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TextToKeys(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("TextToKeys(%q) error = %v, wantErr %v", tt.text, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TextToKeys(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}
