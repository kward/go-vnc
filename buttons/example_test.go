package buttons_test

import (
	"fmt"

	"github.com/kward/go-vnc/buttons"
)

// ExampleButton demonstrates using individual mouse button constants.
func ExampleButton() {
	// Individual buttons
	fmt.Printf("Left button: %s (0x%x)\n", buttons.Left, uint8(buttons.Left))
	fmt.Printf("Middle button: %s (0x%x)\n", buttons.Middle, uint8(buttons.Middle))
	fmt.Printf("Right button: %s (0x%x)\n", buttons.Right, uint8(buttons.Right))

	// Additional buttons
	fmt.Printf("Button 4: %s (0x%x)\n", buttons.Four, uint8(buttons.Four))
	fmt.Printf("Button 5: %s (0x%x)\n", buttons.Five, uint8(buttons.Five))

	// Output:
	// Left button: Left (0x1)
	// Middle button: Middle (0x2)
	// Right button: Right (0x4)
	// Button 4: Four (0x8)
	// Button 5: Five (0x10)
}

// Example_buttonMask demonstrates combining multiple buttons into a button mask.
// VNC PointerEvent messages use a button mask to indicate which buttons are pressed.
func Example_buttonMask() {
	// Single button press
	mask := uint8(buttons.Left)
	fmt.Printf("Left click: mask=0x%x\n", mask)

	// Multiple buttons pressed simultaneously (e.g., left+right)
	mask = uint8(buttons.Left | buttons.Right)
	fmt.Printf("Left+Right: mask=0x%x\n", mask)

	// All mouse buttons pressed
	mask = uint8(buttons.Left | buttons.Middle | buttons.Right)
	fmt.Printf("All buttons: mask=0x%x\n", mask)

	// Output:
	// Left click: mask=0x1
	// Left+Right: mask=0x5
	// All buttons: mask=0x7
}

// Example_checkButton demonstrates checking if a specific button is pressed
// in a button mask.
func Example_checkButton() {
	// Simulate a mask with left and middle buttons pressed
	mask := uint8(buttons.Left | buttons.Middle)

	// Check individual buttons
	leftPressed := (mask & uint8(buttons.Left)) != 0
	middlePressed := (mask & uint8(buttons.Middle)) != 0
	rightPressed := (mask & uint8(buttons.Right)) != 0

	fmt.Printf("Button mask: 0x%x\n", mask)
	fmt.Printf("Left pressed: %v\n", leftPressed)
	fmt.Printf("Middle pressed: %v\n", middlePressed)
	fmt.Printf("Right pressed: %v\n", rightPressed)

	// Output:
	// Button mask: 0x3
	// Left pressed: true
	// Middle pressed: true
	// Right pressed: false
}
