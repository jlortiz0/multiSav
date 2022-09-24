package pixivapi

import "time"

type Illustration struct {
	ID    int
	Title string
	// Type enum{string}
	Image_urls multisize
	Caption    string
	// What does this mean?
	Restrict int
	User     *User
	Tags     []struct {
		Name, Translated_name string
	}
	Create_date time.Time
	Page_count  int
	Width       int
	Height      int
	// What do these mean?
	Sanity_level int
	X_restrict   int
	Series       struct {
		ID    string
		Title string
	}
	Meta_single_page struct {
		Original_image_url string
	}
	Total_view      int
	Total_bookmarks int
	Is_bookmarked   bool
	Is_muted        bool
	Total_comments  int
	client          *Client
}
