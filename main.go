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

func loginHelper() *HybridImgurRedditSite {
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
		ImgurID  string
	}
	err = json.Unmarshal(data[:n], &fields)
	if err != nil {
		panic(fmt.Errorf("failed to decode login data: %s", err.Error()))
	}
	red := redditapi.NewReddit("linux:org.jlortiz.rediSav:v0.3.2 (by /u/jlortiz)", fields.Id, fields.Secret)
	err = red.Login(fields.Login, fields.Password)
	if err != nil {
		panic(fmt.Errorf("failed to log in: %s", err.Error()))
	}
	return &HybridImgurRedditSite{RedditSite{*red}, *NewImgurSite(fields.ImgurID)}
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
	rl.SetExitKey(0)
	rg.GuiSetFont(font)
	rg.GuiSetStyle(rg.LABEL, rg.TEXT_COLOR_NORMAL, 0xf5f5f5ff)
	rg.GuiSetStyle(rg.LABEL, rg.TEXT_ALIGNMENT, rg.TEXT_ALIGN_RIGHT)
	// menu := NewChoiceMenu([]string{"Hello", "World", "test1", "Sort", "Trash", "Options", "New..."}, rl.Rectangle{X: 100, Y: 200, Height: 200, Width: 500})
	if _, err := os.Stat("jlortiz_TEST"); err == nil {
		os.Chdir("jlortiz_TEST")
	}
	os.Mkdir("Downloads", 0700)
	// menu, err := NewOfflineImageMenu("Sort", rl.Rectangle{Height: 768, Width: 1024})
	// if err != nil {
	// 	panic(err)
	// }
	// producer := NewRedditProducer(red, 2, nil)
	// producer := NewHybridImgurRedditProducer(red, 0, []interface{}{"pics"})
	producer := SetUpProducer(red, func(i ImageSite, k int, a []interface{}) ImageProducer {
		return NewHybridImgurRedditProducer(i.(*HybridImgurRedditSite), k, a)
	})
	if producer == nil {
		red.Destroy()
		rl.CloseWindow()
		return
	}
	menu := NewImageMenu(producer, rl.Rectangle{Height: 768, Width: 1024})
	stdEventLoop(menu, func() rl.Rectangle {
		return rl.Rectangle{Width: float32(rl.GetScreenWidth()), Height: float32(rl.GetScreenHeight())}
	})
	menu.Destroy()
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

func displayMessage(text string, menu Menu) {
	msg := drawMessage(text)
	x := (int32(rl.GetScreenWidth()) - msg.Width) / 2
	y := (int32(rl.GetScreenHeight()) - msg.Height) / 2
	for !rl.WindowShouldClose() {
		if rl.GetKeyPressed() != 0 {
			break
		}
		if rl.IsWindowResized() {
			x = (int32(rl.GetScreenWidth()) - msg.Width) / 2
			y = (int32(rl.GetScreenHeight()) - msg.Height) / 2
			menu.SetTarget(rl.Rectangle{Width: float32(rl.GetScreenWidth()), Height: float32(rl.GetScreenHeight())})
		}
		menu.Prerender()
		rl.BeginDrawing()
		rl.ClearBackground(color.RGBA{R: 64, G: 64, B: 64})
		menu.Renderer()
		rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), color.RGBA{A: 128})
		rl.DrawTexture(msg, x, y, rl.White)
		rl.EndDrawing()
	}
	rl.UnloadTexture(msg)
}

func stdEventLoop(menu Menu, targetGen func() rl.Rectangle) LoopStatus {
	for !rl.WindowShouldClose() {
		key := rl.GetKeyPressed()
		for key != 0 {
			ret := menu.HandleKey(key)
			if ret != LOOP_CONT {
				return ret
			}
			key = rl.GetKeyPressed()
		}
		if rl.IsWindowResized() {
			menu.SetTarget(targetGen())
		}
		ret := menu.Prerender()
		if ret != LOOP_CONT {
			return ret
		}
		rl.BeginDrawing()
		rl.ClearBackground(color.RGBA{R: 64, G: 64, B: 64})
		menu.Renderer()
		rl.EndDrawing()
	}
	return LOOP_QUIT
}
