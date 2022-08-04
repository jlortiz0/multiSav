package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	rg "jlortiz.org/redisav/raygui-go"
)

const CHOICEMENU_SPACE_BETWEEN_ITEM = 16

type Menu interface {
	HandleKey(int32) LoopStatus
	Renderer(rl.Rectangle)
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
}

func NewChoiceMenu(items []string) *ChoiceMenu {
	menu := new(ChoiceMenu)
	menu.itemList = items
	return menu
}

func (cm *ChoiceMenu) Renderer(target rl.Rectangle) {
	bigger := len(cm.itemList)*(int(font.BaseSize)+CHOICEMENU_SPACE_BETWEEN_ITEM) > int(target.Height)
	if bigger {
		view := rg.GuiScrollPanel(target, rl.Rectangle{Height: float32(len(cm.itemList) * (int(font.BaseSize) + CHOICEMENU_SPACE_BETWEEN_ITEM)), Width: target.Width - 16}, &cm.scroll)
		view2 := view.ToInt32()
		rl.BeginScissorMode(view2.X, view2.Y, view2.Width, view2.Height)
		target.Width -= 10
		target.X += 5
	}

	for i, x := range cm.itemList {
		rg.GuiButton(rl.Rectangle{X: target.X + cm.scroll.X, Y: target.Y + cm.scroll.Y + float32(i)*(float32(font.BaseSize)+CHOICEMENU_SPACE_BETWEEN_ITEM), Width: target.Width - 10, Height: float32(font.BaseSize)}, x)
	}

	if bigger {
		rl.EndScissorMode()
	}
	// DEBUG
	// rl.DrawRectangleLinesEx(rl.Rectangle{X: target.X - 5, Y: target.Y - 5, Width: target.Width + 10, Height: target.Height + 10}, 5, rl.Magenta)
}

func (cm *ChoiceMenu) SuggestWidth() float32 {
	var longest string
	for _, x := range cm.itemList {
		if len(x) > len(longest) {
			longest = x
		}
	}
	vec := rl.MeasureTextEx(font, longest, float32(font.BaseSize), 0)
	return vec.X * 1.5
}

func (cm *ChoiceMenu) HandleKey(keycode int32) LoopStatus {
	return LOOP_CONT
}
