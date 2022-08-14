package main

import (
	"fmt"
	"image/color"
	"os"
	"sort"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	rg "jlortiz.org/redisav/raygui-go"
)

type ImageMenu struct {
	Selected int
	Producer ImageProducer
	target   rl.Rectangle
	img      *rl.Image
	texture  rl.Texture2D
	// prevMoveDir  bool
	state        imageMenuState
	loadingFrame int
	cam          rl.Camera2D
	tol          rl.Vector2
	tempSel      int
	ffmpeg       *ffmpegReader
	fName        string
}

type imageMenuState int

const (
	IMSTATE_NORMAL     imageMenuState = 0
	IMSTATE_SHOULDLOAD imageMenuState = 1
	IMSTATE_LOADING    imageMenuState = 2
	IMSTATE_ERROR      imageMenuState = 4
	IMSTATE_ERRORMINOR imageMenuState = 8
	IMSTATE_SHOULDEXIT imageMenuState = 16
	IMSTATE_GOTO       imageMenuState = 32
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

func NewOfflineImageMenu(fldr string, target rl.Rectangle) (*ImageMenu, error) {
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
			case "mp4":
				fallthrough
			case "webm":
				fallthrough
			case "gif":
				fallthrough
			case "mov":
				fallthrough
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
	menu := new(ImageMenu)
	menu.Producer = &OfflineImageProducer{ls, fldr}
	rl.SetWindowTitle(menu.Producer.GetTitle())
	menu.state = IMSTATE_SHOULDLOAD
	menu.target = target
	menu.target.Height -= TEXT_SIZE + 10
	menu.cam.Offset = rl.Vector2{Y: target.Height/2 - 5 - TEXT_SIZE/2, X: target.Width / 2}
	menu.cam.Zoom = 1
	if len(ls) == 0 {
		menu.state = IMSTATE_ERROR
		if len(entries) == 0 {
			menu.texture = drawMessage("No supported files found.")
		} else {
			menu.texture = drawMessage("Empty.")
		}
		menu.cam.Target = rl.Vector2{Y: float32(menu.texture.Height) / 2, X: float32(menu.texture.Width) / 2}
	}
	return menu, nil
}

func NewImageMenu(prod ImageProducer, target rl.Rectangle) *ImageMenu {
	menu := new(ImageMenu)
	menu.Producer = prod
	rl.SetWindowTitle(menu.Producer.GetTitle())
	menu.state = IMSTATE_SHOULDLOAD
	menu.target = target
	menu.target.Height -= TEXT_SIZE + 10
	menu.cam.Offset = rl.Vector2{Y: target.Height/2 - 5 - TEXT_SIZE/2, X: target.Width / 2}
	menu.cam.Zoom = 1
	return menu
}

func (menu *ImageMenu) loadImage() {
	if menu.Producer.Len() == 0 {
		if !menu.Producer.IsLazy() || !menu.Producer.BoundsCheck(0) {
			menu.state = IMSTATE_SHOULDEXIT
			return
		}
	}
	if menu.ffmpeg != nil {
		menu.ffmpeg.Destroy()
		menu.ffmpeg = nil
	}
	menu.state = IMSTATE_LOADING
	go func() {
		menu.fName = menu.Producer.Get(menu.Selected, &menu.img, &menu.ffmpeg)
		if (menu.img == nil || menu.img.Height == 0) && menu.ffmpeg == nil {
			if !menu.Producer.BoundsCheck(menu.Selected) {
				menu.Selected = menu.Producer.Len() - 1
				menu.loadImage()
			}
		} else if menu.fName[:5] == "\\/err" {
			menu.state = IMSTATE_ERRORMINOR
			menu.fName = menu.fName[5:]
		} else {
			menu.state = IMSTATE_NORMAL
		}
	}()
}

func (menu *ImageMenu) HandleKey(keycode int32) LoopStatus {
	if keycode == rl.KeyEscape {
		return LOOP_BACK
	}
	if menu.state&IMSTATE_ERRORMINOR != 0 {
		switch keycode {
		case rl.KeyLeft:
			if menu.Selected > 0 {
				menu.Selected--
				menu.state = IMSTATE_SHOULDLOAD
			}
		case rl.KeyRight:
			if !menu.Producer.IsLazy() || menu.Selected+1 < menu.Producer.Len() {
				menu.Selected++
				menu.state = IMSTATE_SHOULDLOAD
			}
		}
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
		if !menu.Producer.IsLazy() || menu.Selected+1 < menu.Producer.Len() {
			menu.Selected++
			menu.state = IMSTATE_SHOULDLOAD
		}
	case rl.KeyHome:
		if menu.Selected != 0 {
			menu.Selected = 0
			menu.state = IMSTATE_SHOULDLOAD
		}
	case rl.KeyEnd:
		if menu.Selected != menu.Producer.Len()-1 {
			menu.Selected = menu.Producer.Len() - 1
			menu.state = IMSTATE_SHOULDLOAD
		}
	case rl.KeyF3:
		menu.cam.Target = rl.Vector2{Y: float32(menu.texture.Height) / 2, X: float32(menu.texture.Width) / 2}
		menu.cam.Zoom = getZoomForTexture(menu.texture, menu.target)
		menu.tol = rl.Vector2{Y: menu.cam.Offset.Y / menu.cam.Zoom, X: menu.cam.Offset.X / menu.cam.Zoom}
	case rl.KeySpace:
		fmt.Println(menu.cam.Target, menu.cam.Zoom, menu.tol)
	default:
		call := 0
		state := ARET_AGAIN
		for state&ARET_AGAIN != 0 {
			state = menu.Producer.ActionHandler(keycode, menu.Selected, call)
			if state&ARET_MOVEDOWN != 0 {
				moveFactor := float32(25)
				for menu.cam.Target.Y > menu.tol.Y-menu.target.Height-float32(menu.texture.Height) {
					menu.cam.Target.Y -= moveFactor
					if moveFactor < menu.target.Height {
						moveFactor *= 1.18
					}
					rl.BeginDrawing()
					rl.DrawRectangleRec(menu.target, color.RGBA{R: 64, G: 64, B: 64, A: 255})
					menu.Renderer()
					rl.EndDrawing()
				}
			} else if state&ARET_MOVEUP != 0 {
				moveFactor := float32(25)
				for menu.cam.Target.Y < menu.tol.Y+float32(menu.texture.Height) {
					menu.cam.Target.Y += moveFactor
					if moveFactor < menu.target.Height {
						moveFactor *= 1.18
					}
					rl.BeginDrawing()
					rl.DrawRectangleRec(menu.target, color.RGBA{R: 64, G: 64, B: 64, A: 255})
					menu.Renderer()
					rl.EndDrawing()
				}
			}
			if state&ARET_CLOSEFFMPEG != 0 && menu.ffmpeg != nil {
				menu.ffmpeg.Destroy()
				menu.ffmpeg = nil
			}
			if state&ARET_REMOVE != 0 {
				menu.state = IMSTATE_SHOULDLOAD
			}
			call += 1
		}
	}
	return LOOP_CONT
}

func (menu *ImageMenu) Prerender() LoopStatus {
	if menu.Producer.Len() == 0 {
		if !menu.Producer.IsLazy() || !menu.Producer.BoundsCheck(0) {
			return LOOP_EXIT
		}
	}
	if menu.state == IMSTATE_SHOULDEXIT {
		return LOOP_EXIT
	}
	if menu.state&IMSTATE_SHOULDLOAD != 0 {
		menu.loadImage()
		return menu.Prerender()
	}
	if menu.state&IMSTATE_ERRORMINOR != 0 && menu.img != nil {
		if menu.texture.ID > 0 {
			rl.UnloadTexture(menu.texture)
		}
		menu.texture = rl.LoadTextureFromImage(menu.img)
		rl.UnloadImage(menu.img)
		menu.img = nil
		menu.cam.Target = rl.Vector2{Y: float32(menu.texture.Height) / 2, X: float32(menu.texture.Width) / 2}
		menu.cam.Zoom = 1
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
	if menu.ffmpeg != nil {
		data, err := menu.ffmpeg.Read()
		if err != nil {
			if menu.texture.ID > 0 {
				rl.UnloadTexture(menu.texture)
			}
			menu.texture = drawMessage(err.Error())
			menu.cam.Target = rl.Vector2{Y: float32(menu.texture.Height) / 2, X: float32(menu.texture.Width) / 2}
			menu.state = IMSTATE_ERRORMINOR
			return LOOP_CONT
		}
		data2 := make([]color.RGBA, len(data)/3)
		for i := range data2 {
			data2[i] = color.RGBA{R: data[i*3], G: data[i*3+1], B: data[i*3+2], A: 255}
		}
		rl.UpdateTexture(menu.texture, data2)
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
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		v := rl.GetMousePosition()
		v.X -= menu.target.X
		v.Y -= menu.target.Y
		if v.X < menu.target.Width/16 {
			if rl.CheckCollisionPointCircle(v, rl.Vector2{Y: menu.target.Height / 2}, TEXT_SIZE*2) {
				menu.HandleKey(rl.KeyLeft)
			}
		} else if v.X-menu.target.Width > -menu.target.Width/16 {
			if rl.CheckCollisionPointCircle(v, rl.NewVector2(menu.target.Width, menu.target.Height/2), TEXT_SIZE*2) {
				menu.HandleKey(rl.KeyRight)
			}
		}
	}
	return LOOP_CONT
}

func (menu *ImageMenu) Renderer() {
	if menu.state&IMSTATE_LOADING != 0 {
		rl.BeginMode2D(menu.cam)
		rl.DrawTexture(menu.texture, 0, 0, rl.White)
		rl.EndMode2D()
		// x := int32(menu.target.Width)/2 - 50 + int32(menu.target.X)
		// y := int32(menu.target.Height)/2 - 50 + int32(menu.target.Y)
		// rl.DrawRectangle(x, y, 100, 100, rl.RayWhite)
		// if menu.loadingFrame < 10 {
		// 	rl.DrawRectangle(x+int32(menu.loadingFrame)*5, y, 50, 50, rl.Black)
		// 	rl.DrawRectangle(x+50-int32(menu.loadingFrame)*5, y+50, 50, 50, rl.Black)
		// } else if menu.loadingFrame < 16 {
		// 	rl.DrawRectangle(x+50, y, 50, 50, rl.Black)
		// 	rl.DrawRectangle(x, y+50, 50, 50, rl.Black)
		// } else if menu.loadingFrame < 26 {
		// 	rl.DrawRectangle(x+50, y-80+int32(menu.loadingFrame)*5, 50, 50, rl.Black)
		// 	rl.DrawRectangle(x, y+130-int32(menu.loadingFrame)*5, 50, 50, rl.Black)
		// } else {
		// 	rl.DrawRectangle(x, y, 50, 50, rl.Black)
		// 	rl.DrawRectangle(x+50, y+50, 50, 50, rl.Black)
		// }
		// menu.loadingFrame++
		// menu.loadingFrame %= 32
		rl.DrawRectangle(int32(menu.target.X), int32(menu.target.Height), int32(menu.target.Width), TEXT_SIZE+10, rl.Black)
		if menu.loadingFrame < 0 {
			rl.DrawRectangle(int32(menu.target.X)-int32(menu.loadingFrame)*5, int32(menu.target.Height)+5, TEXT_SIZE, TEXT_SIZE, rl.Green)
			rl.DrawRectangle(int32(menu.target.X)-int32(menu.loadingFrame+4)*5, int32(menu.target.Height)+5, TEXT_SIZE, TEXT_SIZE, rl.Green)
			rl.DrawRectangle(int32(menu.target.X)-int32(menu.loadingFrame+8)*5, int32(menu.target.Height)+5, TEXT_SIZE, TEXT_SIZE, rl.Green)
			if menu.loadingFrame == -9 {
				menu.loadingFrame = -1
			}
		} else {
			rl.DrawRectangle(int32(menu.target.X)+int32(menu.loadingFrame+1)*5, int32(menu.target.Height)+5, TEXT_SIZE, TEXT_SIZE, rl.Green)
			rl.DrawRectangle(int32(menu.target.X)+int32(menu.loadingFrame+5)*5, int32(menu.target.Height)+5, TEXT_SIZE, TEXT_SIZE, rl.Green)
			rl.DrawRectangle(int32(menu.target.X)+int32(menu.loadingFrame+9)*5, int32(menu.target.Height)+5, TEXT_SIZE, TEXT_SIZE, rl.Green)
			if TEXT_SIZE+(menu.loadingFrame+9)*5 > int(menu.target.Width) {
				menu.loadingFrame = -menu.loadingFrame - 7
			}
		}
		menu.loadingFrame++
	} else {
		rl.BeginMode2D(menu.cam)
		rl.DrawTexture(menu.texture, 0, 0, rl.White)
		rl.EndMode2D()
		if menu.state&IMSTATE_ERROR == 0 {
			rl.DrawRectangle(int32(menu.target.X), int32(menu.target.Height), int32(menu.target.Width), TEXT_SIZE+10, rl.Black)
			s := fmt.Sprintf("%dx%d (%.0fx%.0f)", menu.texture.Width, menu.texture.Height,
				float32(menu.texture.Width)*menu.cam.Zoom, float32(menu.texture.Height)*menu.cam.Zoom)
			vec := rl.NewVector2(5+menu.target.X, menu.target.Height+5+menu.target.Y)
			rl.DrawTextEx(font, s, vec, TEXT_SIZE, 0, rl.RayWhite)
			vec = rl.MeasureTextEx(font, menu.fName, TEXT_SIZE, 0)
			vec.Y = menu.target.Height + 5 + menu.target.Y
			vec.X = menu.target.X + menu.target.Width/2 - vec.X/2
			rl.DrawTextEx(font, menu.fName, vec, TEXT_SIZE, 0, rl.RayWhite)
			if menu.state&IMSTATE_GOTO == 0 {
				s := fmt.Sprintf("%d/%d", menu.Selected+1, menu.Producer.Len())
				if menu.Producer.IsLazy() {
					s += "+"
				}
				if (rg.GuiLabelButton(rl.Rectangle{X: menu.target.X + menu.target.Width - 75, Y: menu.target.Height, Width: 70, Height: TEXT_SIZE + 10}, s)) {
					menu.state |= IMSTATE_GOTO
					menu.tempSel = menu.Selected + 1
				}
			} else {
				if rg.GuiValueBox(rl.Rectangle{X: menu.target.X + menu.target.Width - 75, Y: menu.target.Height, Width: 75, Height: TEXT_SIZE + 10}, "goto", &menu.tempSel, 1, menu.Producer.Len(), true) {
					menu.state &= ^IMSTATE_GOTO
					menu.Selected = menu.tempSel - 1
					if menu.Selected < 0 {
						menu.Selected = 0
					}
					// if menu.Selected >= len(menu.itemList) {
					// 	menu.Selected = len(menu.itemList) - 1
					// }
					menu.loadImage()
				}
			}
			y := rl.GetMouseY() - int32(menu.target.Y)
			if y > int32(menu.target.Height)/4 && y-int32(menu.target.Height)/4 < int32(menu.target.Height)/2 {
				x := rl.GetMouseX() - int32(menu.target.X)
				if x < int32(menu.target.Width)/8 && menu.Selected > 0 {
					a := float32(x) - menu.target.Width/16
					if a < 0 {
						a = 0
					} else {
						a = a / (menu.target.Width / 16) * 255
					}
					rl.DrawCircleV(rl.NewVector2(menu.target.X, menu.target.Y+menu.target.Height/2), TEXT_SIZE*2, color.RGBA{250, 250, 250, 255 - uint8(a)})
					rl.DrawCircleLines(int32(menu.target.X), int32(menu.target.Y+menu.target.Height/2), TEXT_SIZE*2, color.RGBA{A: 255 - uint8(a)})
					rl.DrawLineEx(rl.Vector2{X: menu.target.X + TEXT_SIZE, Y: menu.target.Y + menu.target.Height/2 - TEXT_SIZE},
						rl.Vector2{X: menu.target.X + menu.target.Width/128, Y: menu.target.Y + menu.target.Height/2}, TEXT_SIZE/4, color.RGBA{128, 128, 128, 255 - uint8(a)})
					rl.DrawLineEx(rl.Vector2{X: menu.target.X + menu.target.Width/128, Y: menu.target.Y + menu.target.Height/2},
						rl.Vector2{X: menu.target.X + TEXT_SIZE, Y: menu.target.Y + menu.target.Height/2 + TEXT_SIZE}, TEXT_SIZE/4, color.RGBA{128, 128, 128, 255 - uint8(a)})
				} else if x-int32(menu.target.Width) > -int32(menu.target.Width)/8 && (menu.Producer.IsLazy() || menu.Selected+1 < menu.Producer.Len()) {
					a := float32(x) - menu.target.Width + menu.target.Width/8
					if a > menu.target.Width/16 {
						a = 255
					} else {
						a = a / (menu.target.Width / 16) * 255
					}
					rl.DrawCircleV(rl.NewVector2(menu.target.X+menu.target.Width, menu.target.Y+menu.target.Height/2), TEXT_SIZE*2, color.RGBA{250, 250, 250, uint8(a)})
					rl.DrawCircleLines(int32(menu.target.X+menu.target.Width), int32(menu.target.Y+menu.target.Height/2), TEXT_SIZE*2, color.RGBA{A: uint8(a)})
					rl.DrawLineEx(rl.Vector2{X: menu.target.X + menu.target.Width - TEXT_SIZE, Y: menu.target.Y + menu.target.Height/2 + TEXT_SIZE},
						rl.Vector2{X: menu.target.X + menu.target.Width - menu.target.Width/128, Y: menu.target.Y + menu.target.Height/2}, TEXT_SIZE/4, color.RGBA{128, 128, 128, uint8(a)})
					rl.DrawLineEx(rl.Vector2{X: menu.target.X + +menu.target.Width - menu.target.Width/128, Y: menu.target.Y + menu.target.Height/2},
						rl.Vector2{X: menu.target.X + menu.target.Width - TEXT_SIZE, Y: menu.target.Y + menu.target.Height/2 - TEXT_SIZE}, TEXT_SIZE/4, color.RGBA{128, 128, 128, uint8(a)})
				}
			}
		}
	}
}

func (menu *ImageMenu) Cleanup() {
	if menu.texture.ID > 0 {
		rl.UnloadTexture(menu.texture)
	}
	if menu.img != nil {
		rl.UnloadImage(menu.img)
	}
	if menu.ffmpeg != nil {
		menu.ffmpeg.Destroy()
	}
}

func (menu *ImageMenu) SetTarget(target rl.Rectangle) {
	menu.target = target
	menu.target.Height -= TEXT_SIZE + 10
	menu.cam.Offset = rl.Vector2{Y: target.Height/2 - 5 - TEXT_SIZE/2, X: target.Width / 2}
	menu.cam.Zoom = getZoomForTexture(menu.texture, menu.target)
	menu.tol = rl.Vector2{Y: menu.cam.Offset.Y / menu.cam.Zoom, X: menu.cam.Offset.X / menu.cam.Zoom}
}
