package main

import (
	"fmt"
	"image/color"
	"math"
	"sync"

	"github.com/rainycape/unidecode"

	rl "github.com/gen2brain/raylib-go/raylib"
	rg "github.com/jlortiz0/multisav/raygui-go"
)

type ImageMenu struct {
	ffmpeg       VideoReader
	Producer     ImageProducer
	ffmpegData   chan []color.RGBA
	prodLoader   <-chan ImageProducer
	img          *rl.Image
	imgLock      *sync.Mutex
	fName        string
	Selected     int
	tempSel      int
	loadingFrame int
	target       rl.Vector2
	texture      rl.Texture2D
	cam          rl.Camera2D
	tol          rl.Vector2
	state        imageMenuState
	prevMoveDir  bool
}

type imageMenuState uint8

const (
	IMSTATE_NORMAL imageMenuState = iota
	IMSTATE_ERRORMINOR
	IMSTATE_SHOULDLOAD
	IMSTATE_LOADING
	IMSTATE_GOTO imageMenuState = 8
)

// 2^(1/24)
const ZOOM_STEP = 1.02930224

func minf32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func getZoomForTexture(tex rl.Texture2D, target rl.Vector2) float32 {
	return minf32(target.Y/float32(tex.Height), target.X/float32(tex.Width))
}

func NewImageMenu(prod func() <-chan ImageProducer) *ImageMenu {
	menu := new(ImageMenu)
	menu.prodLoader = prod()
	menu.state = IMSTATE_LOADING | IMSTATE_GOTO
	menu.target = rl.Vector2{X: float32(rl.GetScreenWidth()), Y: float32(rl.GetScreenHeight())}
	menu.target.Y -= TEXT_SIZE + 10
	menu.cam.Offset = rl.Vector2{Y: menu.target.Y / 2, X: menu.target.X / 2}
	menu.cam.Zoom = 1
	menu.imgLock = new(sync.Mutex)
	return menu
}

func (menu *ImageMenu) loadImage() {
	if menu.ffmpeg != nil {
		ffm := menu.ffmpeg
		ffd := menu.ffmpegData
		go func() {
			ffm.Destroy()
			<-ffd
		}()
		menu.ffmpeg = nil
	}
	if menu.Selected < 0 {
		menu.Selected = 0
	}
	if !menu.Producer.BoundsCheck(menu.Selected) {
		menu.Selected = menu.Producer.Len() - 1
	}
	menu.state = IMSTATE_LOADING
	go func() {
		menu.imgLock.Lock()
		menu.fName = menu.Producer.Get(menu.Selected, &menu.img, &menu.ffmpeg)
		menu.fName = unidecode.Unidecode(menu.fName)
		menu.imgLock.Unlock()
		if (menu.img == nil || menu.img.Height == 0) && menu.ffmpeg == nil {
			if !menu.Producer.BoundsCheck(menu.Selected) {
				menu.Selected = menu.Producer.Len() - 1
				menu.loadImage()
			}
			return
		} else if len(menu.fName) > 5 && menu.fName[:5] == "\\/err" {
			menu.state = IMSTATE_ERRORMINOR
			menu.fName = menu.fName[5:]
			return
		} else if menu.ffmpeg == nil {
			menu.state = IMSTATE_NORMAL
		} else {
			menu.ffmpegData = make(chan []color.RGBA)
			data, i, _ := menu.ffmpeg.Read()
			if i == nil {
				menu.state = IMSTATE_NORMAL
				menu.ffmpegData <- data
			} else {
				rl.UnloadImage(menu.img)
				menu.img = i
				menu.state = IMSTATE_NORMAL
			}
			for menu.ffmpeg != nil {
				data, i, err := menu.ffmpeg.Read()
				if err != nil {
					menu.ffmpeg.Destroy()
					menu.ffmpeg = nil
					text := err.Error()
					vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
					menu.imgLock.Lock()
					menu.img = rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
					rl.ImageDrawTextEx(menu.img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
					menu.imgLock.Unlock()
					menu.state = IMSTATE_ERRORMINOR
					break
				}
				if data != nil {
					menu.ffmpegData <- data
				} else if i != nil {
					menu.imgLock.Lock()
					if menu.img != nil {
						rl.UnloadImage(menu.img)
					}
					menu.img = i
					menu.imgLock.Unlock()
				}
			}
			close(menu.ffmpegData)
		}
	}()
}

func (menu *ImageMenu) HandleKey(keycode int32) LoopStatus {
	if keycode == rl.KeyEscape {
		return LOOP_BACK
	}
	if menu.state != IMSTATE_ERRORMINOR && menu.state != IMSTATE_NORMAL {
		return LOOP_CONT
	}
	switch keycode {
	case rl.KeyLeft:
		if menu.Selected > 0 {
			menu.Selected--
			menu.state = IMSTATE_SHOULDLOAD
			menu.prevMoveDir = true
		}
	case rl.KeyRight:
		if menu.Producer.IsLazy() || menu.Selected+1 < menu.Producer.Len() {
			menu.Selected++
			menu.state = IMSTATE_SHOULDLOAD
			menu.prevMoveDir = false
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
		if menu.state == IMSTATE_NORMAL {
			menu.cam.Target = rl.Vector2{Y: float32(menu.texture.Height) / 2, X: float32(menu.texture.Width) / 2}
			menu.cam.Zoom = getZoomForTexture(menu.texture, menu.target)
			menu.tol = rl.Vector2{Y: menu.cam.Offset.Y / menu.cam.Zoom, X: menu.cam.Offset.X / menu.cam.Zoom}
		}
	case rl.KeyZ:
		s := menu.Producer.GetInfo(menu.Selected)
		if s != "" {
			if messageOverlay(wordWrapper(s), menu) == LOOP_QUIT {
				return LOOP_QUIT
			}
		}
	case rl.KeyG:
		menu.state |= IMSTATE_GOTO
	default:
		call := 0
		state := ARET_AGAIN
		for state&ARET_AGAIN != 0 {
			state = menu.Producer.ActionHandler(keycode, menu.Selected, call)
			cam := menu.cam
			if state&ARET_QUIT != 0 {
				return LOOP_QUIT
			}
			if state&ARET_FADEIN != 0 {
				fadeIn(menu.Renderer)
			}
			if state&ARET_MOVEDOWN != 0 {
				moveFactor := float32(26)
				for menu.cam.Target.Y > menu.tol.Y-menu.target.Y-float32(menu.texture.Height) {
					menu.cam.Target.Y -= moveFactor / menu.cam.Zoom
					if moveFactor < menu.target.Y {
						moveFactor *= 1.1
					}
					rl.BeginDrawing()
					rl.DrawRectangleV(rl.Vector2{}, menu.target, color.RGBA{R: 64, G: 64, B: 64, A: 255})
					menu.Renderer()
					rl.EndDrawing()
				}
			} else if state&ARET_MOVEUP != 0 {
				moveFactor := float32(26)
				for menu.cam.Target.Y < menu.tol.Y+float32(menu.texture.Height) {
					menu.cam.Target.Y += moveFactor / menu.cam.Zoom
					if moveFactor < menu.target.Y {
						moveFactor *= 1.1
					}
					rl.BeginDrawing()
					rl.DrawRectangleV(rl.Vector2{}, menu.target, color.RGBA{R: 64, G: 64, B: 64, A: 255})
					menu.Renderer()
					rl.EndDrawing()
				}
			}
			if state&ARET_CLOSEFFMPEG != 0 && menu.ffmpeg != nil {
				menu.ffmpeg.Destroy()
				menu.ffmpeg = nil
				<-menu.ffmpegData
			}
			if state&ARET_REMOVE != 0 {
				menu.state = IMSTATE_SHOULDLOAD
				menu.cam.Zoom = 0
				if menu.prevMoveDir && menu.Selected != 0 {
					menu.Selected--
				}
			} else {
				menu.cam = cam
			}
			if state&ARET_FADEOUT != 0 {
				fadeOut(menu.Renderer)
			}
			call += 1
		}
	}
	return LOOP_CONT
}

func (menu *ImageMenu) Prerender() LoopStatus {
	if menu.prodLoader != nil {
		select {
		case menu.Producer = <-menu.prodLoader:
			if menu.Producer == nil {
				return LOOP_EXIT
			}
			menu.prodLoader = nil
			rl.SetWindowTitle(menu.Producer.GetTitle())
			menu.state = IMSTATE_SHOULDLOAD | IMSTATE_GOTO
		default:
			return LOOP_CONT
		}
	}
	if menu.state == IMSTATE_SHOULDLOAD|IMSTATE_GOTO {
		menu.state = IMSTATE_SHOULDLOAD
		if menu.Producer.Len() == 0 && (!menu.Producer.IsLazy() || !menu.Producer.BoundsCheck(0)) {
			fadeOut(menu.Renderer)
			msg := NewMessage("No images!")
			return stdEventLoop(msg)
		}

	}
	if menu.Producer.Len() == 0 {
		if !menu.Producer.IsLazy() || !menu.Producer.BoundsCheck(0) {
			return LOOP_EXIT
		}
	}
	if menu.state == IMSTATE_SHOULDLOAD {
		menu.loadImage()
		return menu.Prerender()
	}
	if menu.state == IMSTATE_ERRORMINOR && menu.img != nil {
		if menu.texture.ID > 0 {
			rl.UnloadTexture(menu.texture)
		}
		menu.imgLock.Lock()
		menu.texture = rl.LoadTextureFromImage(menu.img)
		rl.UnloadImage(menu.img)
		menu.img = nil
		menu.imgLock.Unlock()
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
		menu.imgLock.Lock()
		menu.texture = rl.LoadTextureFromImage(menu.img)
		rl.UnloadImage(menu.img)
		menu.img = nil
		menu.imgLock.Unlock()
		rl.SetTextureFilter(menu.texture, rl.FilterBilinear)
		menu.cam.Target = rl.Vector2{Y: float32(menu.texture.Height) / 2, X: float32(menu.texture.Width) / 2}
		menu.cam.Zoom = getZoomForTexture(menu.texture, menu.target)
		menu.tol = rl.Vector2{Y: menu.cam.Offset.Y / menu.cam.Zoom, X: menu.cam.Offset.X / menu.cam.Zoom}
	}
	select {
	case data := <-menu.ffmpegData:
		if len(data) > 0 {
			rl.UpdateTexture(menu.texture, data)
		}
	default:
	}
	if rl.IsKeyDown(rl.KeyA) && menu.cam.Zoom*float32(menu.texture.Width) > menu.target.X {
		menu.cam.Target.X -= 6.5 / menu.cam.Zoom
		if menu.cam.Target.X < menu.tol.X {
			menu.cam.Target.X = menu.tol.X
		}
	}
	if rl.IsKeyDown(rl.KeyD) && menu.cam.Zoom*float32(menu.texture.Width) > menu.target.X {
		menu.cam.Target.X += 6.5 / menu.cam.Zoom
		if float32(menu.texture.Width)-menu.cam.Target.X < menu.tol.X {
			menu.cam.Target.X = float32(menu.texture.Width) - menu.tol.X
		}
	}
	if rl.IsKeyDown(rl.KeyW) && menu.cam.Zoom*float32(menu.texture.Height) > menu.target.Y {
		menu.cam.Target.Y -= 6.5 / menu.cam.Zoom
		if menu.cam.Target.Y < menu.tol.Y {
			menu.cam.Target.Y = menu.tol.Y
		}
	}
	if rl.IsKeyDown(rl.KeyS) && menu.cam.Zoom*float32(menu.texture.Height) > menu.target.Y {
		menu.cam.Target.Y += 6.5 / menu.cam.Zoom
		if float32(menu.texture.Height)-menu.cam.Target.Y < menu.tol.Y {
			menu.cam.Target.Y = float32(menu.texture.Height) - menu.tol.Y
		}
	}
	if rl.IsKeyDown(rl.KeyDown) || rl.GetMouseWheelMove() < 0 && menu.cam.Zoom > 0.1 {
		menu.cam.Zoom /= ZOOM_STEP
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
	if rl.IsKeyDown(rl.KeyUp) || rl.GetMouseWheelMove() > 0 && menu.cam.Zoom < 6 {
		menu.cam.Zoom *= ZOOM_STEP
		menu.tol = rl.Vector2{Y: menu.cam.Offset.Y / menu.cam.Zoom, X: menu.cam.Offset.X / menu.cam.Zoom}
	}
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		v := rl.GetMousePosition()
		if v.X < menu.target.X/16 {
			if rl.CheckCollisionPointCircle(v, rl.Vector2{Y: menu.target.Y / 2}, TEXT_SIZE*2) {
				menu.HandleKey(rl.KeyLeft)
			}
		} else if v.X-menu.target.X > -menu.target.X/16 {
			if rl.CheckCollisionPointCircle(v, rl.NewVector2(menu.target.X, menu.target.Y/2), TEXT_SIZE*2) {
				menu.HandleKey(rl.KeyRight)
			}
		}
	}
	return LOOP_CONT
}

func (menu *ImageMenu) Renderer() {
	if menu.state&^IMSTATE_GOTO == IMSTATE_LOADING {
		rl.BeginMode2D(menu.cam)
		rl.DrawTexture(menu.texture, 0, 0, rl.White)
		rl.EndMode2D()
		// x := int32(menu.target.X)/2 - 50
		// y := int32(menu.target.Y)/2 - 50
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
		rl.DrawRectangle(0, int32(menu.target.Y), int32(menu.target.X), TEXT_SIZE+10, rl.Black)
		if menu.loadingFrame < 0 {
			rl.DrawRectangle(-int32(menu.loadingFrame)*5, int32(menu.target.Y)+5, TEXT_SIZE, TEXT_SIZE, rl.Green)
			rl.DrawRectangle(-int32(menu.loadingFrame+4)*5, int32(menu.target.Y)+5, TEXT_SIZE, TEXT_SIZE, rl.Green)
			rl.DrawRectangle(-int32(menu.loadingFrame+8)*5, int32(menu.target.Y)+5, TEXT_SIZE, TEXT_SIZE, rl.Green)
			if menu.loadingFrame == -9 {
				menu.loadingFrame = -1
			}
		} else {
			rl.DrawRectangle(int32(menu.loadingFrame+1)*5, int32(menu.target.Y)+5, TEXT_SIZE, TEXT_SIZE, rl.Green)
			rl.DrawRectangle(int32(menu.loadingFrame+5)*5, int32(menu.target.Y)+5, TEXT_SIZE, TEXT_SIZE, rl.Green)
			rl.DrawRectangle(int32(menu.loadingFrame+9)*5, int32(menu.target.Y)+5, TEXT_SIZE, TEXT_SIZE, rl.Green)
			if TEXT_SIZE+(menu.loadingFrame+9)*5 > int(menu.target.X) {
				menu.loadingFrame = -menu.loadingFrame - 7
			}
		}
		menu.loadingFrame++
	} else {
		rl.BeginMode2D(menu.cam)
		rl.DrawTexture(menu.texture, 0, 0, rl.White)
		rl.EndMode2D()
		rl.DrawRectangle(0, int32(menu.target.Y), int32(menu.target.X), TEXT_SIZE+10, rl.Black)
		s := fmt.Sprintf("%dx%d (%.0fx%.0f)", menu.texture.Width, menu.texture.Height,
			float32(menu.texture.Width)*menu.cam.Zoom, float32(menu.texture.Height)*menu.cam.Zoom)
		vec := rl.NewVector2(5, menu.target.Y+5)
		rl.DrawTextEx(font, s, vec, TEXT_SIZE, 0, rl.RayWhite)
		vec2 := float32(rl.MeasureText(menu.fName, TEXT_SIZE))
		vec.Y = menu.target.Y
		vec.X = menu.target.X/2 - vec2/2
		vec3 := rl.MeasureTextEx(font, s, TEXT_SIZE, 0)
		if vec2 > menu.target.X-86-vec3.X {
			c := float32(rl.MeasureText("...", TEXT_SIZE))
			try := menu.fName
			v := 0
			for vec2 > menu.target.X-90-c-vec3.X && v < 4 {
				per := int((menu.target.X - 90 - c - vec3.X) / vec2 * float32(len(try)))
				try = menu.fName[:per]
				vec2 = float32(rl.MeasureText(try, TEXT_SIZE))
				v++
			}
			menu.fName = try + "..."
			vec2 += c
			vec.X = menu.target.X/2 - vec2/2
		}
		if vec.X < vec3.X+10 {
			vec.X = vec3.X + 10
		}
		rg.GuiSetStyle(rg.LABEL, rg.TEXT_ALIGNMENT, rg.TEXT_ALIGN_LEFT)
		if rg.GuiLabelButton(rl.Rectangle{X: vec.X, Y: vec.Y, Height: TEXT_SIZE + 10, Width: vec2}, menu.fName) {
			s := menu.Producer.GetInfo(menu.Selected)
			if s != "" {
				rl.EndDrawing()
				// TODO: exit here too
				messageOverlay(wordWrapper(s), menu)
				rl.BeginDrawing()
			}
		}
		rg.GuiSetStyle(rg.LABEL, rg.TEXT_ALIGNMENT, rg.TEXT_ALIGN_RIGHT)
		if menu.state&IMSTATE_GOTO == 0 {
			s := "..."
			if menu.Producer != nil {
				s = fmt.Sprintf("%d/%d", menu.Selected+1, menu.Producer.Len())
				if menu.Producer.IsLazy() {
					s += "+"
				}
			}
			if (rg.GuiLabelButton(rl.Rectangle{X: menu.target.X - 75, Y: menu.target.Y, Width: 70, Height: TEXT_SIZE + 10}, s)) {
				menu.state |= IMSTATE_GOTO
				menu.tempSel = menu.Selected + 1
			}
		} else {
			var max int
			if menu.Producer != nil {
				max = menu.Producer.Len()
			}
			if rg.GuiValueBox(rl.Rectangle{X: menu.target.X - 75, Y: menu.target.Y, Width: 75, Height: TEXT_SIZE + 10}, "", &menu.tempSel, 0, max, true) {
				menu.state = IMSTATE_SHOULDLOAD
				menu.Selected = menu.tempSel - 1
				if menu.Selected < 0 {
					menu.Selected = 0
				}
			}
		}
		y := rl.GetMouseY()
		if y > int32(menu.target.Y)/4 && y-int32(menu.target.Y)/4 < int32(menu.target.Y)/2 {
			x := rl.GetMouseX()
			if x < int32(menu.target.X)/8 && menu.Selected > 0 {
				a := math.Hypot(float64(x), float64(y)-float64(menu.target.Y)/2) - float64(menu.target.X)/16
				if a < 0 {
					a = 255
				} else {
					a = 255 - a/(float64(menu.target.X)/16)*255
					if a < 0 {
						a = 0
					}
				}
				rl.DrawCircleV(rl.NewVector2(0, menu.target.Y/2), TEXT_SIZE*2, color.RGBA{250, 250, 250, uint8(a)})
				rl.DrawCircleLines(0, int32(menu.target.Y/2), TEXT_SIZE*2, color.RGBA{A: uint8(a)})
				rl.DrawLineEx(rl.Vector2{X: TEXT_SIZE, Y: menu.target.Y/2 - TEXT_SIZE},
					rl.Vector2{X: menu.target.X / 128, Y: menu.target.Y / 2}, TEXT_SIZE/4, color.RGBA{128, 128, 128, uint8(a)})
				rl.DrawLineEx(rl.Vector2{X: menu.target.X / 128, Y: menu.target.Y / 2},
					rl.Vector2{X: TEXT_SIZE, Y: menu.target.Y/2 + TEXT_SIZE}, TEXT_SIZE/4, color.RGBA{128, 128, 128, uint8(a)})
			} else if x-int32(menu.target.X) > -int32(menu.target.X)/8 && (menu.Producer.IsLazy() || menu.Selected+1 < menu.Producer.Len()) {
				a := math.Hypot(float64(x)-float64(menu.target.X), float64(y)-float64(menu.target.Y)/2) - float64(menu.target.X)/16
				if a < 0 {
					a = 255
				} else {
					a = 255 - a/(float64(menu.target.X)/16)*255
					if a < 0 {
						a = 0
					}
				}
				rl.DrawCircleV(rl.NewVector2(menu.target.X, menu.target.Y/2), TEXT_SIZE*2, color.RGBA{250, 250, 250, uint8(a)})
				rl.DrawCircleLines(int32(menu.target.X), int32(menu.target.Y/2), TEXT_SIZE*2, color.RGBA{A: uint8(a)})
				rl.DrawLineEx(rl.Vector2{X: menu.target.X - TEXT_SIZE, Y: menu.target.Y/2 + TEXT_SIZE},
					rl.Vector2{X: menu.target.X - menu.target.X/128, Y: menu.target.Y / 2}, TEXT_SIZE/4, color.RGBA{128, 128, 128, uint8(a)})
				rl.DrawLineEx(rl.Vector2{X: menu.target.X - menu.target.X/128, Y: menu.target.Y / 2},
					rl.Vector2{X: menu.target.X - TEXT_SIZE, Y: menu.target.Y/2 - TEXT_SIZE}, TEXT_SIZE/4, color.RGBA{128, 128, 128, uint8(a)})
			}
		}
	}
}

func (menu *ImageMenu) Destroy() {
	if menu.texture.ID > 0 {
		rl.UnloadTexture(menu.texture)
	}
	if menu.ffmpeg != nil {
		menu.ffmpeg.Destroy()
		menu.ffmpeg = nil
		<-menu.ffmpegData
	}
	if menu.img != nil {
		rl.UnloadImage(menu.img)
	}
	if menu.prodLoader != nil {
		go func() {
			<-menu.prodLoader
		}()
	} else {
		menu.Producer.Destroy()
	}
}

func (menu *ImageMenu) RecalcTarget() {
	menu.target = rl.Vector2{X: float32(rl.GetScreenWidth()), Y: float32(rl.GetScreenHeight())}
	menu.target.Y -= TEXT_SIZE + 10
	menu.cam.Offset = rl.Vector2{Y: menu.target.Y / 2, X: menu.target.X / 2}
	menu.cam.Zoom = getZoomForTexture(menu.texture, menu.target)
	menu.tol = rl.Vector2{Y: menu.cam.Offset.Y / menu.cam.Zoom, X: menu.cam.Offset.X / menu.cam.Zoom}
}
