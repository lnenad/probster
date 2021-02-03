package main

import (
	"log"

	"github.com/lnenad/probster/storage"
	"github.com/lnenad/probster/window"

	"os"

	evbus "github.com/asaskevich/EventBus"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/xujiajun/nutsdb"
)

const appID = "com.mockadillo.probster"
const dataFile = "/data/file"
const logFile = "/data/log"

func main() {
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal("Could not create application:", err)
	}

	// fmt.Printf("OS ARGS: %#v", os.Args)
	// if len(os.Args) >= 2 && os.Args[1] == "debug" {
	// 	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// 	if err != nil {
	// 		log.Fatalf("error opening file: %v", err)
	// 	}
	// 	defer f.Close()

	// 	log.SetOutput(f)
	// }
	opt := nutsdb.DefaultOptions
	opt.Dir = dataFile
	db, err := nutsdb.Open(opt)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	h := storage.SetupHistory(db)
	bus := evbus.New()

	application.Connect("activate", func() {
		window.BuildWindow(application, &h, bus)

		aNew := glib.SimpleActionNew("new", nil)
		aNew.Connect("activate", func() {
			window.BuildWindow(application, &h, bus).ShowAll()
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
