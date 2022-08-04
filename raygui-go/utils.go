package rayguigo

import (
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// #include "raylib.h"
// #include "raygui.h"
import "C"

func cRectangle(rect rl.Rectangle) C.Rectangle {
	return *((*C.Rectangle)(unsafe.Pointer(&rect)))
}

func goRectangle(rect C.Rectangle) rl.Rectangle {
	return *((*rl.Rectangle)(unsafe.Pointer(&rect)))
}
