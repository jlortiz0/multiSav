package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// An image producer produces images given a listing
// TODO: figure out if I should make producers or consumers buffer image objects
// Given the design of this at this time, maybe producer should
type ImageProducer interface {
	// Get the length of this producer
	// This may incrase if the producer is lazy and not everything is loaded yet
	Len() int
	// Return if there may be more images not yet listed
	IsLazy() bool
	// Get the image at the given index
	// One of the two pointers will be filled depending on the image type
	// A filename to display to the user will be returned
	Get(int, **rl.Image, **ffmpegReader) string
	// Ensure that the given index is within bounds
	// May load additional images to check this
	BoundsCheck(int) bool
	Destroy()
	// Something to use as the title of the current window
	GetTitle() string
	// KeyHandler for a producer
	// Some keys are reserved by the menu and should not be overriden
	ActionHandler(int32, int, int) ActionRet
	// Get some textual info about the image at the given index
	// This info will be displayed to the user in a Message
	// If the returned string is empty, the Message will not be opened
	GetInfo(int) string
}

type ListingArgumentType int

const (
	LARGTYPE_STRING ListingArgumentType = iota
	LARGTYPE_INT
	LARGTYPE_TIME
	LARGTYPE_BOOL
	LARGTYPE_FLOAT
	LARGTYPE_CHAR
	LARGTYPE_URL
)

type ListingArgument struct {
	name     string
	kind     ListingArgumentType
	optional bool
}

type ListingInfo struct {
	// ID, should be identical to index in GetListingInfo slice
	// id int
	// User-friendly name
	name string
	// Names of arguments (length gives number)
	args []ListingArgument
	// If this listing needs to save data in between sessions to work properly
	// This may be used to implement a listing which dislays all new content since the last time it was viewed
	persistent bool
}

// A website where images can be retrived from
// Might not actually be a site (cough cough LocalFSImageSite)
type ImageSite interface {
	Destroy()
	// Get a list of user-friendly listing types
	// Different listing types may get images in different ways
	// If this list has only one item, force the user to use that type
	GetListingInfo() []ListingInfo
	// First argument is listing type, second argument is slice of parameters
	// Unneeded parameters may be left blank or ommitted if towards the end
	// Returns a listing continuance object, which can be used to extend later, and a slice of urls
	// In case of an error, slice will be nil and interface will be an error
	GetListing(int, []interface{}) (interface{}, []ImageEntry)
	// Return further objects from a listing using a continuance object
	// If the returned slice is empty or nil, the listing has concluded
	ExtendListing(interface{}) []ImageEntry
}

type ImageEntryType int

const (
	IETYPE_REGULAR ImageEntryType = iota
	IETYPE_ANIMATED
	IETYPE_TEXT
	IETYPE_GALLERY
	IETYPE_WEBPAGE
	IETYPE_NONE
)

type ImageEntry interface {
	GetName() string
	GetURL() string
	GetText() string
	GetGalleryInfo() []ImageEntry
	GetType() ImageEntryType
	GetDimensions() (int, int)
	GetPostURL() string
	GetInfo() string
}

type ActionRet int

const (
	ARET_NOTHING  ActionRet = 0
	ARET_MOVEDOWN ActionRet = 1 << iota
	ARET_MOVEUP
	ARET_AGAIN
	ARET_REMOVE
	ARET_CLOSEFFMPEG
)

type OfflineImageProducer struct {
	items []string
	fldr  string
}

func (*OfflineImageProducer) Destroy() {}

func (*OfflineImageProducer) IsLazy() bool { return false }

func (prod *OfflineImageProducer) Len() int { return len(prod.items) }

func (prod *OfflineImageProducer) ActionHandler(keycode int32, sel int, call int) ActionRet {
	if keycode == rl.KeyX || keycode == rl.KeyC {
		if call != 0 {
			targetFldr := "Sort" + string(os.PathSeparator)
			if keycode == rl.KeyC {
				targetFldr = "Trash" + string(os.PathSeparator)
			}
			newName := prod.items[sel]
			if _, err := os.Stat(targetFldr + newName); err == nil {
				x := -1
				dLoc := strings.IndexByte(newName, '.')
				before := newName
				var after string
				if dLoc != -1 {
					before = newName[:dLoc]
					after = newName[dLoc+1:]
				}
				for ; err == nil; _, err = os.Stat(fmt.Sprintf("%s%s_%d.%s", targetFldr, before, x, after)) {
					x++
				}
				newName = fmt.Sprintf("%s_%d.%s", before, x, after)
			}
			os.Rename(prod.fldr+string(os.PathSeparator)+prod.items[sel], targetFldr+newName)
			return ARET_REMOVE
		} else if keycode == rl.KeyC {
			if strings.HasSuffix(prod.fldr, "Trash") {
				return ARET_NOTHING
			}
			return ARET_MOVEDOWN | ARET_AGAIN | ARET_CLOSEFFMPEG
		} else if strings.HasSuffix(prod.fldr, "Sort") {
			return ARET_NOTHING
		} else {
			return ARET_MOVEUP | ARET_AGAIN | ARET_CLOSEFFMPEG
		}
	} else if keycode == rl.KeyV && os.PathSeparator == '\\' {
		exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", prod.fldr+string(os.PathSeparator)+prod.items[sel]).Run()
	} else if keycode == rl.KeyH && os.PathSeparator == '\\' {
		cwd, _ := os.Getwd()
		cmd := exec.Command("explorer", "/select,", fmt.Sprintf("\"%s%c%s%c%s\"", cwd, os.PathSeparator, prod.fldr, os.PathSeparator, prod.items[sel]))
		cwd = fmt.Sprintf("explorer /select, %s", cmd.Args[2])
		cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: cwd}
		cmd.Run()
	}
	return ARET_NOTHING
}

func (prod *OfflineImageProducer) GetTitle() string {
	return "rediSav - Offline - " + prod.fldr
}

func (prod *OfflineImageProducer) BoundsCheck(i int) bool {
	return i >= 0 && i < len(prod.items)
}

func (prod *OfflineImageProducer) Get(sel int, img **rl.Image, ffmpeg **ffmpegReader) string {
	if !prod.BoundsCheck(sel) {
		return ""
	}
	_, err := os.Stat(prod.fldr + string(os.PathSeparator) + prod.items[sel])
	for err != nil {
		if len(prod.items) == 1 {
			prod.items = nil
			return ""
		}
		if sel != prod.Len()-1 {
			copy(prod.items[sel:], prod.items[sel+1:])
			// if menu.prevMoveDir && menu.Selected > 0 {
			// 	menu.Selected--
			// }
		}
		prod.items = prod.items[:prod.Len()-1]
		if sel == prod.Len() {
			return ""
		}
		_, err = os.Stat(prod.fldr + string(os.PathSeparator) + prod.items[sel])
	}
	ind := strings.LastIndexByte(prod.items[sel], '.')
	switch strings.ToLower(prod.items[sel][ind+1:]) {
	case "mp4":
		fallthrough
	case "webm":
		fallthrough
	case "gif":
		fallthrough
	case "mov":
		*ffmpeg = NewFfmpegReader(prod.fldr + string(os.PathSeparator) + prod.items[sel])
		*img = rl.GenImageColor(int((*ffmpeg).w), int((*ffmpeg).h), rl.Blank)
	default:
		*img = rl.LoadImage(prod.fldr + string(os.PathSeparator) + prod.items[sel])
		if (*img).Height == 0 {
			text := "Failed to load image?"
			vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
			*img = rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
			rl.ImageDrawTextEx(*img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
			return "\\/err" + prod.items[sel]
		}
	}
	return prod.items[sel]
}

func (prod *OfflineImageProducer) GetInfo(sel int) string {
	stat, err := os.Stat(prod.fldr + string(os.PathSeparator) + prod.items[sel])
	if err == nil {
		size := stat.Size()
		if size > 1024*1024 {
			return fmt.Sprintf("%s\n%.1f MiB", prod.items[sel], float32(size)/(1024*1024))
		} else if size > 1024 {
			return fmt.Sprintf("%s\n%.1f KiB", prod.items[sel], float32(size)/1024)
		} else {
			return fmt.Sprintf("%s\n%d B", prod.items[sel], size)
		}
	} else {
		return fmt.Sprintf("%s\n%s", prod.items[sel], err.Error())
	}
}
