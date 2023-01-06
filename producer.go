package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

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
	Get(int, **rl.Image, *VideoReader) string
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
	// User-friendly name
	name string
	// Names of arguments (length gives number)
	args []ListingArgument
}

type ImageListing interface {
	GetInfo() (int, []interface{})
	GetPersistent() interface{}
}

type ErrorListing struct {
	error
}

func (ErrorListing) GetInfo() (int, []interface{}) {
	return -1, nil
}

func (err ErrorListing) GetPersistent() interface{} {
	return err.error
}

// A website where images can be retrived from
// Might not actually be a site (cough cough LocalFSImageSite)
type ImageSite interface {
	Resolver
	Destroy()
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
}

type Resolver interface {
	// Resolve a URL to another URL, which is hopefully one step closer to getting us an image
	// If the URL is a link to an image, this function should return RESOLVE_FINAL
	// To avoid unneeded calls, the caller should check the extension of the URL first
	ResolveURL(string) (string, ImageEntry)
	// Get a list of all domains that this site can resolve for
	GetResolvableDomains() []string
	// Might be best to have the resolver do this, since some sites may need different auths
	GetRequest(string) (*http.Response, error)
}

const RESOLVE_FINAL = "?"

type ImageEntryType int

const (
	IETYPE_REGULAR ImageEntryType = iota
	IETYPE_UGOIRA
	IETYPE_TEXT
	IETYPE_GALLERY
	// IETYPE_WEBPAGE
	// IETYPE_NONE
)

type ImageEntry interface {
	GetName() string
	GetURL() string
	GetText() string
	// If true is passed, no network requests should be made
	// Additionally, the returned slice need only be the correct length; it can otherwise be blank
	GetGalleryInfo(bool) []ImageEntry
	GetType() ImageEntryType
	GetPostURL() string
	GetInfo() string
	GetSaveName() string
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

func (w *WrapperImageEntry) GetSaveName() string {
	ind := strings.LastIndexByte(w.url, '/')
	if ind == -1 {
		return w.ImageEntry.GetSaveName()
	}
	s := w.url[ind+1:]
	if s == "" {
		return w.ImageEntry.GetSaveName()
	}
	ind = strings.IndexByte(s, '?')
	if ind != -1 {
		s = s[:ind]
	}
	return s
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

func (*TextImageEntry) GetPostURL() string { return "" }

func (*TextImageEntry) GetInfo() string { return "" }

func (*TextImageEntry) Combine(ImageEntry) {}

func (*TextImageEntry) GetSaveName() string { return "" }

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
			if getExtType(strings.ToLower(v.Name()[ind+1:])) != EXT_NONE {
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
	switch keycode {
	case rl.KeyC:
		if call != 0 {
			os.Remove(path.Join(prod.fldr, prod.items[sel]))
			return ARET_REMOVE
		} else {
			return ARET_MOVEDOWN | ARET_AGAIN | ARET_CLOSEFFMPEG
		}
	case rl.KeyX:
		if prod.fldr == saveData.Downloads {
			break
		}
		newName := prod.items[sel]
		if _, err := os.Stat(path.Join(saveData.Downloads, newName)); err == nil {
			x := -1
			dLoc := strings.IndexByte(newName, '.')
			before := newName
			var after string
			if dLoc != -1 {
				before = newName[:dLoc]
				after = newName[dLoc+1:]
			}
			for ; err == nil; _, err = os.Stat(fmt.Sprintf("%s%s_%d.%s", saveData.Downloads, before, x, after)) {
				x++
			}
			newName = fmt.Sprintf("%s_%d.%s", before, x, after)
		}
		os.Rename(path.Join(prod.fldr, prod.items[sel]), path.Join(saveData.Downloads, newName))
	case rl.KeyV:
		openFile(path.Join(prod.fldr, prod.items[sel]))
	case rl.KeyH:
		highlightFile(path.Join(prod.fldr, prod.items[sel]))
	}
	return ARET_NOTHING
}

func (prod *OfflineImageProducer) GetTitle() string {
	return "multiSav - Offline - " + prod.fldr
}

func (prod *OfflineImageProducer) BoundsCheck(i int) bool {
	return i >= 0 && i < len(prod.items)
}

func (prod *OfflineImageProducer) Get(sel int, img **rl.Image, ffmpeg *VideoReader) string {
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
	_, err := os.Stat(path.Join(prod.fldr, prod.items[sel]))
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
		_, err = os.Stat(path.Join(prod.fldr, prod.items[sel]))
	}
	ind := strings.LastIndexByte(prod.items[sel], '.')
	if getExtType(strings.ToLower(prod.items[sel][ind+1:])) == EXT_VIDEO {
		var err error
		*ffmpeg, err = NewStreamy(path.Join(prod.fldr, prod.items[sel]))
		if err != nil {
			panic(err)
		}
		w, h := (*ffmpeg).GetDimensions()
		*img = rl.GenImageColor(int(w), int(h), rl.Blank)
	} else {
		*img = rl.LoadImage(path.Join(prod.fldr, prod.items[sel]))
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
	stat, err := os.Stat(path.Join(prod.fldr, prod.items[sel]))
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
