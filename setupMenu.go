package main

import (
	"fmt"
	"image/color"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	rg "github.com/jlortiz0/multisav/raygui-go"
	"github.com/sqweek/dialog"
)

const ARGUI_SPACING = CHOICEMENU_SPACE_BETWEEN_ITEM * 1.5

func DrawArgumentsUI(name string, args []ListingArgument, out []interface{}, sel *int) []interface{} {
	if out[0] == nil {
		for i, v := range args {
			if len(v.options) != 0 {
				out[i] = 0
			} else {
				switch v.kind {
				case LARGTYPE_BOOL:
					out[i] = false
				case LARGTYPE_URL:
					fallthrough
				case LARGTYPE_STRING:
					out[i] = ""
				case LARGTYPE_INT:
					out[i] = 0
					// case LARGTYPE_TIME:
					// 	out[i] = time.Time{}
				}
			}
		}
		*sel = -1
	}
	target := rl.Vector2{X: float32(rl.GetScreenWidth()), Y: float32(rl.GetScreenHeight())}
	vec := rl.MeasureTextEx(font, name, TEXT_SIZE, 0)
	vec2 := rl.Vector2{X: ((target.X - vec.X) / 2), Y: target.Y/2 - (TEXT_SIZE+ARGUI_SPACING)*float32(len(args)+2)/2}
	rl.DrawRectangle(int32(target.X/4), int32(vec2.Y)-5, int32(target.X)/2, (TEXT_SIZE+ARGUI_SPACING)*int32(len(args)+2)+10, rl.RayWhite)
	vec2.Y += (ARGUI_SPACING + TEXT_SIZE) / 4
	rl.DrawTextEx(font, name, vec2, TEXT_SIZE, 0, rl.Black)
	for i, v := range args {
		vec2.Y += ARGUI_SPACING + TEXT_SIZE
		name := v.name + ":"
		vec = rl.MeasureTextEx(font, name, TEXT_SIZE, 0)
		vec2.X = target.X/2 - vec.X - 5
		rl.DrawTextEx(font, name, vec2, TEXT_SIZE, 0, rl.Black)
		if *sel == i {
			rg.GuiSetState(rg.GUI_STATE_FOCUSED)
		} else {
			rg.GuiSetState(rg.GUI_STATE_NORMAL)
		}
		if len(v.options) != 0 {
			if v.kind == LARGTYPE_LABEL {
				rl.DrawTextEx(font, args[i].options[0].(string), rl.Vector2{X: target.X/2 + 5, Y: vec2.Y}, TEXT_SIZE, 0, rl.Black)
				continue
			}
			s, ok := out[i].(string)
			if ok {
				rg.GuiComboBox(rl.Rectangle{X: target.X/2 + 5, Y: vec2.Y - 5, Width: target.X/4 - 10, Height: TEXT_SIZE + 10}, s, 0)
				continue
			}
			name := strings.Builder{}
			name.WriteString(fmt.Sprint(v.options[0]))
			for _, n := range v.options[1:] {
				name.WriteByte(';')
				name.WriteString(fmt.Sprint(n))
			}
			if *sel == i && rl.IsKeyPressed(rl.KeyEnter) {
				out[i] = out[i].(int) + 1
				if out[i].(int) >= len(v.options) {
					out[i] = 0
				}
			}
			out[i] = rg.GuiComboBox(rl.Rectangle{X: target.X/2 + 5, Y: vec2.Y - 5, Width: target.X/4 - 10, Height: TEXT_SIZE + 10}, name.String(), out[i].(int))
		} else {
			switch v.kind {
			case LARGTYPE_BOOL:
				if *sel == i && rl.IsKeyPressed(rl.KeyEnter) {
					out[i] = !(out[i].(bool))
				}
				temp := rg.GuiCheckBox(rl.Rectangle{X: target.X/2 + 5, Y: vec2.Y - 3, Width: TEXT_SIZE + 5, Height: TEXT_SIZE + 5}, "", out[i].(bool))
				if temp != out[i] {
					out[i] = temp
					*sel = i
				}
			case LARGTYPE_URL:
				// TODO: write a scrolling text box
				fallthrough
			case LARGTYPE_STRING:
				var ret bool
				ret, out[i] = rg.GuiTextBox(rl.Rectangle{X: target.X/2 + 5, Y: vec2.Y - 5, Width: target.X/4 - 10, Height: TEXT_SIZE + 10}, out[i].(string), *sel == i)
				if ret {
					if *sel == i {
						*sel = -1
					} else {
						*sel = i
					}
				}
			case LARGTYPE_INT:
				temp := out[i].(int)
				if rg.GuiValueBox(rl.Rectangle{X: target.X/2 + 5, Y: vec2.Y - 5, Width: target.X/8 - 10, Height: TEXT_SIZE + 10}, "", &temp, -999, 999, *sel == i) {
					if *sel == i {
						*sel = -1
					} else {
						*sel = i
					}
				}
				out[i] = temp
			}
		}
	}
	if rl.IsKeyPressed(rl.KeyTab) {
		if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
			*sel -= 1
			if *sel < 0 {
				*sel = len(args)
			}
		} else {
			*sel += 1
			if *sel > len(args) {
				*sel = 0
			}
		}
	}
	fini := rl.IsKeyPressed(rl.KeyEnter) && *sel == len(args)
	vec2.Y += (ARGUI_SPACING + TEXT_SIZE)
	rg.GuiSetState(rg.GUI_STATE_NORMAL)
	if rg.GuiButton(rl.Rectangle{X: target.X/4 + 10, Y: vec2.Y, Height: TEXT_SIZE + 5, Width: target.X/4 - 20}, "Cancel") || rl.IsKeyPressed(rl.KeyEscape) {
		return []interface{}{}
	}
	if *sel >= len(args) {
		rg.GuiSetState(rg.GUI_STATE_FOCUSED)
		defer rg.GuiSetState(rg.GUI_STATE_NORMAL)
	}
	if rg.GuiButton(rl.Rectangle{X: target.X/2 + 10, Y: vec2.Y, Height: TEXT_SIZE + 5, Width: target.X/4 - 20}, "Confirm") || fini {
		for i := range out {
			if len(args[i].options) != 0 {
				continue
			}
			switch args[i].kind {
			case LARGTYPE_STRING:
				if out[i].(string) == "" {
					return nil
				}
			case LARGTYPE_INT:
				// Do I need to do any verification here? 0 is probably valid...
			// case LARGTYPE_TIME:
			// 	if out[i].(time.Time).IsZero() {
			// 		return nil
			// 	}
			case LARGTYPE_URL:
				_, err := url.ParseRequestURI(out[i].(string))
				if err != nil {
					return nil
				}
			}
		}
		for i := range out {
			if len(args[i].options) != 0 && args[i].kind != LARGTYPE_LABEL {
				out[i] = args[i].options[out[i].(int)]
			}
		}
		return out
	}
	return nil
}

func SetUpListing(site ImageSite) (int, []interface{}) {
	choices := site.GetListingInfo()
	names := make([]string, len(choices)+1)
	for i, v := range choices {
		names[i] = v.name
	}
	names[len(choices)] = "Cancel"
cm:
	cm := NewChoiceMenu(names)
	if stdEventLoop(cm) == LOOP_QUIT {
		return -1, nil
	}
	kind := cm.Selected
	cm.Destroy()
	if kind == len(choices) {
		return -1, nil
	}
	// if len(choices[kind].args) == 0 {
	// 	return kind, nil
	// }
	args := make([]interface{}, len(choices[kind].args)+1)
	var sel int
	cArgs := make([]ListingArgument, 1, len(choices[kind].args)+1)
	cArgs[0].name = "Name"
	cArgs[0].kind = LARGTYPE_STRING
	cArgs = append(cArgs, choices[kind].args...)
	fadeIn(func() { DrawArgumentsUI(choices[kind].name, cArgs, args, &sel) })
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(color.RGBA{R: 64, G: 64, B: 64})
		out := DrawArgumentsUI(choices[kind].name, cArgs, args, &sel)
		rl.EndDrawing()
		if out != nil && len(out) == 0 {
			fadeOut(func() { DrawArgumentsUI(choices[kind].name, cArgs, args, &sel) })
			goto cm
		} else if len(out) != 0 {
			fadeOut(func() { DrawArgumentsUI(choices[kind].name, cArgs, args, &sel) })
			return kind, args
		}
	}
	return -1, nil
}

func SetUpSites() bool {
	cm := NewChoiceMenu([]string{"Downloads", "Options", "Logout", "Back"})
	ret := stdEventLoop(cm)
	if ret != LOOP_EXIT {
		return ret == LOOP_QUIT
	}
	kind := cm.Selected
	cm.Destroy()
	switch kind {
	case 3:
		return false
	case 0:
		db := dialog.Directory()
		dir := saveData.Downloads
		if !filepath.IsAbs(dir) {
			tmp, _ := os.Getwd()
			dir = filepath.Join(tmp, dir)
		}
		s, err := db.SetStartDir(dir).Title("Select downloads folder").Browse()
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
			saveData.Downloads = s
		}
	case 1:
		args := []interface{}{saveData.Settings.SaveOnX, saveData.Settings.HideOnZ, saveData.Settings.PixivBookPriv}
		sel := -1
		cArgs := []ListingArgument{
			{
				name: "Always download on X",
				kind: LARGTYPE_BOOL,
			},
			{
				name: "Hide background on Z",
				kind: LARGTYPE_BOOL,
			},
			{
				name: "Pixiv: Private bookmarks",
				kind: LARGTYPE_BOOL,
			},
		}
		fadeIn(func() { DrawArgumentsUI("Options", cArgs, args, &sel) })
		for !rl.WindowShouldClose() {
			rl.BeginDrawing()
			rl.ClearBackground(color.RGBA{R: 64, G: 64, B: 64})
			out := DrawArgumentsUI("Options", cArgs, args, &sel)
			rl.EndDrawing()
			if out != nil && len(out) == 0 {
				fadeOut(func() { DrawArgumentsUI("Options", cArgs, args, &sel) })
				break
			} else if len(out) != 0 {
				fadeOut(func() { DrawArgumentsUI("Options", cArgs, args, &sel) })
				saveData.Settings.SaveOnX = args[0].(bool)
				saveData.Settings.HideOnZ = args[1].(bool)
				saveData.Settings.PixivBookPriv = args[2].(bool)
				break
			}
		}
	case 2:
		args := []interface{}{false, false, false}
		sel := -1
		cArgs := make([]ListingArgument, 3)
		if saveData.Reddit == "" {
			cArgs[0] = ListingArgument{
				name:    "Reddit",
				kind:    LARGTYPE_LABEL,
				options: []interface{}{"Logged out"},
			}
		} else {
			cArgs[0] = ListingArgument{
				name: "Reddit",
				kind: LARGTYPE_BOOL,
			}
		}
		if saveData.Twitter == "" {
			cArgs[1] = ListingArgument{
				name:    "Twitter",
				kind:    LARGTYPE_LABEL,
				options: []interface{}{"Logged out"},
			}
		} else {
			cArgs[1] = ListingArgument{
				name: "Twitter",
				kind: LARGTYPE_BOOL,
			}
		}
		if saveData.Pixiv == "" {
			cArgs[2] = ListingArgument{
				name:    "Pixiv",
				kind:    LARGTYPE_LABEL,
				options: []interface{}{"Logged out"},
			}
		} else {
			cArgs[2] = ListingArgument{
				name: "Pixiv",
				kind: LARGTYPE_BOOL,
			}
		}
		fadeIn(func() { DrawArgumentsUI("Logout", cArgs, args, &sel) })
		for !rl.WindowShouldClose() {
			rl.BeginDrawing()
			rl.ClearBackground(color.RGBA{R: 64, G: 64, B: 64})
			out := DrawArgumentsUI("Logout", cArgs, args, &sel)
			rl.EndDrawing()
			if out != nil && len(out) == 0 {
				fadeOut(func() { DrawArgumentsUI("Logout", cArgs, args, &sel) })
				break
			} else if len(out) != 0 {
				fadeOut(func() { DrawArgumentsUI("Logout", cArgs, args, &sel) })
				if args[0].(bool) {
					siteReddit.Logout()
					saveData.Reddit = ""
				}
				if args[1].(bool) {
					siteTwitter.Logout()
					saveData.Twitter = ""
				}
				if args[2].(bool) {
					saveData.Pixiv = ""
				}
				loginToSites()
				break
			}
		}
	}
	return SetUpSites()
}
