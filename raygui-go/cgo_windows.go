// +build windows

package rayguigo

// #include "raylib.h"
// #define RAYGUI_IMPLEMENTATION
// #define RAYGUI_STATIC
// #include "raygui.h"
// #cgo LDFLAGS: -lraylib -lm -lgdi32 -lwinmm
import "C"
