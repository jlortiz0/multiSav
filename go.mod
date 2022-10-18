module jlortiz.org/multisav

go 1.16

require (
	github.com/adrg/sysfont v0.1.2
	github.com/g8rswimmer/go-twitter/v2 v2.1.2
	github.com/gen2brain/raylib-go/raylib v0.0.0-20220702153720-2ba84634ed1e
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
	github.com/sqweek/dialog v0.0.0-20220809060634-e981b270ebbf
	github.com/stretchr/testify v1.8.0 // indirect
	golang.org/x/oauth2 v0.0.0-20221014153046-6fdb5e3db783
	jlortiz.org/multisav/pixivapi v0.0.0-00010101000000-000000000000
	jlortiz.org/multisav/raygui-go v0.0.0
	jlortiz.org/multisav/redditapi v0.0.0
)

replace jlortiz.org/multisav/redditapi => ./redditapi

replace jlortiz.org/multisav/raygui-go => ./raygui-go

replace jlortiz.org/multisav/pixivapi => ./pixivapi
