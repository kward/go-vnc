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
