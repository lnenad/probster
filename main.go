package main

import (
	"log"

	history "github.com/lnenad/probster/storage"
	"github.com/lnenad/probster/window"

	"os"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/xujiajun/nutsdb"
)

const appID = "com.mockadillo.probster"

func main() {
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal("Could not create application:", err)
	}

	opt := nutsdb.DefaultOptions
	opt.Dir = "/data/file"
	db, err := nutsdb.Open(opt)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	h := history.Setup(db)

	application.Connect("activate", func() {
		window.BuildWindow(application, &h)

		aNew := glib.SimpleActionNew("new", nil)
		aNew.Connect("activate", func() {
			window.BuildWindow(application, &h).ShowAll()
		})
		application.AddAction(aNew)

		aQuit := glib.SimpleActionNew("quit", nil)
		aQuit.Connect("activate", func() {
			application.Quit()
		})
		application.AddAction(aQuit)
	})

	os.Exit(application.Run(os.Args))
}
