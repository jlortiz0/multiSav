package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// An image producer produces images given a listing
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
	// Returns the listing for saving purposes
	// The returned listing should not be modified
	GetListing() ImageListing
}

type ListingArgumentType int

const (
	LARGTYPE_STRING ListingArgumentType = iota
	LARGTYPE_INT
	// LARGTYPE_TIME
	LARGTYPE_BOOL
	// LARGTYPE_FLOAT
	// LARGTYPE_CHAR
	LARGTYPE_URL
	// To display information
	// options must be a 1-length slice where the first option is a string with the info
	LARGTYPE_LABEL
)

type ListingArgument struct {
	name string
	kind ListingArgumentType
	// If not nil, possible options to select from a drop-down
	options []interface{}
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

type ImageListing interface {
	GetInfo() (int, []interface{})
	GetPersistent() interface{}
}

type ErrorListing struct {
	err error
}

func (*ErrorListing) GetInfo() (int, []interface{}) {
	return -1, nil
}

func (err *ErrorListing) GetPersistent() interface{} {
	return err.err
}

// A website where images can be retrived from
// Might not actually be a site (cough cough LocalFSImageSite)
type ImageSite interface {
	Resolver
	// Get a list of user-friendly listing types
	// Different listing types may get images in different ways
	// If this list has only one item, force the user to use that type
	GetListingInfo() []ListingInfo
	// First argument is listing type, second argument is slice of parameters, third argument is any persistent data
	// Unneeded parameters in args should be zeroed. Unused persistent data should be nil
	// Returns a listing continuance object, which can be used to extend later, and a slice of urls
	// In case of an error, slice will be nil and listing will be an ErrorListing
	GetListing(int, []interface{}, interface{}) (ImageListing, []ImageEntry)
	// Return further objects from a listing using a continuance object
	// If the returned slice is empty or nil, the listing has concluded
	ExtendListing(ImageListing) []ImageEntry
	GetRequest(string) (*http.Response, error)
}

type Resolver interface {
	Destroy()
	// Resolve a URL to another URL, which is hopefully one step closer to getting us an image
	// If the URL is a link to an image, this function should return RESOLVE_FINAL
	// To avoid unneeded calls, the caller should check the extension of the URL first
	ResolveURL(string) (string, ImageEntry)
	// Get a list of all domains that this site can resolve for
	GetResolvableDomains() []string
}

const RESOLVE_FINAL = "finalfinalfinal"

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
	// If true is passed, no network requests should be made
	// Additionally, the returned slice need only be the correct length; it can otherwise be blank
	GetGalleryInfo(bool) []ImageEntry
	GetType() ImageEntryType
	GetDimensions() (int, int)
	GetPostURL() string
	GetInfo() string
	// Modify the current ImageEntry inplace to contain some fields from this new one
	// What gets copied is up to the implementation
	Combine(ImageEntry)
}

type WrapperImageEntry struct {
	ImageEntry
	url string
}

func (w *WrapperImageEntry) GetURL() string {
	return w.url
}

type TextImageEntry struct {
	msg  string
	name string
}

func (t *TextImageEntry) GetText() string {
	return t.msg
}

func (t *TextImageEntry) GetName() string { return t.name }

func (*TextImageEntry) GetURL() string { return "" }

func (*TextImageEntry) GetGalleryInfo(bool) []ImageEntry { return nil }

func (*TextImageEntry) GetType() ImageEntryType { return IETYPE_TEXT }

func (*TextImageEntry) GetDimensions() (int, int) { return -1, -1 }

func (*TextImageEntry) GetPostURL() string { return "" }

func (*TextImageEntry) GetInfo() string { return "" }

func (*TextImageEntry) Combine(ImageEntry) {}

type ActionRet int

const (
	ARET_NOTHING  ActionRet = 0
	ARET_MOVEDOWN ActionRet = 1 << iota
	ARET_MOVEUP
	ARET_AGAIN
	ARET_REMOVE
	ARET_CLOSEFFMPEG
	ARET_QUIT
	ARET_FADEOUT
	ARET_FADEIN
)

type OfflineImageProducer struct {
	items []string
	fldr  string
	empty uint8
}

func NewOfflineImageProducer(fldr string) *OfflineImageProducer {
	f, err := os.Open(fldr)
	if err != nil {
		return nil
	}
	entries, err := f.ReadDir(0)
	if err != nil {
		return nil
	}
	ls := make([]string, 0, len(entries))
	for _, v := range entries {
		if !v.IsDir() {
			ind := strings.LastIndexByte(v.Name(), '.')
			if ind == -1 {
				continue
			}
			switch strings.ToLower(v.Name()[ind+1:]) {
			case "mp4":
				fallthrough
			case "webm":
				fallthrough
			case "gif":
				fallthrough
			case "mov":
				fallthrough
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
	var b uint8
	if len(ls) == 0 {
		b = 1
		if len(entries) != 0 {
			b |= 2
		}
	}
	return &OfflineImageProducer{ls, fldr, b}
	// if len(ls) == 0 {
	// 	menu.state = IMSTATE_ERROR
	// 	if len(entries) != 0 {
	// 		menu.texture = drawMessage("No supported files found.")
	// 	} else {
	// 		menu.texture = drawMessage("Empty.")
	// 	}
	// 	menu.cam.Target = rl.Vector2{Y: float32(menu.texture.Height) / 2, X: float32(menu.texture.Width) / 2}
	// }
}

func (*OfflineImageProducer) Destroy() {}

func (*OfflineImageProducer) IsLazy() bool { return false }

func (prod *OfflineImageProducer) Len() int {
	if prod.empty != 0 {
		return 1
	}
	return len(prod.items)
}

func (prod *OfflineImageProducer) ActionHandler(keycode int32, sel int, call int) ActionRet {
	if prod.empty != 0 {
		return ARET_NOTHING
	}
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
		if prod.empty != 0 {
			text := "No supported images"
			if prod.empty&2 == 0 {
				text = "Folder is empty"
			}
			*img = drawMessage(text)
			return "\\/errNo place to go, nothing to do"
		}
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
			*img = drawMessage("Failed to load image?")
			return "\\/err" + prod.items[sel]
		}
	}
	return prod.items[sel]
}

func (prod *OfflineImageProducer) GetInfo(sel int) string {
	if !prod.BoundsCheck(sel) {
		return ""
	}
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

func (*OfflineImageProducer) GetListing() ImageListing { return nil }
