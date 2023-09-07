# ~~rediSav~~ multiSav, the multiple source image saver

[Grabber](https://github.com/Bionus/imgbrd-grabber) is better than this. I just wanted to save my Reddit saved. And maybe my Pixiv saved.

You know, it would be nice if I combined this and ImageSort, ignoring that they're on different engines. I suppose separating offline and online does make programming things a bit easier. I mean, I could interface everything, but that seems like a poor idea when the semantics of offline and online radically differ.

## Building

In addition to what is needed from `go mod download`, the following dependencies are required to build:

- [raylib](https://github.com/raysan5/raylib) v4.2
- [raygui](https://github.com/raysan5/raygui) v3.0
- libavcodec, libavformat, libswscale, libavutil v4.2.7 (this corresponds to FFmpeg version/package version, not library version)
  - For Linux, install the dev packages from your package manager of choice
  - For Windows, download the windows-shared build from [BtbN](https://github.com/BtbN/FFmpeg-Builds/releases) and install the libraries and headers.

To avoid hitting the API caps (people probably won't use this, but bots sure will), client tokens for Imgur and Reddit are not included in this package. You will have to define the following constants yourself to get the program to compile:

- ImgurID
- RedditID
- RedditSecret (can be blank)

I suggest placing these definitions in `token.go`, because I did that so it's already in the gitignore.

## Instructions

Images are retrived through listings. You can create a listing by going to Edit Listings > New. Select a site, then select the kind of listing. Give the new listing a name and adjust other options. At time of writing, up to 20 characters can be inputted into an option.

To browse images, use the left and right arrow keys or the targets on the left and right of the screen to switch images. Press Z or click the title at the bottom to view information about the current entry. Use the up and down arrow keys to zoom in and out. When zoomed in, use WASD to move the image around.

In an online listing, press V to open the current image or link in a browser, or H to open the post URL in a browser. Press C to remove the current entry (on some sites this may hide the entry from appearing again). Press X to download the image to the downloads folder. On some sites, pressing X may instead save or bookmark the image. Hold shift while pressing X or change the settings to override this. Press enter to view galleries if prompted.

Some features of sites may require that you sign in. To do this, use superAuthorizer. Run the program and select the site to sign in to. A new page will open in your browser prompting you to give access. If your firewall asks if you want to allow superAuthorizer through the firewall, say no. To sign out later, use the Logout menu in multisav's options menu.

## Legal notes

While the source code of this program is licensed under the zlib License, this program links with FFmpeg's libraries, which may be licensed under the GPL depending on compile options. As such, the compiled version may be licensed under the GPL. If you wish to avoid this, consider using an LGPL-compliant version of the libav libraries.

## To do list

- Write better instructions
- Find a bunch of weird edge cases
- Ability for a resolver to pass an error
- Get to a point that I can call this "released"
- Look into a better renderer (why did I pick raylib, of all things?)
