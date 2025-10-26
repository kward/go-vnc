package keys_test

import (
	"fmt"

	"github.com/kward/go-vnc/keys"
)

// ExampleFromRune demonstrates converting individual runes to Key values.
// FromRune handles printable ASCII, extended Latin-1, and common control characters.
func ExampleFromRune() {
	// Printable ASCII characters
	k, ok := keys.FromRune('A')
	if ok {
		fmt.Printf("'A' -> %s (0x%x)\n", k, uint32(k))
	}

	// Control character
	k, ok = keys.FromRune('\n')
	if ok {
		fmt.Printf("'\\n' -> %s (0x%x)\n", k, uint32(k))
	}

	// Unsupported character
	_, ok = keys.FromRune('😀')
	fmt.Printf("'😀' supported: %v\n", ok)

	// Output:
	// 'A' -> A (0x41)
	// '\n' -> Linefeed (0xff0a)
	// '😀' supported: false
}

// ExampleTextToKeys demonstrates converting a string to a slice of Key values.
// This is useful for simulating typing text via VNC.
func ExampleTextToKeys() {
	// Convert a simple string to keys
	ks, err := keys.TextToKeys("Hello")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("'Hello' -> %d keys\n", len(ks))
	fmt.Printf("First key: %s\n", ks[0])

	// String with control characters
	ks, err = keys.TextToKeys("line1\nline2")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("'line1\\nline2' -> %d keys (includes Linefeed)\n", len(ks))

	// Unsupported characters produce an error
	_, err = keys.TextToKeys("test😀")
	fmt.Printf("Error with emoji: %v\n", err != nil)

	// Output:
	// 'Hello' -> 5 keys
	// First key: H
	// 'line1\nline2' -> 11 keys (includes Linefeed)
	// Error with emoji: true
}

// ExampleIntToKeys demonstrates converting an integer to Key values.
// This is useful for typing numbers via VNC.
func ExampleIntToKeys() {
	// Positive number
	ks := keys.IntToKeys(123)
	fmt.Printf("123 -> %d keys:", len(ks))
	for _, k := range ks {
		fmt.Printf(" %s", k)
	}
	fmt.Println()

	// Negative number (includes minus sign)
	ks = keys.IntToKeys(-42)
	fmt.Printf("-42 -> %d keys:", len(ks))
	for _, k := range ks {
		fmt.Printf(" %s", k)
	}
	fmt.Println()

	// Zero
	ks = keys.IntToKeys(0)
	fmt.Printf("0 -> %d keys: %s\n", len(ks), ks[0])

	// Output:
	// 123 -> 3 keys: Digit1 Digit2 Digit3
	// -42 -> 3 keys: Minus Digit4 Digit2
	// 0 -> 1 keys: Digit0
}

// ExampleKey demonstrates direct Key constant usage for special keys.
func ExampleKey() {
	// Special keys are available as constants
	fmt.Printf("Enter key: %s (0x%x)\n", keys.Return, uint32(keys.Return))
	fmt.Printf("Escape key: %s (0x%x)\n", keys.Escape, uint32(keys.Escape))
	fmt.Printf("Tab key: %s (0x%x)\n", keys.Tab, uint32(keys.Tab))

	// Function keys
	fmt.Printf("F1 key: %s (0x%x)\n", keys.F1, uint32(keys.F1))

	// Modifier keys
	fmt.Printf("Left Shift: %s (0x%x)\n", keys.ShiftLeft, uint32(keys.ShiftLeft))
	fmt.Printf("Left Control: %s (0x%x)\n", keys.ControlLeft, uint32(keys.ControlLeft))

	// Output:
	// Enter key: Return (0xff0d)
	// Escape key: Escape (0xff1b)
	// Tab key: Tab (0xff09)
	// F1 key: F1 (0xffbe)
	// Left Shift: ShiftLeft (0xffe1)
	// Left Control: ControlLeft (0xffe3)
}
