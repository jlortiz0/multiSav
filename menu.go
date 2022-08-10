package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	rg "jlortiz.org/redisav/raygui-go"
)

const CHOICEMENU_SPACE_BETWEEN_ITEM = 14

type Menu interface {
	HandleKey(int32) LoopStatus
	Prerender() LoopStatus
	Renderer()
	Cleanup()
	SetTarget(rl.Rectangle)
}

type LoopStatus int

const (
	LOOP_CONT LoopStatus = iota
	LOOP_EXIT
	LOOP_BACK
	LOOP_QUIT
)

type ChoiceMenu struct {
	Selected int
	itemList []string
	scroll   rl.Vector2
	target   rl.Rectangle
	height   float32
}

func NewChoiceMenu(items []string, target rl.Rectangle) *ChoiceMenu {
	menu := new(ChoiceMenu)
	menu.itemList = items
	menu.target = target
	menu.height = float32(len(items) * (TEXT_SIZE + CHOICEMENU_SPACE_BETWEEN_ITEM))
	return menu
}

func (cm *ChoiceMenu) Renderer() {
	if cm.height > cm.target.Height {
		view := rg.GuiScrollPanel(cm.target, rl.Rectangle{Height: cm.height, Width: cm.target.Width - 16}, &cm.scroll)
		view2 := view.ToInt32()
		rl.BeginScissorMode(view2.X, view2.Y, view2.Width, view2.Height)
		cm.target.Width -= 10
	} else {
		rl.DrawRectangle(int32(cm.target.X), int32(cm.target.Y), int32(cm.target.Width), int32(cm.height), rl.RayWhite)
	}

	for i, x := range cm.itemList {
		rg.GuiButton(rl.Rectangle{X: cm.target.X + 5 + cm.scroll.X, Y: cm.target.Y + cm.scroll.Y + float32(i)*(TEXT_SIZE+CHOICEMENU_SPACE_BETWEEN_ITEM) + CHOICEMENU_SPACE_BETWEEN_ITEM/2, Width: cm.target.Width - 10, Height: TEXT_SIZE + 2}, x)
	}

	if cm.height > cm.target.Height {
		rl.EndScissorMode()
		cm.target.Width += 10
	}
	// DEBUG
	// rl.DrawRectangleLinesEx(rl.Rectangle{X: cm.target.X - 5, Y: cm.target.Y - 5, Width: cm.target.Width + 10, Height: cm.target.Height + 10}, 5, rl.Magenta)
}

func (cm *ChoiceMenu) Prerender() LoopStatus {
	return LOOP_CONT
}

func (cm *ChoiceMenu) HandleKey(keycode int32) LoopStatus {
	if keycode == rl.KeyEscape {
		return LOOP_BACK
	}
	return LOOP_CONT
}

func (cm *ChoiceMenu) Cleanup() {}

func (cm *ChoiceMenu) SetTarget(target rl.Rectangle) {
	cm.target = target
}
