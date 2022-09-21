module jlortiz.org/redisav

go 1.16

require (
	github.com/adrg/sysfont v0.1.2
	github.com/dghubble/go-twitter v0.0.0-20220816163853-8a0df96f1e6d
	github.com/gen2brain/raylib-go/raylib v0.0.0-20220702153720-2ba84634ed1e
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
	github.com/sqweek/dialog v0.0.0-20220809060634-e981b270ebbf
	golang.org/x/oauth2 v0.0.0-20220909003341-f21342109be1
	jlortiz.org/redisav/raygui-go v0.0.0
	jlortiz.org/redisav/redditapi v0.0.0
)

replace jlortiz.org/redisav/redditapi => ./redditapi

replace jlortiz.org/redisav/raygui-go => ./raygui-go
