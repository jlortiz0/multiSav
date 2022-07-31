package redditapi

type Redditor struct {
	ID                                                              string
	Is_employee, Is_friend, Is_mod, Is_gold, Is_suspended, Verified bool
	Created_utc                                                     uint64
	Name                                                            string
	Icon_img                                                        string
	Subreddit                                                       string
	Total_karma                                                     int
	reddit                                                          *Reddit
}
