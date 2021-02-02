# Probester

API probing tool

## Setup

Install gtk

https://github.com/gotk3/gotk3/wiki/Installing-on-Windows
https://github.com/gotk3/gotk3/wiki/Installing-on-macOS

## Compiling (windows)

Extract and copy deps to build (exec from mingw64)

`ldd probester.exe | grep '\/mingw.*\.dll' -o | xargs -I{} cp "{}" build`

Copy `dbus.exe` from `/msys64/mingw64/bin` folder to `build`

Copy `lib/gdk-pixbuf-2.0` from `/msys64/mingw64/` folder to `build/lib`

Copy `share/icons` from `/msys64/mingw64/` to `build/share`

Compile with `CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -i -ldflags -H=windowsgui`