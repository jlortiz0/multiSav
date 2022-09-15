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
	TEXT_SIZE        GuiControlProperty = iota + 16 // Text size (glyphs max height)
	TEXT_SPACING                                    // Text spacing between glyphs
	LINE_COLOR                                      // Line control color
	BACKGROUND_COLOR                                // Background color
)

const GROUP_PADDING = 16 // ToggleGroup separation between toggles

// Slider/SliderBar
const (
	SLIDER_WIDTH   = iota + 16 // Slider size of internal bar
	SLIDER_PADDING             // Slider/SliderBar internal bar padding
)

// ProgressBar
const PROGRESS_PADDING = 16 // ProgressBar internal padding

// ScrollBar
const (
	ARROWS_SIZE = iota + 16
	ARROWS_VISIBLE
	SCROLL_SLIDER_PADDING // (SLIDERBAR, SLIDER_PADDING)
	SCROLL_SLIDER_SIZE
	SCROLL_PADDING
	SCROLL_SPEED
)

// CheckBox
const CHECK_PADDING = 16 // CheckBox internal check padding

// ComboBox
const (
	COMBO_BUTTON_WIDTH   = iota + 16 // ComboBox right button width
	COMBO_BUTTON_SPACING             // ComboBox button separation
)

// DropdownBox
const (
	ARROW_PADDING          = iota + 16 // DropdownBox arrow separation from border and items
	DROPDOWN_ITEMS_SPACING             // DropdownBox items separation
)

// TextBox/TextBoxMulti/ValueBox/Spinner
const (
	TEXT_INNER_PADDING = iota + 16 // TextBox/TextBoxMulti/ValueBox/Spinner inner text padding
	TEXT_LINES_SPACING             // TextBoxMulti lines separation
)

// Spinner
const (
	SPIN_BUTTON_WIDTH   = iota + 16 // Spinner left/right buttons width
	SPIN_BUTTON_SPACING             // Spinner buttons separation
)

// ListView
const (
	LIST_ITEMS_HEIGHT  = iota + 16 // ListView items height
	LIST_ITEMS_SPACING             // ListView items separation
	SCROLLBAR_WIDTH                // ListView scrollbar size (usually width)
	SCROLLBAR_SIDE                 // ListView scrollbar side (0-left, 1-right)
)

// ColorPicker
const (
	COLOR_SELECTOR_SIZE      = iota + 16
	HUEBAR_WIDTH             // ColorPicker right hue bar width
	HUEBAR_PADDING           // ColorPicker right hue bar separation from panel
	HUEBAR_SELECTOR_HEIGHT   // ColorPicker right hue bar selector height
	HUEBAR_SELECTOR_OVERFLOW // ColorPicker right hue bar selector overflow
)

const SCROLLBAR_LEFT_SIDE = 0
const SCROLLBAR_RIGHT_SIDE = 1

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

func GuiGetStyle(control GuiControl, property GuiControlProperty) int {
	return int(C.GuiGetStyle(C.int(control), C.int(property)))
}
