package main

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	rg "github.com/jlortiz0/multisav/raygui-go"
)

const CHOICEMENU_SPACE_BETWEEN_ITEM = 14

type Menu interface {
	HandleKey(int32) LoopStatus
	Prerender() LoopStatus
	Renderer()
	Destroy()
	RecalcTarget()
}

type LoopStatus int

const (
	LOOP_CONT LoopStatus = iota
	LOOP_EXIT
	LOOP_BACK
	LOOP_QUIT
)

type ChoiceMenu struct {
	itemList []string
	scroll   rl.Vector2
	target   rl.Rectangle
	height   float32
	Selected int
	status   LoopStatus
}

func NewChoiceMenu(items []string) *ChoiceMenu {
	menu := new(ChoiceMenu)
	menu.itemList = items
	menu.target = GetCenteredCoiceMenuRect(len(items), float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight()))
	menu.height = float32(len(items) * (TEXT_SIZE + CHOICEMENU_SPACE_BETWEEN_ITEM))
	menu.status = LOOP_CONT
	menu.Selected = -1
	return menu
}

func (cm *ChoiceMenu) Renderer() {
	flag := cm.height > cm.target.Height
	if flag {
		view := rg.GuiScrollPanel(cm.target, rl.Rectangle{Height: cm.height, Width: cm.target.Width - 16}, &cm.scroll)
		view2 := view.ToInt32()
		rl.BeginScissorMode(view2.X, view2.Y, view2.Width, view2.Height)
		cm.target.Width -= 10
	} else {
		rl.DrawRectangle(int32(cm.target.X), int32(cm.target.Y), int32(cm.target.Width), int32(cm.height), rl.RayWhite)
	}

	calc := cm.scroll.Y + CHOICEMENU_SPACE_BETWEEN_ITEM/2
	for i, x := range cm.itemList {
		if !flag || (calc > -TEXT_SIZE-2 && calc < cm.target.Height) {
			if i == cm.Selected {
				rg.GuiSetState(rg.GUI_STATE_FOCUSED)
			}
			if (rg.GuiButton(rl.Rectangle{X: cm.target.X + 5 + cm.scroll.X, Y: cm.target.Y + calc, Width: cm.target.Width - 10, Height: TEXT_SIZE + 2}, x) && cm.status == LOOP_CONT) {
				cm.Selected = i
				cm.status = LOOP_EXIT
			}
			if i == cm.Selected {
				rg.GuiSetState(rg.GUI_STATE_NORMAL)
			}
		}
		calc += CHOICEMENU_SPACE_BETWEEN_ITEM + TEXT_SIZE
	}

	if flag {
		rl.EndScissorMode()
		cm.target.Width += 10
	}
}

func (cm *ChoiceMenu) Prerender() LoopStatus {
	return cm.status
}

func (cm *ChoiceMenu) HandleKey(keycode int32) LoopStatus {
	switch keycode {
	case rl.KeyEscape:
		cm.Selected = -1
		return LOOP_BACK
	case rl.KeyDown:
		cm.Selected++
		if cm.Selected >= len(cm.itemList) {
			cm.Selected = 0
		}
		calc := cm.scroll.Y + CHOICEMENU_SPACE_BETWEEN_ITEM/2 + (CHOICEMENU_SPACE_BETWEEN_ITEM+TEXT_SIZE)*float32(cm.Selected)
		if calc < 0 {
			cm.scroll.Y = -CHOICEMENU_SPACE_BETWEEN_ITEM/2 - (CHOICEMENU_SPACE_BETWEEN_ITEM+TEXT_SIZE)*float32(cm.Selected) + TEXT_SIZE/2
		} else if calc >= cm.target.Height-TEXT_SIZE {
			cm.scroll.Y = -CHOICEMENU_SPACE_BETWEEN_ITEM/2 - (CHOICEMENU_SPACE_BETWEEN_ITEM+TEXT_SIZE)*float32(cm.Selected+1) + cm.target.Height
		}
	case rl.KeyUp:
		cm.Selected--
		if cm.Selected < 0 {
			cm.Selected = len(cm.itemList) - 1
		}
		calc := cm.scroll.Y + CHOICEMENU_SPACE_BETWEEN_ITEM/2 + (CHOICEMENU_SPACE_BETWEEN_ITEM+TEXT_SIZE)*float32(cm.Selected)
		if calc < 0 {
			cm.scroll.Y = -CHOICEMENU_SPACE_BETWEEN_ITEM/2 - (CHOICEMENU_SPACE_BETWEEN_ITEM+TEXT_SIZE)*float32(cm.Selected) + TEXT_SIZE/2
		} else if calc >= cm.target.Height-TEXT_SIZE {
			cm.scroll.Y = -CHOICEMENU_SPACE_BETWEEN_ITEM/2 - (CHOICEMENU_SPACE_BETWEEN_ITEM+TEXT_SIZE)*float32(cm.Selected+1) + cm.target.Height
		}
	case rl.KeyHome:
		cm.Selected = 0
		cm.scroll.Y = -CHOICEMENU_SPACE_BETWEEN_ITEM/2 + TEXT_SIZE + 2
	case rl.KeyEnd:
		cm.Selected = len(cm.itemList) - 1
		cm.scroll.Y = cm.target.Height - cm.height
	case rl.KeyEnter:
		if cm.Selected >= 0 {
			return LOOP_EXIT
		}
	case rl.KeyPageDown:
		if cm.Selected == len(cm.itemList)-1 {
			return cm.HandleKey(rl.KeyDown)
		}
		cm.Selected += int(cm.target.Height / (TEXT_SIZE + CHOICEMENU_SPACE_BETWEEN_ITEM))
		if cm.Selected >= len(cm.itemList) {
			cm.Selected = len(cm.itemList) - 1
		}
		calc := cm.scroll.Y + CHOICEMENU_SPACE_BETWEEN_ITEM/2 + (CHOICEMENU_SPACE_BETWEEN_ITEM+TEXT_SIZE)*float32(cm.Selected)
		if calc >= cm.target.Height-TEXT_SIZE {
			cm.scroll.Y = -CHOICEMENU_SPACE_BETWEEN_ITEM/2 - (CHOICEMENU_SPACE_BETWEEN_ITEM+TEXT_SIZE)*float32(cm.Selected+1) + cm.target.Height
		}
	case rl.KeyPageUp:
		if cm.Selected == 0 {
			return cm.HandleKey(rl.KeyUp)
		}
		cm.Selected -= int(cm.target.Height / (TEXT_SIZE + CHOICEMENU_SPACE_BETWEEN_ITEM))
		if cm.Selected < 0 {
			cm.Selected = 0
		}
		calc := cm.scroll.Y + CHOICEMENU_SPACE_BETWEEN_ITEM/2 + (CHOICEMENU_SPACE_BETWEEN_ITEM+TEXT_SIZE)*float32(cm.Selected)
		if calc < 0 {
			cm.scroll.Y = -CHOICEMENU_SPACE_BETWEEN_ITEM/2 - (CHOICEMENU_SPACE_BETWEEN_ITEM+TEXT_SIZE)*float32(cm.Selected) + TEXT_SIZE/2
		}
	case rl.KeyTab:
		if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
			return cm.HandleKey(rl.KeyUp)
		} else {
			return cm.HandleKey(rl.KeyDown)
		}
	}
	return LOOP_CONT
}

func (*ChoiceMenu) Destroy() {}

func (cm *ChoiceMenu) RecalcTarget() {
	cm.target = GetCenteredCoiceMenuRect(len(cm.itemList), float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight()))
}

func GetCenteredCoiceMenuRect(len int, width float32, height float32) rl.Rectangle {
	rect := rl.Rectangle{X: width / 4, Width: width / 2}
	mHeight := float32(len * (TEXT_SIZE + CHOICEMENU_SPACE_BETWEEN_ITEM))
	// Padding area of height / 8 on both borders
	if mHeight >= height*0.75 {
		rect.Y = height / 8
		rect.Height = height * 0.75
	} else {
		rect.Height = mHeight
		rect.Y = (height - mHeight) / 2
	}
	return rect
}

var fadeScreen rl.RenderTexture2D

func fadeOut(renderer func()) {
	if fadeScreen.ID == 0 {
		fadeScreen = rl.LoadRenderTexture(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()))
	} else if fadeScreen.Texture.Height != int32(rl.GetScreenHeight()) || fadeScreen.Texture.Width != int32(rl.GetScreenWidth()) {
		rl.UnloadRenderTexture(fadeScreen)
		fadeScreen = rl.LoadRenderTexture(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()))
	}
	rl.BeginTextureMode(fadeScreen)
	rl.ClearBackground(color.RGBA{})
	renderer()
	rl.EndTextureMode()
}

func fadeIn(renderer func()) {
	if fadeScreen.ID == 0 {
		return
	}
	render2 := rl.LoadRenderTexture(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()))
	w, h := float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())
	rl.BeginTextureMode(render2)
	rl.ClearBackground(color.RGBA{})
	renderer()
	rl.EndTextureMode()
	i2 := uint8(0)
	for i := uint8(0); i >= i2; i += 24 {
		// just let me check the overflow flag
		i2 = i
		rl.BeginDrawing()
		rl.ClearBackground(color.RGBA{64, 64, 64, 0})
		rl.DrawTextureRec(fadeScreen.Texture, rl.Rectangle{Height: -h, Width: w}, rl.Vector2{}, color.RGBA{255, 255, 255, ^i})
		rl.DrawTextureRec(render2.Texture, rl.Rectangle{Height: -h, Width: w}, rl.Vector2{}, color.RGBA{255, 255, 255, i})
		rl.EndDrawing()
	}
	rl.UnloadRenderTexture(render2)
}

type Message struct {
	txt rl.Texture2D
}

func NewMessage(msg string) *Message {
	img := drawMessage(msg)
	out := &Message{rl.LoadTextureFromImage(img)}
	rl.UnloadImage(img)
	return out
}

func (*Message) Prerender() LoopStatus {
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		return LOOP_BACK
	}
	return LOOP_CONT
}

func (*Message) RecalcTarget() {}

func (msg *Message) Destroy() { rl.UnloadTexture(msg.txt) }

func (msg *Message) Renderer() {
	x := (int32(rl.GetScreenWidth()) - msg.txt.Width) / 2
	y := (int32(rl.GetScreenHeight()) - msg.txt.Height) / 2
	rl.DrawTexture(msg.txt, x, y, rl.White)
}

func (*Message) HandleKey(keycode int32) LoopStatus {
	if keycode == rl.KeyX || keycode == rl.KeyEscape {
		return LOOP_BACK
	}
	if keycode == rl.KeyZ || keycode == rl.KeyEnter {
		return LOOP_EXIT
	}
	return LOOP_CONT
}
