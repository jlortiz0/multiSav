package main

import (
	"os"
	"strings"

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
	ActionHandler(int) LoopStatus
}

// A website where images can be retrived from
// Might not actually be a site (cough cough LocalFSImageSite)
type ImageSite interface {
	Destroy()
	// Get a list of user-friendly listing types
	// Different listing types may get images in different ways
	// If this list has only one item, force the user to use that type
	GetListingTypeNames() []string
	// First argument is listing type, second and third arguments are query
	// Unneeded query arguments may be left blank
	// Returns a listing continuance object, which can be used to extend later, and a slice of urls
	GetListing(int, string, string) (interface{}, []string)
	// Return further objects from a listing using a continuance object
	// If the returned slice is empty or nil, the listing has concluded
	ExtendListing(interface{}) []string
}

type OfflineImageProducer struct {
	items []string
	fldr  string
}

func (OfflineImageProducer) Destroy() {}

func (OfflineImageProducer) IsLazy() bool { return false }

func (prod OfflineImageProducer) Len() int { return len(prod.items) }

func (OfflineImageProducer) ActionHandler(int) LoopStatus { return LOOP_CONT }

func (prod OfflineImageProducer) GetTitle() string { return prod.fldr }

func (prod OfflineImageProducer) BoundsCheck(i int) bool {
	return i >= 0 && i < len(prod.items)
}

func (prod OfflineImageProducer) Get(sel int, img **rl.Image, ffmpeg **ffmpegReader) string {
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
		prod.items = prod.items[:len(prod.items)-1]
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
