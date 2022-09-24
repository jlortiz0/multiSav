package pixivapi

type multisize struct {
	Square_medium string
	Medium        string
	Large         string
}

type User struct {
	ID                 int
	Name               string
	Account            string
	Profile_image_urls multisize
	Is_followed        bool
	Comment            string
	client             *Client
}

type Account struct {
	User
	Mail_address string
	Is_premium   bool
	// TODO: What do these mean?
	X_restrict         int
	Is_mail_authorized bool
}
