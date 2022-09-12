package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"os"

	"github.com/adrg/sysfont"
	rl "github.com/gen2brain/raylib-go/raylib"
	rg "jlortiz.org/redisav/raygui-go"
	"jlortiz.org/redisav/redditapi"
)

var resolveMap map[string]Resolver

const TEXT_SIZE = 18
const FRAME_RATE = 60

func loginHelper() *RedditSite {
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
	r := &RedditSite{*red}
	resolveMap = make(map[string]Resolver)
	for _, x := range r.GetResolvableDomains() {
		resolveMap[x] = r
	}
	img := NewImgurResolver(fields.ImgurID)
	for _, x := range img.GetResolvableDomains() {
		resolveMap[x] = img
	}
	return r
}

var font rl.Font

type SavedListing struct {
	Name string
	// Site int
	Kind       int
	Args       []interface{}
	Persistent interface{} `json:",omitempty"`
}

var saveData struct {
	Listings []SavedListing
}

func loadSaveData() error {
	f, err := os.Open("rediSav.json")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &saveData)
}

func saveSaveData() error {
	data, err := json.Marshal(saveData)
	if err != nil {
		return err
	}
	f, err := os.OpenFile("rediSav.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	for len(data) > 0 {
		n, err := f.Write(data)
		if err != nil {
			f.Close()
			return err
		}
		data = data[n:]
	}
	return f.Close()
}

func main() {
	red := loginHelper()
	rl.SetConfigFlags(rl.FlagVsyncHint)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(1024, 768, "rediSav")
	finder := sysfont.NewFinder(nil)
	font = rl.LoadFontEx(finder.Match("Ubuntu").Filename, TEXT_SIZE, nil, 250)
	rl.SetExitKey(0)
	rg.GuiSetFont(font)
	rg.GuiSetStyle(rg.LABEL, rg.TEXT_COLOR_NORMAL, 0xf5f5f5ff)
	rg.GuiSetStyle(rg.LABEL, rg.TEXT_ALIGNMENT, rg.TEXT_ALIGN_RIGHT)
	if _, err := os.Stat("jlortiz_TEST"); err == nil {
		os.Chdir("jlortiz_TEST")
	}
	os.Mkdir("Downloads", 0700)
	err := loadSaveData()
	if err != nil {
		panic(err)
	}
	fadeOut(func() {})
	// TODO: Allow the user to see (and maybe edit) the parameters of an existing listing
MainLoop:
	for {
		names := make([]string, len(saveData.Listings), len(saveData.Listings)+4)
		for i, v := range saveData.Listings {
			names[i] = v.Name
		}
		names = append(names, "Reset Lisiting", "Delete Listing", "New Listing", "Quit")
		cm := NewChoiceMenu(names, GetCenteredCoiceMenuRect(len(saveData.Listings)+4, float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())))
		if stdEventLoop(cm, func() rl.Rectangle {
			return GetCenteredCoiceMenuRect(len(saveData.Listings)+4, float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight()))
		}) != LOOP_EXIT {
			break
		}
		sel := cm.Selected
		cm.Destroy()
		if sel >= len(saveData.Listings) {
			switch sel - len(saveData.Listings) {
			case 0:
				names[len(names)-4] = "Back"
				cm := NewChoiceMenu(names[:len(names)-3], GetCenteredCoiceMenuRect(len(names)-3, float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())))
				stdEventLoop(cm, func() rl.Rectangle {
					return GetCenteredCoiceMenuRect(len(saveData.Listings)+4, float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight()))
				})
				sel := cm.Selected
				cm.Destroy()
				if sel != len(names)-4 {
					saveData.Listings[sel].Persistent = nil
				}
			case 1:
				names[len(names)-4] = "Back"
				cm := NewChoiceMenu(names[:len(names)-3], GetCenteredCoiceMenuRect(len(names)-3, float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())))
				stdEventLoop(cm, func() rl.Rectangle {
					return GetCenteredCoiceMenuRect(len(saveData.Listings)+4, float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight()))
				})
				sel := cm.Selected
				cm.Destroy()
				if sel != len(names)-4 {
					copy(saveData.Listings[sel:], saveData.Listings[sel+1:])
					saveData.Listings = saveData.Listings[:len(saveData.Listings)-1]
				}
			case 2:
				kind, args := SetUpListing(red)
				if kind != -1 {
					saveData.Listings = append(saveData.Listings, SavedListing{Kind: kind, Args: args[1:], Name: args[0].(string)})
				}
			case 3:
				break MainLoop
			}
		} else {
			data := saveData.Listings[sel]
			producer := NewRedditProducer(red, data.Kind, data.Args, data.Persistent)
			menu := NewImageMenu(producer, rl.Rectangle{Width: float32(rl.GetScreenWidth()), Height: float32(rl.GetScreenHeight())})
			stdEventLoop(menu, func() rl.Rectangle {
				return rl.Rectangle{Width: float32(rl.GetScreenWidth()), Height: float32(rl.GetScreenHeight())}
			})
			listing := producer.GetListing()
			if listing != nil {
				saveData.Listings[sel].Persistent = listing.GetPersistent()
			}
			menu.Destroy()
			rl.SetWindowTitle("rediSav")
		}
	}
	fadeIn(func() {})
	rl.UnloadFont(font)
	rl.CloseWindow()
	red.Destroy()
	err = saveSaveData()
	if err != nil {
		panic(err)
	}
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
	fadeIn(menu.Renderer)
	for !rl.WindowShouldClose() {
		key := rl.GetKeyPressed()
		for key != 0 {
			ret := menu.HandleKey(key)
			if ret != LOOP_CONT {
				fadeOut(menu.Renderer)
				return ret
			}
			key = rl.GetKeyPressed()
		}
		if rl.IsWindowResized() {
			menu.SetTarget(targetGen())
		}
		ret := menu.Prerender()
		if ret != LOOP_CONT {
			fadeOut(menu.Renderer)
			return ret
		}
		rl.BeginDrawing()
		rl.ClearBackground(color.RGBA{R: 64, G: 64, B: 64})
		menu.Renderer()
		rl.EndDrawing()
	}
	return LOOP_QUIT
}
