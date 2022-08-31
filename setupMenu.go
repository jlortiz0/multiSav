package main

import (
	"net/url"

	rl "github.com/gen2brain/raylib-go/raylib"
	rg "jlortiz.org/redisav/raygui-go"
)

const ARGUI_SPACING = CHOICEMENU_SPACE_BETWEEN_ITEM * 2

func DrawArgumentsUI(target rl.Rectangle, name string, args []ListingArgument, out []interface{}, flags []bool) []interface{} {
	if out[0] == nil {
		for i, v := range args {
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
	vec := rl.MeasureTextEx(font, name, TEXT_SIZE, 0)
	vec2 := rl.Vector2{X: target.X + ((target.Width - vec.X) / 2), Y: target.Y + target.Height/2 - (TEXT_SIZE+ARGUI_SPACING)*float32(len(args)+1)/2}
	rl.DrawRectangle(int32(target.Width/4+target.X), int32(vec2.Y)-5, int32(target.Width)/2, (TEXT_SIZE+ARGUI_SPACING)*int32(len(args)+1)+10, rl.RayWhite)
	vec2.Y += (CHOICEMENU_SPACE_BETWEEN_ITEM + TEXT_SIZE) / 2
	rl.DrawTextEx(font, name, vec2, TEXT_SIZE, 0, rl.Black)
	for i, v := range args {
		vec2.Y += CHOICEMENU_SPACE_BETWEEN_ITEM + TEXT_SIZE
		name := v.name + ":"
		vec = rl.MeasureTextEx(font, name, TEXT_SIZE, 0)
		vec2.X = target.Width/2 + target.X - vec.X - 5
		rl.DrawTextEx(font, name, vec2, TEXT_SIZE, 0, rl.Black)
		switch v.kind {
		case LARGTYPE_BOOL:
			out[i] = rg.GuiCheckBox(rl.Rectangle{X: target.Width/2 + target.X + 5, Y: vec2.Y - 3, Width: TEXT_SIZE + 5, Height: TEXT_SIZE + 5}, "", out[i].(bool))
		case LARGTYPE_URL:
			// Character limit seems too small for a url
			// Further investigation needed
			fallthrough
		case LARGTYPE_STRING:
			var ret bool
			ret, out[i] = rg.GuiTextBox(rl.Rectangle{X: target.Width/2 + target.X + 5, Y: vec2.Y - 5, Width: target.Width/4 - 10, Height: TEXT_SIZE + 10}, out[i].(string), TEXT_SIZE, flags[i])
			if ret {
				flags[i] = !flags[i]
			}
		case LARGTYPE_INT:
			temp := out[i].(int)
			if rg.GuiValueBox(rl.Rectangle{X: target.Width/2 + target.X + 5, Y: vec2.Y - 5, Width: target.Width/8 - 10, Height: TEXT_SIZE + 10}, "", &temp, -999, 999, flags[i]) {
				flags[i] = !flags[i]
			}
			out[i] = temp
		}
	}
	vec2.Y += (CHOICEMENU_SPACE_BETWEEN_ITEM + TEXT_SIZE) * 1.5
	if rg.GuiButton(rl.Rectangle{X: target.Width/4 + target.X + 10, Y: vec2.Y, Height: TEXT_SIZE + 5, Width: target.Width/4 - 20}, "Cancel") {
		return []interface{}{}
	}
	if rg.GuiButton(rl.Rectangle{X: target.Width/2 + target.X + 10, Y: vec2.Y, Height: TEXT_SIZE + 5, Width: target.Width/4 - 20}, "Confirm") {
		for i := range out {
			switch args[i].kind {
			case LARGTYPE_STRING:
				if out[i].(string) == "" {
					// fmt.Println("Empty string")
					return nil
				}
			case LARGTYPE_INT:
				// Do I need to do any verification here? 0 is probably valid...
			// case LARGTYPE_TIME:
			// 	if out[i].(time.Time).IsZero() {
			// 		fmt.Println("OUTATIME")
			// 		return nil
			// 	}
			case LARGTYPE_URL:
				_, err := url.ParseRequestURI(out[i].(string))
				if err != nil {
					// fmt.Println(err)
					return nil
				}
			}
		}
		return out
	}
	return nil
}
