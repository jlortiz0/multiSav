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

const TEXT_SIZE = 18
const FRAME_RATE = 60

func loginHelper() ImageSite {
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
	red := redditapi.NewReddit("linux:org.jlortiz.test.GolangRedditAPI:v0.2.1 (by /u/jlortiz)", fields.Id, fields.Secret)
	err = red.Login(fields.Login, fields.Password)
	if err != nil {
		panic(fmt.Errorf("failed to log in: %s", err.Error()))
	}
	return &RedditSite{*red}
}

var font rl.Font

func main() {
	red := loginHelper()
	// sub, _ := red.Subreddit("cats")
	// iter, _ := sub.ListHot(300)
	// ls := NewLazySubmissionList(iter)
	// NewLazyImageMenu(ls)
	// red.Logout()
	rl.SetConfigFlags(rl.FlagVsyncHint)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(1024, 768, "rediSav Test Window")
	finder := sysfont.NewFinder(nil)
	font = rl.LoadFontEx(finder.Match("Ubuntu").Filename, TEXT_SIZE, nil, 250)
	// menu := NewChoiceMenu([]string{"Hello", "World", "test1", "Sort", "Trash", "Options", "New..."}, rl.Rectangle{X: 100, Y: 200, Height: 200, Width: 500})
	if _, err := os.Stat("jlortiz_TEST"); err == nil {
		os.Chdir("jlortiz_TEST")
	}
	os.Mkdir("Downloads", 0700)
	// menu, err := NewOfflineImageMenu("Sort", rl.Rectangle{Height: 768, Width: 1024})
	// if err != nil {
	// 	panic(err)
	// }
	producer := NewBufferedImageProducer(red, 0, []interface{}{"gifs"})
	menu := NewImageMenu(producer, rl.Rectangle{Height: 768, Width: 1024})
	rl.SetExitKey(0)
	rg.GuiSetFont(font)
	rg.GuiSetStyle(rg.LABEL, rg.TEXT_COLOR_NORMAL, 0xf5f5f5ff)
	rg.GuiSetStyle(rg.LABEL, rg.TEXT_ALIGNMENT, rg.TEXT_ALIGN_RIGHT)
Outer:
	for !rl.WindowShouldClose() {
		key := rl.GetKeyPressed()
		for key != 0 {
			if menu.HandleKey(key) != LOOP_CONT {
				break Outer
			}
			key = rl.GetKeyPressed()
		}
		if rl.IsWindowResized() {
			menu.SetTarget(rl.Rectangle{Width: float32(rl.GetScreenWidth()), Height: float32(rl.GetScreenHeight())})
		}
		menu.Prerender()
		rl.BeginDrawing()
		rl.ClearBackground(color.RGBA{R: 64, G: 64, B: 64})
		menu.Renderer()
		rl.EndDrawing()
	}
	rl.UnloadFont(font)
	rl.CloseWindow()
	red.Destroy()
}

func drawMessage(text string) rl.Texture2D {
	vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
	img := rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
	rl.ImageDrawTextEx(img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
	texture := rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)
	return texture
}
