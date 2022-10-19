package main

import (
	"fmt"
	"image/color"
	"net/url"
	"os"
	"path"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/sqweek/dialog"
	rg "jlortiz.org/multisav/raygui-go"
)

const ARGUI_SPACING = CHOICEMENU_SPACE_BETWEEN_ITEM * 1.5

func DrawArgumentsUI(name string, args []ListingArgument, out []interface{}, flags []bool) []interface{} {
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
		if len(v.options) != 0 {
			if v.kind == LARGTYPE_LABEL {
				rl.DrawTextEx(font, args[i].options[0].(string), rl.Vector2{X: target.X/2 + 5, Y: vec2.Y}, TEXT_SIZE, 0, rl.Black)
				continue
			}
			name := strings.Builder{}
			name.WriteString(fmt.Sprint(v.options[0]))
			for _, n := range v.options[1:] {
				name.WriteByte(';')
				name.WriteString(fmt.Sprint(n))
			}
			out[i] = rg.GuiComboBox(rl.Rectangle{X: target.X/2 + 5, Y: vec2.Y - 5, Width: target.X/4 - 10, Height: TEXT_SIZE + 10}, name.String(), out[i].(int))
		} else {
			switch v.kind {
			case LARGTYPE_BOOL:
				out[i] = rg.GuiCheckBox(rl.Rectangle{X: target.X/2 + 5, Y: vec2.Y - 3, Width: TEXT_SIZE + 5, Height: TEXT_SIZE + 5}, "", out[i].(bool))
			case LARGTYPE_URL:
				// Character limit seems too small for a url
				// Further investigation needed
				fallthrough
			case LARGTYPE_STRING:
				var ret bool
				ret, out[i] = rg.GuiTextBox(rl.Rectangle{X: target.X/2 + 5, Y: vec2.Y - 5, Width: target.X/4 - 10, Height: TEXT_SIZE + 10}, out[i].(string), TEXT_SIZE, flags[i])
				if ret {
					flags[i] = !flags[i]
				}
			case LARGTYPE_INT:
				temp := out[i].(int)
				if rg.GuiValueBox(rl.Rectangle{X: target.X/2 + 5, Y: vec2.Y - 5, Width: target.X/8 - 10, Height: TEXT_SIZE + 10}, "", &temp, -999, 999, flags[i]) {
					flags[i] = !flags[i]
				}
				out[i] = temp
			}
		}
	}
	vec2.Y += (ARGUI_SPACING + TEXT_SIZE)
	if rg.GuiButton(rl.Rectangle{X: target.X/4 + 10, Y: vec2.Y, Height: TEXT_SIZE + 5, Width: target.X/4 - 20}, "Cancel") {
		return []interface{}{}
	}
	if rg.GuiButton(rl.Rectangle{X: target.X/2 + 10, Y: vec2.Y, Height: TEXT_SIZE + 5, Width: target.X/4 - 20}, "Confirm") {
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
	flags := make([]bool, len(choices[kind].args)+1)
	cArgs := make([]ListingArgument, 1, len(choices[kind].args)+1)
	cArgs[0].name = "Name"
	cArgs[0].kind = LARGTYPE_STRING
	cArgs = append(cArgs, choices[kind].args...)
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(color.RGBA{R: 64, G: 64, B: 64})
		out := DrawArgumentsUI(choices[kind].name, cArgs, args, flags)
		rl.EndDrawing()
		if out != nil && len(out) == 0 {
			goto cm
		} else if len(out) != 0 {
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
		if !path.IsAbs(dir) {
			tmp, _ := os.Getwd()
			dir = path.Join(tmp, dir)
		}
		s, err := db.SetStartDir(dir).Title("Select downloads folder").Browse()
		if err == nil {
			if s[len(s)-1] == os.PathSeparator {
				s = s[:len(s)-1]
			}
			saveData.Downloads = s
		}
	case 1:
		args := []interface{}{saveData.Settings.SaveOnX}
		flags := make([]bool, 1)
		cArgs := []ListingArgument{
			{
				name: "Always download on X",
				kind: LARGTYPE_BOOL,
			},
		}
		for !rl.WindowShouldClose() {
			rl.BeginDrawing()
			rl.ClearBackground(color.RGBA{R: 64, G: 64, B: 64})
			out := DrawArgumentsUI("Options", cArgs, args, flags)
			rl.EndDrawing()
			if out != nil && len(out) == 0 {
				break
			} else if len(out) != 0 {
				saveData.Settings.SaveOnX = args[0].(bool)
				break
			}
		}
	case 2:
		args := []interface{}{false, false, false}
		flags := make([]bool, 3)
		cArgs := make([]ListingArgument, 3)
		if saveData.Reddit.Refresh == "" {
			cArgs[0] = ListingArgument{
				name: "Not logged in to Reddit",
				kind: LARGTYPE_LABEL,
			}
		} else {
			cArgs[0] = ListingArgument{
				name: "Reddit",
				kind: LARGTYPE_BOOL,
			}
		}
		if saveData.Twitter.Refresh == "" {
			cArgs[1] = ListingArgument{
				name: "Not logged in to Twitter",
				kind: LARGTYPE_LABEL,
			}
		} else {
			cArgs[1] = ListingArgument{
				name: "Twitter",
				kind: LARGTYPE_BOOL,
			}
		}
		if saveData.Pixiv.Refresh == "" {
			cArgs[2] = ListingArgument{
				name: "Not logged in to Pixiv",
				kind: LARGTYPE_LABEL,
			}
		} else {
			cArgs[2] = ListingArgument{
				name: "Pixiv",
				kind: LARGTYPE_BOOL,
			}
		}
		for !rl.WindowShouldClose() {
			rl.BeginDrawing()
			rl.ClearBackground(color.RGBA{R: 64, G: 64, B: 64})
			out := DrawArgumentsUI("Logout", cArgs, args, flags)
			rl.EndDrawing()
			if out != nil && len(out) == 0 {
				break
			} else if len(out) != 0 {
				if args[0].(bool) {
					saveData.Reddit = struct {
						Token   string
						Refresh string
					}{}
				}
				if args[1].(bool) {
					saveData.Twitter = struct {
						Token   string
						Refresh string
					}{}
				}
				if args[2].(bool) {
					saveData.Pixiv = struct {
						Token   string
						Refresh string
					}{}
				}
				loginToSites()
				break
			}
		}
	}
	return SetUpSites()
}
