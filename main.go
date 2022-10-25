package main

import (
	"encoding/json"
	"image/color"
	"io"
	"os"

	"github.com/adrg/sysfont"
	rl "github.com/gen2brain/raylib-go/raylib"
	rg "jlortiz.org/multisav/raygui-go"
)

var resolveMap map[string]Resolver
var siteReddit RedditSite
var siteTwitter TwitterSite
var sitePixiv PixivSite

const TEXT_SIZE = 18
const FRAME_RATE = 60

var font rl.Font

const (
	SITE_LOCAL = iota
	SITE_REDDIT
	// SITE_IMGUR
	SITE_TWITTER
	SITE_PIXIV
)

type SavedListing struct {
	Name       string
	Site       int
	Kind       int
	Args       []interface{} `json:",omitempty"`
	Persistent interface{}   `json:",omitempty"`
}

var saveData struct {
	Reddit    string
	Twitter   string
	Pixiv     string
	Downloads string
	Settings  struct {
		SaveOnX bool
	}
	Listings []SavedListing
}

func loadSaveData() error {
	f, err := os.Open("multiSav.json")
	if err != nil {
		if os.IsNotExist(err) {
			saveData.Downloads = "Downloads"
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

func loginToSites() {
	if _, err := os.Stat(saveData.Downloads); err != nil {
		os.MkdirAll(saveData.Downloads, 0600)
	}
	siteReddit = NewRedditSite(saveData.Reddit)
	resolveMap = make(map[string]Resolver)
	for _, x := range siteReddit.GetResolvableDomains() {
		resolveMap[x] = siteReddit
	}
	img := NewImgurResolver(ImgurID)
	for _, x := range img.GetResolvableDomains() {
		resolveMap[x] = img
	}
	siteTwitter = NewTwitterSite(saveData.Twitter)
	for _, x := range siteTwitter.GetResolvableDomains() {
		resolveMap[x] = siteTwitter
	}
	var err error
	sitePixiv, err = NewPixivSite(saveData.Pixiv)
	if err != nil {
		panic(err)
	}
	for _, x := range sitePixiv.GetResolvableDomains() {
		resolveMap[x] = sitePixiv
	}
	var b byte
	for _, x := range BlockingResolver(b).GetResolvableDomains() {
		resolveMap[x] = BlockingResolver(b)
	}
	for _, x := range StripQueryResolver(b).GetResolvableDomains() {
		resolveMap[x] = StripQueryResolver(b)
	}
}

func saveSaveData() error {
	data, err := json.Marshal(saveData)
	if err != nil {
		return err
	}
	f, err := os.OpenFile("multiSav.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
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
	rl.SetConfigFlags(rl.FlagVsyncHint)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(1024, 768, "multiSav")
	finder := sysfont.NewFinder(nil)
	font = rl.LoadFontEx(finder.Match("Ubuntu").Filename, TEXT_SIZE, nil, 0)
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
	loginToSites()
	fadeOut(func() { rl.ClearBackground(rl.Black) })
MainLoop:
	for {
		names := make([]string, len(saveData.Listings)+1, len(saveData.Listings)+4)
		names[0] = "Downloads"
		for i, v := range saveData.Listings {
			names[i+1] = v.Name
		}
		names = append(names, "Edit Listings", "Options", "Quit")
		cm := NewChoiceMenu(names)
		if stdEventLoop(cm) != LOOP_EXIT {
			break
		}
		sel := cm.Selected
		cm.Destroy()
		if sel >= len(saveData.Listings)+1 {
			switch sel - len(saveData.Listings) - 1 {
			case 0:
				if EditListings() {
					break MainLoop
				}
			case 1:
				if SetUpSites() {
					break MainLoop
				}
			case 2:
				break MainLoop
			}
		} else if sel == 0 {
			menu := NewImageMenu(func() <-chan ImageProducer {
				ch := make(chan ImageProducer)
				go func() {
					ch <- NewOfflineImageProducer(saveData.Downloads)
				}()
				return ch
			})
			if stdEventLoop(menu) == LOOP_QUIT {
				break MainLoop
			}
			menu.Destroy()
			rl.SetWindowTitle("multiSav")
		} else {
			data := saveData.Listings[sel-1]
			menu := NewImageMenu(func() <-chan ImageProducer {
				ch := make(chan ImageProducer)
				go func() {
					var producer ImageProducer
					switch data.Site {
					case SITE_LOCAL:
						producer = NewOfflineImageProducer(data.Args[0].(string))
					case SITE_REDDIT:
						producer = NewRedditProducer(siteReddit, data.Kind, data.Args, data.Persistent)
					case SITE_TWITTER:
						producer = NewTwitterProducer(siteTwitter, data.Kind, data.Args, data.Persistent)
					case SITE_PIXIV:
						producer = NewPixivProducer(sitePixiv, data.Kind, data.Args, data.Persistent)
					}
					ch <- producer
				}()
				return ch
			})
			if stdEventLoop(menu) == LOOP_QUIT {
				break MainLoop
			}
			if menu.Producer != nil {
				listing := menu.Producer.GetListing()
				if listing != nil {
					saveData.Listings[sel-1].Persistent = listing.GetPersistent()
				}
			}
			menu.Destroy()
			rl.SetWindowTitle("multiSav")
		}
	}
	fadeIn(func() { rl.ClearBackground(rl.Black) })
	rl.UnloadFont(font)
	rl.CloseWindow()
	siteReddit.Destroy()
	siteTwitter.Destroy()
	sitePixiv.Destroy()
	err = saveSaveData()
	if err != nil {
		panic(err)
	}
}

func drawMessage(text string) *rl.Image {
	vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
	img := rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
	rl.ImageDrawTextEx(img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
	return img
}

func messageOverlay(text string, menu Menu) {
	msgI := drawMessage(text)
	msg := rl.LoadTextureFromImage(msgI)
	rl.UnloadImage(msgI)
	x := (int32(rl.GetScreenWidth()) - msg.Width) / 2
	y := (int32(rl.GetScreenHeight()) - msg.Height) / 2
	for !rl.WindowShouldClose() {
		if rl.GetKeyPressed() != 0 {
			break
		}
		if rl.IsWindowResized() {
			x = (int32(rl.GetScreenWidth()) - msg.Width) / 2
			y = (int32(rl.GetScreenHeight()) - msg.Height) / 2
			menu.RecalcTarget()
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

func stdEventLoop(menu Menu) LoopStatus {
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
			menu.RecalcTarget()
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
