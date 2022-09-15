package rayguigo

// #include "raylib.h"
// #include "raygui.h"
// #include <stdlib.h>
import "C"
import (
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func GuiListView(bounds rl.Rectangle, text string, scrollIndex *int, active int) int {
	txt := C.CString(text)
	defer C.free(unsafe.Pointer(txt))
	return int(C.GuiListView(cRectangle(bounds), txt, (*C.int)(unsafe.Pointer(scrollIndex)), C.int(active)))
}

func GuiListViewEx(bounds rl.Rectangle, text []string, count int, focus *int, scrollIndex *int, active int) int {
	txt := make([](*C.char), len(text))
	for i, x := range text {
		txt[i] = C.CString(x)
		defer C.free(unsafe.Pointer(txt[i]))
	}
	return int(C.GuiListViewEx(cRectangle(bounds), (**C.char)(unsafe.Pointer(&txt[0])), C.int(count), (*C.int)(unsafe.Pointer(focus)), (*C.int)(unsafe.Pointer(scrollIndex)), C.int(active)))
}

func GuiMessageBox(bounds rl.Rectangle, title string, message string, buttons string) int {
	titleC := C.CString(title)
	defer C.free(unsafe.Pointer(titleC))
	msgC := C.CString(message)
	defer C.free(unsafe.Pointer(msgC))
	butC := C.CString(buttons)
	defer C.free(unsafe.Pointer(butC))
	return int(C.GuiMessageBox(cRectangle(bounds), titleC, msgC, butC))
}
