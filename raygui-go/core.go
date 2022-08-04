package rayguigo

// #include "raylib.h"
// #include "raygui.h"
// #include <stdlib.h>
import "C"
import (
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GuiControlState int

const (
	GUI_STATE_NORMAL GuiControlState = iota
	GUI_STATE_FOCUSED
	GUI_STATE_PRESSED
	GUI_STATE_DISABLED
)

func GuiSetState(state GuiControlState) {
	C.GuiSetState(C.int(state))
}

func GuiSetFont(font rl.Font) {
	C.GuiSetFont(*((*C.Font)(unsafe.Pointer(&font))))
}
