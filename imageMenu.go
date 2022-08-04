package main

import (
	"os"
	"sort"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type OfflineImageMenu struct {
	Selected     int
	itemList     []string
	img          *rl.Image
	texture      rl.Texture2D
	prevMoveDir  bool
	state        imageMenuState
	fldr         string
	loadingFrame int
	// ffmpeg *void
}

type imageMenuState int

const (
	IMSTATE_NORMAL imageMenuState = iota
	IMSTATE_SHOULDLOAD
	IMSTATE_LOADING
	IMSTATE_EMPTY
	IMSTATE_NOSUPPORT
	IMSTATE_ERROR
	IMSTATE_SHOULDEXIT
)

func NewOfflineImageMenu(fldr string) (*OfflineImageMenu, error) {
	f, err := os.Open(fldr)
	if err != nil {
		return nil, err
	}
	entries, err := f.ReadDir(0)
	if err != nil {
		return nil, err
	}
	ls := make([]string, 0, len(entries))
	for _, v := range entries {
		if !v.IsDir() {
			ind := strings.LastIndexByte(v.Name(), '.')
			if ind == -1 {
				continue
			}
			switch strings.ToLower(v.Name()[ind+1:]) {
			// case "mp4":
			// 	fallthrough
			// case "webm":
			// 	fallthrough
			// case "gif":
			// 	fallthrough
			// case "mov":
			// 	fallthrough
			case "bmp":
				fallthrough
			case "jpg":
				fallthrough
			case "png":
				fallthrough
			case "jpeg":
				ls = append(ls, v.Name())
			}
		}
	}
	sort.Strings(ls)
	menu := new(OfflineImageMenu)
	menu.fldr = fldr
	menu.itemList = ls
	menu.state = IMSTATE_SHOULDLOAD
	if len(ls) == 0 {
		if len(entries) == 0 {
			menu.state = IMSTATE_NOSUPPORT
		} else {
			menu.state = IMSTATE_EMPTY
		}
	}
	return menu, nil
}

func (menu *OfflineImageMenu) loadImage() LoopStatus {
	if len(menu.itemList) == 0 {
		return LOOP_EXIT
	}
	menu.state = IMSTATE_LOADING
	go func() {
		_, err := os.Stat(menu.fldr + string(os.PathSeparator) + menu.itemList[menu.Selected])
		for err != nil {
			if len(menu.itemList) == 1 {
				menu.state = IMSTATE_SHOULDEXIT
				menu.itemList = nil
				return
			}
			if menu.Selected == len(menu.itemList)-1 {
				menu.Selected--
			} else {
				copy(menu.itemList[menu.Selected:], menu.itemList[menu.Selected+1:])
				if menu.prevMoveDir && menu.Selected > 0 {
					menu.Selected--
				}
			}
			menu.itemList = menu.itemList[:len(menu.itemList)-1]
			_, err = os.Stat(menu.fldr + string(os.PathSeparator) + menu.itemList[menu.Selected])
		}
		menu.img = rl.LoadImage(menu.fldr + string(os.PathSeparator) + menu.itemList[menu.Selected])
		menu.state = IMSTATE_NORMAL
	}()
	return LOOP_CONT
}

func (menu *OfflineImageMenu) HandleKey(keycode int32) LoopStatus {
	switch keycode {
	case rl.KeyLeft:
		if menu.state == IMSTATE_NORMAL && menu.Selected > 0 {
			menu.Selected--
			menu.state = IMSTATE_SHOULDLOAD
		}
	case rl.KeyRight:
		if menu.state == IMSTATE_NORMAL && menu.Selected < len(menu.itemList) {
			menu.Selected++
			menu.state = IMSTATE_SHOULDLOAD
		}
	case rl.KeyEscape:
		return LOOP_BACK
	}
	return LOOP_CONT
}

func (menu *OfflineImageMenu) Renderer(target rl.Rectangle) LoopStatus {
	switch menu.state {
	case IMSTATE_SHOULDEXIT:
		return LOOP_EXIT
	case IMSTATE_SHOULDLOAD:
		menu.loadImage()
		return menu.Renderer(target)
	case IMSTATE_EMPTY:
		if menu.img.Width == 0 {
			menu.texture = drawMessage("(this page intentionally left blank)")
		}
		fallthrough
	case IMSTATE_NOSUPPORT:
		if menu.img.Width == 0 {
			menu.texture = drawMessage("(no supported images)")
		}
		fallthrough
	case IMSTATE_LOADING:
		rl.DrawTexture(menu.texture, int32(target.X), int32(target.Y), rl.White)
		x := int32(target.Width)/2 - 50 + int32(target.X)
		y := int32(target.Height)/2 - 50 + int32(target.Y)
		rl.DrawRectangle(x, y, 100, 100, rl.RayWhite)
		if menu.loadingFrame < 10 {
			rl.DrawRectangle(x+int32(menu.loadingFrame)*5, y, 50, 50, rl.Black)
			rl.DrawRectangle(x+50-int32(menu.loadingFrame)*5, y+50, 50, 50, rl.Black)
		} else if menu.loadingFrame < 16 {
			rl.DrawRectangle(x+50, y, 50, 50, rl.Black)
			rl.DrawRectangle(x, y+50, 50, 50, rl.Black)
		} else if menu.loadingFrame < 26 {
			rl.DrawRectangle(x+50, y-80+int32(menu.loadingFrame)*5, 50, 50, rl.Black)
			rl.DrawRectangle(x, y+130-int32(menu.loadingFrame)*5, 50, 50, rl.Black)
		} else {
			rl.DrawRectangle(x, y, 50, 50, rl.Black)
			rl.DrawRectangle(x+50, y+50, 50, 50, rl.Black)
		}
		menu.loadingFrame++
		menu.loadingFrame %= 32
	case IMSTATE_NORMAL:
		if menu.img != nil {
			if menu.texture.ID > 0 {
				rl.UnloadTexture(menu.texture)
			}
			menu.texture = rl.LoadTextureFromImage(menu.img)
			rl.UnloadImage(menu.img)
			menu.img = nil
		}
		rl.DrawTexture(menu.texture, int32(target.X), int32(target.Y), rl.White)
	}
	return LOOP_CONT
}

func (menu *OfflineImageMenu) Cleanup() {
	if menu.texture.ID <= 0 {
		rl.UnloadTexture(menu.texture)
	}
	if menu.img != nil {
		rl.UnloadImage(menu.img)
	}
}
