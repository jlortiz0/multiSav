package rayguigo

// #include "raylib.h"
// #include "raygui.h"
// #include <stdlib.h>
import "C"
import (
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func GuiLabel(bounds rl.Rectangle, title string) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.GuiLabel(cRectangle(bounds), cTitle)
}

func GuiButton(bounds rl.Rectangle, title string) bool {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	return bool(C.GuiButton(cRectangle(bounds), cTitle))
}

func GuiLabelButton(bounds rl.Rectangle, title string) bool {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	return bool(C.GuiLabelButton(cRectangle(bounds), cTitle))
}

func GuiToggle(bounds rl.Rectangle, title string, enabled bool) bool {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	return bool(C.GuiToggle(cRectangle(bounds), cTitle, C.bool(enabled)))
}

func GuiToggleGroup(bounds rl.Rectangle, title string, active int) int {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	return int(C.GuiToggleGroup(cRectangle(bounds), cTitle, C.int(active)))
}

func GuiCheckBox(bounds rl.Rectangle, title string, enabled bool) bool {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	return bool(C.GuiCheckBox(cRectangle(bounds), cTitle, C.bool(enabled)))
}

func GuiComboBox(bounds rl.Rectangle, title string, enabled int) int {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	return int(C.GuiComboBox(cRectangle(bounds), cTitle, C.int(enabled)))
}

func GuiDropdownBox(bounds rl.Rectangle, title string, active *int, editMode bool) bool {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	return bool(C.GuiDropdownBox(cRectangle(bounds), cTitle, cIntPtr(active), C.bool(editMode)))
}

func GuiSpinner(bounds rl.Rectangle, title string, value *int, minValue, maxValue int, editMode bool) bool {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	return bool(C.GuiSpinner(cRectangle(bounds), cTitle, cIntPtr(value), C.int(minValue), C.int(maxValue), C.bool(editMode)))
}

func GuiValueBox(bounds rl.Rectangle, title string, value *int, minValue, maxValue int, editMode bool) bool {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	return bool(C.GuiValueBox(cRectangle(bounds), cTitle, cIntPtr(value), C.int(minValue), C.int(maxValue), C.bool(editMode)))
}

func GuiTextBox(bounds rl.Rectangle, text string, editMode bool) (bool, string) {
	var cText [256]C.char
	for i, v := range text {
		cText[i] = C.char(v)
	}
	ret := bool(C.GuiTextBox(cRectangle(bounds), &cText[0], 256, C.bool(editMode)))
	var data [256]byte
	var i int
	for _, v := range cText {
		if v == 0 {
			break
		}
		data[i] = byte(v)
		i++
	}
	return ret, string(data[:i])
}

func GuiTextBoxMulti(bounds rl.Rectangle, text string, editMode bool) (bool, string) {
	var cText [4096]C.char
	for i, v := range text {
		cText[i] = C.char(v)
	}
	ret := bool(C.GuiTextBoxMulti(cRectangle(bounds), &cText[0], 4096, C.bool(editMode)))
	var data [4096]byte
	var i int
	for _, v := range cText {
		if v == 0 {
			break
		}
		data[i] = byte(v)
		i++
	}
	return ret, string(data[:i])
}

func GuiSlider(bounds rl.Rectangle, textLeft, textRight string, value, minValue, maxValue float32) float32 {
	cTextLeft := C.CString(textLeft)
	defer C.free(unsafe.Pointer(cTextLeft))
	cTextRight := C.CString(textRight)
	defer C.free(unsafe.Pointer(cTextRight))
	return float32(C.GuiSlider(cRectangle(bounds), cTextLeft, cTextRight, cFloat(value), cFloat(minValue), cFloat(maxValue)))
}

func GuiSliderBar(bounds rl.Rectangle, textLeft, textRight string, value, minValue, maxValue float32) float32 {
	cTextLeft := C.CString(textLeft)
	defer C.free(unsafe.Pointer(cTextLeft))
	cTextRight := C.CString(textRight)
	defer C.free(unsafe.Pointer(cTextRight))
	return float32(C.GuiSliderBar(cRectangle(bounds), cTextLeft, cTextRight, cFloat(value), cFloat(minValue), cFloat(maxValue)))
}

func GuiProgressBar(bounds rl.Rectangle, textLeft, textRight string, value, minValue, maxValue float32) float32 {
	cTextLeft := C.CString(textLeft)
	defer C.free(unsafe.Pointer(cTextLeft))
	cTextRight := C.CString(textRight)
	defer C.free(unsafe.Pointer(cTextRight))
	return float32(C.GuiProgressBar(cRectangle(bounds), cTextLeft, cTextRight, cFloat(value), cFloat(minValue), cFloat(maxValue)))
}

func GuiStatusBar(bounds rl.Rectangle, text string) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	C.GuiStatusBar(cRectangle(bounds), cText)
}

func GuiDummyRec(bounds rl.Rectangle, text string) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	C.GuiDummyRec(cRectangle(bounds), cText)
}

func GuiScrollBar(bounds rl.Rectangle, value, minValue, maxValue int) int {
	return int(C.GuiScrollBar(cRectangle(bounds), C.int(value), C.int(minValue), C.int(maxValue)))
}
