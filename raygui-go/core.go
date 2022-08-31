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
type GuiControl int
type GuiControlProperty int

const (
	GUI_STATE_NORMAL GuiControlState = iota
	GUI_STATE_FOCUSED
	GUI_STATE_PRESSED
	GUI_STATE_DISABLED
)

const (
	// Default -> populates to all controls when set
	DEFAULT GuiControl = iota
	// Basic controls
	LABEL // Used also for: LABELBUTTON
	BUTTON
	TOGGLE // Used also for: TOGGLEGROUP
	SLIDER // Used also for: SLIDERBAR
	PROGRESSBAR
	CHECKBOX
	COMBOBOX
	DROPDOWNBOX
	TEXTBOX // Used also for: TEXTBOXMULTI
	VALUEBOX
	SPINNER // Uses: BUTTON, VALUEBOX
	LISTVIEW
	COLORPICKER
	SCROLLBAR
	STATUSBAR
)

const (
	BORDER_COLOR_NORMAL GuiControlProperty = iota
	BASE_COLOR_NORMAL
	TEXT_COLOR_NORMAL
	BORDER_COLOR_FOCUSED
	BASE_COLOR_FOCUSED
	TEXT_COLOR_FOCUSED
	BORDER_COLOR_PRESSED
	BASE_COLOR_PRESSED
	TEXT_COLOR_PRESSED
	BORDER_COLOR_DISABLED
	BASE_COLOR_DISABLED
	TEXT_COLOR_DISABLED
	BORDER_WIDTH
	TEXT_PADDING
	TEXT_ALIGNMENT
	RESERVED
)

const (
	TEXT_ALIGN_LEFT = iota
	TEXT_ALIGN_CENTER
	TEXT_ALIGN_RIGHT
)

func GuiSetState(state GuiControlState) {
	C.GuiSetState(C.int(state))
}

func GuiSetFont(font rl.Font) {
	C.GuiSetFont(*((*C.Font)(unsafe.Pointer(&font))))
}

func GuiSetStyle(control GuiControl, property GuiControlProperty, value int) {
	C.GuiSetStyle(C.int(control), C.int(property), C.int(value))
}

func GuiLock() {
	C.GuiLock()
}

func GuiUnlock() {
	C.GuiUnlock()
}
