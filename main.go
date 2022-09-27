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
var siteReddit RedditSite
var siteTwitter TwitterSite
var sitePixiv PixivSite

const TEXT_SIZE = 18
const FRAME_RATE = 60

func loginHelper() RedditSite {
	data := make([]byte, 1024)
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
		Id            string
		Secret        string
		Login         string
		Password      string
		ImgurId       string
		TwitterId     string
		TwitterSecret string
		TwitterBearer string
		PixivToken    string
	}
	err = json.Unmarshal(data[:n], &fields)
	if err != nil {
		panic(fmt.Errorf("failed to decode login data: %s", err.Error()))
	}
	red := redditapi.NewReddit("linux:org.jlortiz.rediSav:v0.5.1 (by /u/jlortiz)", fields.Id, fields.Secret)
	err = red.Login(fields.Login, fields.Password)
	if err != nil {
		panic(fmt.Errorf("failed to log in: %s", err.Error()))
	}
	r := RedditSite{red}
	resolveMap = make(map[string]Resolver)
	for _, x := range r.GetResolvableDomains() {
		resolveMap[x] = r
	}
	img := NewImgurResolver(fields.ImgurId)
	for _, x := range img.GetResolvableDomains() {
		resolveMap[x] = img
	}
	siteTwitter = NewTwitterSite(fields.TwitterBearer)
	for _, x := range siteTwitter.GetResolvableDomains() {
		resolveMap[x] = siteTwitter
	}
	sitePixiv, err = NewPixivSite(fields.PixivToken)
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
	return r
}

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
	Reddit struct {
		Id       string
		Secret   string
		Login    string
		Password string
	}
	Imgur   string
	Twitter struct {
		Id     string
		Secret string
	}
	Listings []SavedListing
}

func loadSaveData() error {
	f, err := os.Open("rediSav.json")
	if err != nil {
		if os.IsNotExist(err) {
			saveData.Listings = []SavedListing{{"Downloads", SITE_LOCAL, 0, []interface{}{"Downloads"}, nil}}
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
	siteReddit = loginHelper()
	rl.SetConfigFlags(rl.FlagVsyncHint)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(1024, 768, "rediSav")
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
	fadeOut(func() { rl.ClearBackground(rl.Black) })
MainLoop:
	for {
		names := make([]string, len(saveData.Listings), len(saveData.Listings)+4)
		for i, v := range saveData.Listings {
			names[i] = v.Name
		}
		names = append(names, "Edit Listings", "Options", "Quit")
		cm := NewChoiceMenu(names)
		if stdEventLoop(cm) != LOOP_EXIT {
			break
		}
		sel := cm.Selected
		cm.Destroy()
		if sel >= len(saveData.Listings) {
			switch sel - len(saveData.Listings) {
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
		} else {
			data := saveData.Listings[sel]
			var producer ImageProducer
			switch data.Site {
			case SITE_LOCAL:
				producer = NewOfflineImageProducer(data.Args[0].(string))
			case SITE_REDDIT:
				producer = NewRedditProducer(siteReddit, data.Kind, data.Args, data.Persistent)
			case SITE_TWITTER:
				producer = NewBufferedImageProducer(siteTwitter, data.Kind, data.Args, data.Persistent)
			case SITE_PIXIV:
				producer = NewBufferedImageProducer(sitePixiv, data.Kind, data.Args, data.Persistent)
			}
			menu := NewImageMenu(producer)
			if stdEventLoop(menu) == LOOP_QUIT {
				break MainLoop
			}
			listing := producer.GetListing()
			if listing != nil {
				saveData.Listings[sel].Persistent = listing.GetPersistent()
			}
			menu.Destroy()
			rl.SetWindowTitle("rediSav")
		}
	}
	fadeIn(func() { rl.ClearBackground(rl.Black) })
	rl.UnloadFont(font)
	rl.CloseWindow()
	siteReddit.Destroy()
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
