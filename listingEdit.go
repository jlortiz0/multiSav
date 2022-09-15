package main

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	rg "jlortiz.org/redisav/raygui-go"
)

const (
	LEM_BUTTON_SIZE           = TEXT_SIZE + 2
	LEM_SPACE_BETWEEN_BUTTONS = 5
	LEM_EDIT                  = iota
	LEM_RESET
	LEM_REMOVE
)

type ListingEditMenu struct {
	ChoiceMenu
	Rem int
}

func NewListingEditMenu(ls []SavedListing) *ListingEditMenu {
	items := make([]string, len(ls)+2)
	for i, x := range ls {
		items[i] = x.Name
	}
	items[len(ls)] = "New..."
	items[len(ls)+1] = "Back"
	cm := NewChoiceMenu(items)
	return &ListingEditMenu{*cm, 0}
}

func (cm *ListingEditMenu) Renderer() {
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
			if (rg.GuiButton(rl.Rectangle{X: cm.target.X + 5 + cm.scroll.X, Y: cm.target.Y + calc, Width: cm.target.Width - 10 - LEM_BUTTON_SIZE*2 - LEM_SPACE_BETWEEN_BUTTONS*2, Height: TEXT_SIZE + 2}, x) && cm.status == LOOP_CONT) {
				cm.Selected = i
				cm.status = LOOP_EXIT
				cm.Rem = LEM_EDIT
			}
			if (rg.GuiButton(rl.Rectangle{X: cm.target.X + 5 + cm.scroll.X + cm.target.Width - 10 - LEM_BUTTON_SIZE*2 - LEM_SPACE_BETWEEN_BUTTONS, Y: cm.target.Y + calc, Width: LEM_BUTTON_SIZE, Height: LEM_BUTTON_SIZE}, "#28#") && cm.status == LOOP_CONT) {
				cm.Selected = i
				cm.status = LOOP_EXIT
				cm.Rem = LEM_RESET
			}
			if (rg.GuiButton(rl.Rectangle{X: cm.target.X + 5 + cm.scroll.X + cm.target.Width - 10 - LEM_BUTTON_SIZE, Y: cm.target.Y + calc, Width: LEM_BUTTON_SIZE, Height: LEM_BUTTON_SIZE}, "#143#") && cm.status == LOOP_CONT) {
				cm.Selected = i
				cm.status = LOOP_EXIT
				cm.Rem = LEM_REMOVE
			}
		}
		calc += CHOICEMENU_SPACE_BETWEEN_ITEM + TEXT_SIZE
	}

	if flag {
		rl.EndScissorMode()
		cm.target.Width += 10
	}
}

func EditListings() bool {
	lem := NewListingEditMenu(saveData.Listings)
	ret := stdEventLoop(lem)
	rem := lem.Rem
	sel := lem.Selected
	lem.Destroy()
	if ret == LOOP_QUIT {
		return true
	} else if ret == LOOP_BACK || sel == len(saveData.Listings)+1 {
		return false
	}
	if sel == len(saveData.Listings) {
		cm := NewChoiceMenu([]string{"Local", "Reddit", "Twitter", "Pixiv", "Cancel"})
		if stdEventLoop(cm) == LOOP_QUIT {
			return true
		}
		sel := cm.Selected
		fadeOut(cm.Renderer)
		cm.Destroy()
		var site ImageSite
		switch sel {
		case 0:
			// TODO: Reimplement file dialog from raygui in Go
			return false
		case 1:
			site = siteReddit
		default:
			return false
		}
		kind, args := SetUpListing(site)
		if kind != -1 {
			saveData.Listings = append(saveData.Listings, SavedListing{Kind: kind, Site: 1, Args: args[1:], Name: args[0].(string)})
		}
	} else if rem == LEM_REMOVE {
		copy(saveData.Listings[sel:], saveData.Listings[sel+1:])
		saveData.Listings = saveData.Listings[:len(saveData.Listings)-1]
	} else if rem == LEM_RESET {
		saveData.Listings[sel].Persistent = nil
	} else {
		data := saveData.Listings[sel]
		var args []ListingArgument
		switch data.Site {
		case SITE_LOCAL:
		case SITE_REDDIT:
			info := siteReddit.GetListingInfo()[data.Kind]
			args = make([]ListingArgument, 2, len(info.args)+2)
			args[0] = ListingArgument{"Site", LARGTYPE_LABEL, []interface{}{"Reddit"}}
			args[1] = ListingArgument{"Kind", LARGTYPE_LABEL, []interface{}{info.name}}
			args = append(args, info.args...)
		case SITE_PIXIV:
		case SITE_TWITTER:
		default:
			panic("unknown site")
		}
		cArgs := make([]interface{}, 2, len(data.Args)+2)
		cArgs[0] = 0
		cArgs = append(cArgs, data.Args...)
		flags := make([]bool, len(data.Args)+2)
		fadeIn(func() { DrawArgumentsUI(data.Name, args, cArgs, flags) })
		for !rl.WindowShouldClose() {
			rl.BeginDrawing()
			rl.ClearBackground(color.RGBA{R: 64, G: 64, B: 64})
			out := DrawArgumentsUI(data.Name, args, cArgs, flags)
			rl.EndDrawing()
			if out != nil && len(out) == 0 {
				fadeOut(func() { DrawArgumentsUI(data.Name, args, cArgs, flags) })
				return false
			} else if len(out) != 0 {
				fadeOut(func() { DrawArgumentsUI(data.Name, args, cArgs, flags) })
				saveData.Listings[sel].Args = cArgs[2:]
				return false
			}
		}
		return true
	}
	return false
}
