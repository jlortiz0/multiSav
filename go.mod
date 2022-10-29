module github.com/jlortiz0/multisav

go 1.17

require (
	github.com/adrg/sysfont v0.1.2
	github.com/g8rswimmer/go-twitter/v2 v2.1.2
	github.com/gen2brain/raylib-go/raylib v0.0.0-20220702153720-2ba84634ed1e
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
	github.com/sqweek/dialog v0.0.0-20220809060634-e981b270ebbf
	golang.org/x/oauth2 v0.1.0
	github.com/jlortiz0/multisav/pixivapi v0.0.0-00010101000000-000000000000
	github.com/jlortiz0/multisav/raygui-go v0.0.0
	github.com/jlortiz0/multisav/redditapi v0.0.0
	github.com/jlortiz0/multisav/streamy v0.0.0
)

require (
	github.com/DaRealFreak/cloudflare-bp-go v1.0.4 // indirect
	github.com/EDDYCJY/fake-useragent v0.2.0 // indirect
	github.com/PuerkitoBio/goquery v1.7.1 // indirect
	github.com/TheTitanrain/w32 v0.0.0-20180517000239-4f5cfb03fabf // indirect
	github.com/adrg/strutil v0.2.2 // indirect
	github.com/adrg/xdg v0.3.0 // indirect
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/stretchr/testify v1.8.0 // indirect
	golang.org/x/net v0.1.0 // indirect
	golang.org/x/sys v0.1.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)

replace github.com/jlortiz0/multisav/redditapi => ./redditapi

replace github.com/jlortiz0/multisav/raygui-go => ./raygui-go

replace github.com/jlortiz0/multisav/pixivapi => ./pixivapi

replace github.com/jlortiz0/multisav/streamy => ./streamy

replace github.com/g8rswimmer/go-twitter/v2 => github.com/jlortiz0/go-twitter/v2 v2.1.3-0.20221018050935-6eff13d54906
