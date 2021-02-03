# Probster

API probing tool

## Setup

Install gtk

https://github.com/gotk3/gotk3/wiki/Installing-on-Windows
https://github.com/gotk3/gotk3/wiki/Installing-on-macOS

## Compiling (windows)

Extract and copy deps to build (exec from mingw64)

`ldd probster.exe | grep '\/mingw.*\.dll' -o | xargs -I{} cp "{}" build`

Copy `dbus.exe` from `/msys64/mingw64/bin` folder to `build`

Copy `lib/gdk-pixbuf-2.0` from `/msys64/mingw64/` folder to `build/lib`

Copy `share/icons` from `/msys64/mingw64/` to `build/share`

Generate syso file to embed icon to executable using `https://github.com/akavel/rsrc`

`rsrc -ico build/icon.ico`

Compile with `CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -i -ldflags -H=windowsgui`

## Important

GTK is not thread safe so this is helpful

https://github.com/conformal/gotk3/blob/master/gtk/examples/goroutines/goroutines.go

## Maintainers 

* [lnenad](https://github.com/lnenad)