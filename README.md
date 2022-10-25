# ~~rediSav~~ multiSav, the multiple source image saver

[Grabber](https://github.com/Bionus/imgbrd-grabber) is better than this. I just wanted to save my Reddit saved. And my Twitter saved. And maybe my Pixiv saved.

You know, it would be nice if I combined this and ImageSort, ignoring that they're on different engines. I suppose separating offline and online does make programming things a bit easier. I mean, I could interface everything, but that seems like a poor idea when the semantics of offline and online radically differ.

## Building

In addition to what is needed from `go mod download`, the following dependencies are required to build:
 - (raylib)[] v???
 - (raygui)[] v???
 - libavcodec, libavformat, libswscale, libavutil v???
   - For Linux, install the dev packages from your package manager of choice
   - For Windows, download the windows-shared build from (BtbN)[https://github.com/BtbN/FFmpeg-Builds/releases] and install the libraries and headers.

Due to Twitter being annoying about distribution of secrets (and they gave me a secret even though I specfically said this was a native application), and also to avoid hitting the API caps (people probably won't use this, but bots sure will), client tokens for Imgur, Reddit, and Twitter are not included in this package. You will have to define the following constants yourself to get the program to compile:
 - ImgurID
 - TwitterBearer (app-only token)
 - TwitterID
 - TwitterSecret
 - RedditID
 - RedditSecret (can be blank)
I suggest placing these definitions in `token.go`, because I did that so it's already in the gitignore.

## Legal notes

While the source code of this program is licensed under the zlib License, this program links with FFmpeg's libraries, which may be licensed under the GPL depending on compile options. As such, the compiled version may be licensed under the GPL. If you wish to avoid this, consider using an LGPL-compliant version of the libav libraries.

### To do list
 - Write instructions
 - The various things I put in the comments
  - Find a bunch of weird edge cases
