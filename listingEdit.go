package main

import (
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/sqweek/dialog"

	rl "github.com/gen2brain/raylib-go/raylib"
	rg "github.com/jlortiz0/multisav/raygui-go"
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
	return &ListingEditMenu{*cm, LEM_EDIT}
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
			if i == cm.Selected {
				rg.GuiSetState(rg.GUI_STATE_FOCUSED)
			}
			if i < len(cm.itemList)-2 {
				if (rg.GuiButton(rl.Rectangle{X: cm.target.X + 5 + cm.scroll.X, Y: cm.target.Y + calc, Width: cm.target.Width - 10 - LEM_BUTTON_SIZE*2 - LEM_SPACE_BETWEEN_BUTTONS*2, Height: TEXT_SIZE + 2}, x) && cm.status == LOOP_CONT) {
					cm.Selected = i
					cm.status = LOOP_EXIT
					cm.Rem = LEM_EDIT
				}
				if i == cm.Selected {
					rg.GuiSetState(rg.GUI_STATE_NORMAL)
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
			} else if (rg.GuiButton(rl.Rectangle{X: cm.target.X + 5 + cm.scroll.X, Y: cm.target.Y + calc, Width: cm.target.Width - 10, Height: TEXT_SIZE + 2}, x) && cm.status == LOOP_CONT) {
				cm.Selected = i
				cm.status = LOOP_EXIT
			}
			rg.GuiSetState(rg.GUI_STATE_NORMAL)
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
		cm := NewChoiceMenu([]string{"Local", "Reddit", "Pixiv", "Lemmy", "Cancel"})
		if stdEventLoop(cm) == LOOP_QUIT {
			return true
		}
		sel := cm.Selected
		fadeOut(cm.Renderer)
		cm.Destroy()
		var site ImageSite
		var sKind int
		switch sel {
		case 0:
			db := dialog.Directory()
			s, _ := os.Getwd()
			s, err := db.SetStartDir(s).Title("Select image folder").Browse()
			if err == nil {
				if s[len(s)-1] == os.PathSeparator {
					s = s[:len(s)-1]
				}
				ind := strings.LastIndexByte(s, os.PathSeparator)
				saveData.Listings = append(saveData.Listings, SavedListing{Site: SITE_LOCAL, Args: []interface{}{s}, Name: filepath.Base(s[ind+1:])})
			}
			return false
		case 1:
			site = siteReddit
			sKind = SITE_REDDIT
		case 2:
			site = sitePixiv
			sKind = SITE_PIXIV
		case 3:
			site = siteLemmy
			sKind = SITE_LEMMY
		default:
			return false
		}
		kind, args := SetUpListing(site)
		if kind != -1 {
			saveData.Listings = append(saveData.Listings, SavedListing{Kind: kind, Site: sKind, Args: args[1:], Name: args[0].(string)})
		}
	} else if rem == LEM_REMOVE {
		copy(saveData.Listings[sel:], saveData.Listings[sel+1:])
		saveData.Listings = saveData.Listings[:len(saveData.Listings)-1]
	} else if rem == LEM_RESET {
		saveData.Listings[sel].Persistent = nil
	} else {
		data := saveData.Listings[sel]
		var args []ListingArgument
		if data.Site == SITE_LOCAL {
			args = []ListingArgument{
				{name: "Name"},
				{name: "Reselect", kind: LARGTYPE_BOOL},
			}
		} else if data.Site == SITE_TWITTER_OBSELETE {
			args = []ListingArgument{
				{kind: LARGTYPE_LABEL, options: []interface{}{"Twitter is no longer supported"}},
			}
		} else {
			var infoLs []ListingInfo
			var sName string
			switch data.Site {
			case SITE_REDDIT:
				infoLs = siteReddit.GetListingInfo()
				sName = "Reddit"
			case SITE_PIXIV:
				infoLs = sitePixiv.GetListingInfo()
				sName = "Pixiv"
			case SITE_LEMMY:
				infoLs = siteLemmy.GetListingInfo()
				sName = "Lemmy"
			default:
				panic("unknown site")
			}
			info := infoLs[data.Kind]
			args = make([]ListingArgument, 3, len(info.args)+3)
			args[0] = ListingArgument{"Site", []interface{}{sName}, LARGTYPE_LABEL}
			args[1] = ListingArgument{"Kind", []interface{}{info.name}, LARGTYPE_LABEL}
			args[2] = ListingArgument{name: "Name"}
			args = append(args, info.args...)
		}
		var cArgs []interface{}
		sel2 := -1
		if data.Site == SITE_LOCAL {
			cArgs = []interface{}{data.Name, false}
		} else if data.Site == SITE_TWITTER_OBSELETE {
			cArgs = []interface{}{nil}
		} else {
			cArgs = make([]interface{}, 3, len(data.Args)+3)
			cArgs[0] = 0
			cArgs[2] = data.Name
			cArgs = append(cArgs, data.Args...)
			for i, x := range args[3:] {
				if len(x.options) != 0 && x.kind == LARGTYPE_STRING {
					thing := cArgs[i+3].(string)
					for i2, x2 := range x.options {
						if thing == x2.(string) {
							cArgs[i+3] = i2
							break
						}
					}
				}
			}
		}
		fadeIn(func() { DrawArgumentsUI(data.Name, args, cArgs, &sel2) })
		for !rl.WindowShouldClose() {
			rl.BeginDrawing()
			rl.ClearBackground(color.RGBA{R: 64, G: 64, B: 64})
			out := DrawArgumentsUI(data.Name, args, cArgs, &sel2)
			rl.EndDrawing()
			if out != nil && len(out) == 0 {
				fadeOut(func() { DrawArgumentsUI(data.Name, args, cArgs, &sel2) })
				return EditListings()
			} else if len(out) != 0 {
				fadeOut(func() { DrawArgumentsUI(data.Name, args, cArgs, &sel2) })
				if data.Site == SITE_LOCAL {
					saveData.Listings[sel].Name = cArgs[0].(string)
					if cArgs[1].(bool) {
						db := dialog.Directory()
						dir := data.Args[0].(string)
						if !filepath.IsAbs(dir) {
							tmp, _ := os.Getwd()
							dir = filepath.Join(tmp, dir)
						}
						s, err := db.SetStartDir(dir).Title("Select image folder").Browse()
						if err == nil {
							wd, _ := os.Getwd()
							if os.PathSeparator == '\\' {
								wd = strings.ToUpper(wd[:1]) + wd[1:]
							}
							if strings.HasPrefix(s, wd) {
								s = s[len(wd):]
								if s[0] == os.PathSeparator {
									s = s[1:]
								}
							}
							saveData.Listings[sel].Args[0] = s
						}
					}
				} else if data.Site != SITE_TWITTER_OBSELETE {
					saveData.Listings[sel].Args = cArgs[3:]
					saveData.Listings[sel].Name = cArgs[2].(string)
				}
				return false
			}
		}
		return true
	}
	return false
}
