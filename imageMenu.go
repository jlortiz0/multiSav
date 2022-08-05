package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type OfflineImageMenu struct {
	Selected     int
	itemList     []string
	target       rl.Rectangle
	img          *rl.Image
	texture      rl.Texture2D
	prevMoveDir  bool
	state        imageMenuState
	fldr         string
	loadingFrame int
	cam          rl.Camera2D
	tol          rl.Vector2
	// ffmpeg *void
}

type imageMenuState int

const (
	IMSTATE_NORMAL imageMenuState = iota
	IMSTATE_SHOULDLOAD
	IMSTATE_LOADING
	IMSTATE_ERROR
	IMSTATE_SHOULDEXIT
)

func minf32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func getZoomForTexture(tex rl.Texture2D, target rl.Rectangle) float32 {
	return minf32(target.Height/float32(tex.Height), target.Width/float32(tex.Width))
}

func NewOfflineImageMenu(fldr string, target rl.Rectangle) (*OfflineImageMenu, error) {
	f, err := os.Open(fldr)
	if err != nil {
		return nil, err
	}
	entries, err := f.ReadDir(0)
	if err != nil {
		return nil, err
	}
	ls := make([]string, 0, len(entries))
	for _, v := range entries {
		if !v.IsDir() {
			ind := strings.LastIndexByte(v.Name(), '.')
			if ind == -1 {
				continue
			}
			switch strings.ToLower(v.Name()[ind+1:]) {
			// case "mp4":
			// 	fallthrough
			// case "webm":
			// 	fallthrough
			// case "gif":
			// 	fallthrough
			// case "mov":
			// 	fallthrough
			case "bmp":
				fallthrough
			case "jpg":
				fallthrough
			case "png":
				fallthrough
			case "jpeg":
				ls = append(ls, v.Name())
			}
		}
	}
	sort.Strings(ls)
	menu := new(OfflineImageMenu)
	menu.fldr = fldr
	menu.itemList = ls
	menu.state = IMSTATE_SHOULDLOAD
	menu.target = target
	menu.cam.Offset = rl.Vector2{Y: target.Height / 2, X: target.Width / 2}
	menu.cam.Zoom = 1
	if len(ls) == 0 {
		menu.state = IMSTATE_ERROR
		if len(entries) == 0 {
			menu.texture = drawMessage("(no supported images)")
		} else {
			menu.texture = drawMessage("(this page intentionally left blank)")
		}
	}
	return menu, nil
}

func (menu *OfflineImageMenu) loadImage() LoopStatus {
	if len(menu.itemList) == 0 {
		return LOOP_EXIT
	}
	menu.state = IMSTATE_LOADING
	go func() {
		_, err := os.Stat(menu.fldr + string(os.PathSeparator) + menu.itemList[menu.Selected])
		for err != nil {
			if len(menu.itemList) == 1 {
				menu.state = IMSTATE_SHOULDEXIT
				menu.itemList = nil
				return
			}
			if menu.Selected == len(menu.itemList)-1 {
				menu.Selected--
			} else {
				copy(menu.itemList[menu.Selected:], menu.itemList[menu.Selected+1:])
				if menu.prevMoveDir && menu.Selected > 0 {
					menu.Selected--
				}
			}
			menu.itemList = menu.itemList[:len(menu.itemList)-1]
			_, err = os.Stat(menu.fldr + string(os.PathSeparator) + menu.itemList[menu.Selected])
		}
		menu.img = rl.LoadImage(menu.fldr + string(os.PathSeparator) + menu.itemList[menu.Selected])
		menu.state = IMSTATE_NORMAL
	}()
	return LOOP_CONT
}

func (menu *OfflineImageMenu) HandleKey(keycode int32) LoopStatus {
	if keycode == rl.KeyEscape {
		return LOOP_BACK
	}
	if menu.state != IMSTATE_NORMAL {
		return LOOP_CONT
	}
	switch keycode {
	case rl.KeyLeft:
		if menu.Selected > 0 {
			menu.Selected--
			menu.state = IMSTATE_SHOULDLOAD
		}
	case rl.KeyRight:
		if menu.Selected+1 < len(menu.itemList) {
			menu.Selected++
			menu.state = IMSTATE_SHOULDLOAD
		}
	case rl.KeyF3:
		menu.cam.Target = rl.Vector2{Y: float32(menu.texture.Height) / 2, X: float32(menu.texture.Width) / 2}
		menu.cam.Zoom = getZoomForTexture(menu.texture, menu.target)
		menu.tol = rl.Vector2{Y: menu.cam.Offset.Y / menu.cam.Zoom, X: menu.cam.Offset.X / menu.cam.Zoom}
	case rl.KeyF1:
		fmt.Println(menu.cam.Target)
		fmt.Println(menu.cam.Zoom)
		fmt.Println(menu.texture.Height, menu.texture.Width)
		fmt.Println(menu.tol)
	}
	return LOOP_CONT
}

func (menu *OfflineImageMenu) Prerender() LoopStatus {
	if len(menu.itemList) == 0 || menu.state == IMSTATE_SHOULDEXIT {
		return LOOP_EXIT
	}
	if menu.state == IMSTATE_SHOULDLOAD {
		menu.loadImage()
		return menu.Prerender()
	}
	if menu.state != IMSTATE_NORMAL {
		return LOOP_CONT
	}
	if menu.img != nil {
		if menu.texture.ID > 0 {
			rl.UnloadTexture(menu.texture)
		}
		menu.texture = rl.LoadTextureFromImage(menu.img)
		rl.UnloadImage(menu.img)
		menu.img = nil
		rl.SetTextureFilter(menu.texture, rl.FilterBilinear)
		menu.cam.Target = rl.Vector2{Y: float32(menu.texture.Height) / 2, X: float32(menu.texture.Width) / 2}
		menu.cam.Zoom = getZoomForTexture(menu.texture, menu.target)
		menu.tol = rl.Vector2{Y: menu.cam.Offset.Y / menu.cam.Zoom, X: menu.cam.Offset.X / menu.cam.Zoom}
	}
	if rl.IsKeyDown(rl.KeyA) && menu.cam.Zoom*float32(menu.texture.Width) > menu.target.Width {
		menu.cam.Target.X -= 6.5 / menu.cam.Zoom
		if menu.cam.Target.X < menu.tol.X {
			menu.cam.Target.X = menu.tol.X
		}
	}
	if rl.IsKeyDown(rl.KeyD) && menu.cam.Zoom*float32(menu.texture.Width) > menu.target.Width {
		menu.cam.Target.X += 6.5 / menu.cam.Zoom
		if float32(menu.texture.Width)-menu.cam.Target.X < menu.tol.X {
			menu.cam.Target.X = float32(menu.texture.Width) - menu.tol.X
		}
	}
	if rl.IsKeyDown(rl.KeyW) && menu.cam.Zoom*float32(menu.texture.Height) > menu.target.Height {
		menu.cam.Target.Y -= 6.5 / menu.cam.Zoom
		if menu.cam.Target.Y < menu.tol.Y {
			menu.cam.Target.Y = menu.tol.Y
		}
	}
	if rl.IsKeyDown(rl.KeyS) && menu.cam.Zoom*float32(menu.texture.Height) > menu.target.Height {
		menu.cam.Target.Y += 6.5 / menu.cam.Zoom
		if float32(menu.texture.Height)-menu.cam.Target.Y < menu.tol.Y {
			menu.cam.Target.Y = float32(menu.texture.Height) - menu.tol.Y
		}
	}
	if rl.IsKeyDown(rl.KeyDown) && menu.cam.Zoom > 0.1 {
		menu.cam.Zoom -= 0.03125
		menu.tol = rl.Vector2{Y: menu.cam.Offset.Y / menu.cam.Zoom, X: menu.cam.Offset.X / menu.cam.Zoom}
		if menu.tol.Y > float32(menu.texture.Height)/2 {
			menu.cam.Target.Y = float32(menu.texture.Height) / 2
		} else if menu.cam.Target.Y < menu.tol.Y {
			menu.cam.Target.Y = menu.tol.Y
		} else if float32(menu.texture.Height)-menu.cam.Target.Y < menu.tol.Y {
			menu.cam.Target.Y = float32(menu.texture.Height) - menu.tol.Y
		}
		if menu.tol.X > float32(menu.texture.Width)/2 {
			menu.cam.Target.X = float32(menu.texture.Width) / 2
		} else if menu.cam.Target.X < menu.tol.X {
			menu.cam.Target.X = menu.tol.X
		} else if float32(menu.texture.Width)-menu.cam.Target.X < menu.tol.X {
			menu.cam.Target.X = float32(menu.texture.Width) - menu.tol.X
		}
	}
	if rl.IsKeyDown(rl.KeyUp) && menu.cam.Zoom < 6 {
		menu.cam.Zoom += 0.03125
		menu.tol = rl.Vector2{Y: menu.cam.Offset.Y / menu.cam.Zoom, X: menu.cam.Offset.X / menu.cam.Zoom}
	}
	if rl.IsWindowResized() {
		menu.cam.Target = rl.Vector2{Y: float32(menu.texture.Height) / 2, X: float32(menu.texture.Width) / 2}
		menu.cam.Zoom = getZoomForTexture(menu.texture, menu.target)
		menu.tol = rl.Vector2{Y: menu.cam.Offset.Y / menu.cam.Zoom, X: menu.cam.Offset.X / menu.cam.Zoom}
	}
	return LOOP_CONT
}

func (menu *OfflineImageMenu) Renderer() {
	switch menu.state {
	case IMSTATE_LOADING:
		rl.BeginMode2D(menu.cam)
		rl.DrawTexture(menu.texture, 0, 0, rl.White)
		rl.EndMode2D()
		x := int32(menu.target.Width)/2 - 50 + int32(menu.target.X)
		y := int32(menu.target.Height)/2 - 50 + int32(menu.target.Y)
		rl.DrawRectangle(x, y, 100, 100, rl.RayWhite)
		if menu.loadingFrame < 10 {
			rl.DrawRectangle(x+int32(menu.loadingFrame)*5, y, 50, 50, rl.Black)
			rl.DrawRectangle(x+50-int32(menu.loadingFrame)*5, y+50, 50, 50, rl.Black)
		} else if menu.loadingFrame < 16 {
			rl.DrawRectangle(x+50, y, 50, 50, rl.Black)
			rl.DrawRectangle(x, y+50, 50, 50, rl.Black)
		} else if menu.loadingFrame < 26 {
			rl.DrawRectangle(x+50, y-80+int32(menu.loadingFrame)*5, 50, 50, rl.Black)
			rl.DrawRectangle(x, y+130-int32(menu.loadingFrame)*5, 50, 50, rl.Black)
		} else {
			rl.DrawRectangle(x, y, 50, 50, rl.Black)
			rl.DrawRectangle(x+50, y+50, 50, 50, rl.Black)
		}
		menu.loadingFrame++
		menu.loadingFrame %= 32
	case IMSTATE_ERROR:
		fallthrough
	case IMSTATE_NORMAL:
		rl.BeginMode2D(menu.cam)
		rl.DrawTexture(menu.texture, 0, 0, rl.White)
		rl.EndMode2D()
	}
}

func (menu *OfflineImageMenu) Cleanup() {
	if menu.texture.ID <= 0 {
		rl.UnloadTexture(menu.texture)
	}
	if menu.img != nil {
		rl.UnloadImage(menu.img)
	}
}
