package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"

	"github.com/adrg/sysfont"
	rl "github.com/gen2brain/raylib-go/raylib"
	rg "jlortiz.org/redisav/raygui-go"
	"jlortiz.org/redisav/redditapi"
)

func loginHelper() *redditapi.Reddit {
	data := make([]byte, 256)
	f, err := os.Open("redditapi/login.json")
	if err != nil {
		panic(fmt.Errorf("failed to open login data file: %s", err.Error()))
	}
	n, err := f.Read(data)
	if err != nil {
		panic(fmt.Errorf("failed to read login data: %s", err.Error()))
	}
	f.Close()
	var fields struct {
		Id       string
		Secret   string
		Login    string
		Password string
	}
	err = json.Unmarshal(data[:n], &fields)
	if err != nil {
		panic(fmt.Errorf("failed to decode login data: %s", err.Error()))
	}
	red := redditapi.NewReddit("linux:org.jlortiz.test.GolangRedditAPI:v0.0.1 (by /u/jlortiz)", fields.Id, fields.Secret)
	err = red.Login(fields.Login, fields.Password)
	if err != nil {
		panic(fmt.Errorf("failed to log in: %s", err.Error()))
	}
	return red
}

var font rl.Font

func main() {
	// red := loginHelper()
	// sub, _ := red.Subreddit("cats")
	// iter, _ := sub.ListHot(300)
	// ls := NewLazySubmissionList(iter)
	// NewLazyImageMenu(ls)
	// red.Logout()
	rl.SetConfigFlags(rl.FlagVsyncHint)
	rl.InitWindow(1024, 768, "rediSav Test Window")
	finder := sysfont.NewFinder(nil)
	font = rl.LoadFontEx(finder.Match("Ubuntu").Filename, 18, nil, 250)
	// menu := NewChoiceMenu([]string{"Hello", "World", "test1", "Sort", "Trash", "Options", "New..."})
	menu, err := NewOfflineImageMenu("jlortiz_TEST/Sort")
	if err != nil {
		panic(err)
	}
	// w := menu.SuggestWidth()
	rl.SetExitKey(0)
	rg.GuiSetFont(font)
Outer:
	for !rl.WindowShouldClose() {
		key := rl.GetKeyPressed()
		for key != 0 {
			if menu.HandleKey(key) != LOOP_CONT {
				break Outer
			}
			key = rl.GetKeyPressed()
		}
		rl.BeginDrawing()
		rl.ClearBackground(color.RGBA{R: 64, G: 64, B: 64})
		s := "Logged in as: somebody"
		vec := rl.MeasureTextEx(font, s, 18, 0)
		menu.Renderer(rl.Rectangle{Height: 768 - vec.Y - 10, Width: 1024})
		rl.DrawRectangle(0, 768-int32(vec.Y)-10, 1024, int32(vec.Y)+10, rl.Black)
		rl.DrawTextEx(font, s, rl.Vector2{Y: 768 - vec.Y - 5, X: 1024/2 - vec.X/2}, vec.Y, 0, rl.RayWhite)
		rl.EndDrawing()
	}
	rl.UnloadFont(font)
	rl.CloseWindow()
}

func drawMessage(text string) rl.Texture2D {
	vec := rl.MeasureTextEx(font, text, 18, 0)
	img := rl.GenImageColor(int(vec.X)+16, int(vec.Y)+12, rl.RayWhite)
	rl.ImageDrawTextEx(img, rl.Vector2{X: 8, Y: 5}, font, text, 18, 0, rl.Black)
	texture := rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)
	return texture
}
