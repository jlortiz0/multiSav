package rayguigo

import (
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// #include "raylib.h"
// #include "raygui.h"
// #include <stdlib.h>
import "C"

func GuiWindowBox(bounds rl.Rectangle, title string) bool {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	return bool(C.GuiWindowBox(cRectangle(bounds), cTitle))
}

func GuiGroupBox(bounds rl.Rectangle, title string) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.GuiGroupBox(cRectangle(bounds), cTitle)
}

func GuiLine(bounds rl.Rectangle, title string) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.GuiLine(cRectangle(bounds), cTitle)
}

func GuiPanel(bounds rl.Rectangle) {
	C.GuiPanel(cRectangle(bounds))
}

func GuiScrollPanel(bounds, content rl.Rectangle, scroll *rl.Vector2) rl.Rectangle {
	return goRectangle(C.GuiScrollPanel(cRectangle(bounds), cRectangle(content), (*C.Vector2)(unsafe.Pointer(scroll))))
}
