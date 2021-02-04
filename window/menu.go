package window

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	evbus "github.com/asaskevich/EventBus"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func registerMenu(win *gtk.ApplicationWindow, bus evbus.Bus, confirmDiag *ConfirmationDialog) {
	// Create a header bar
	header, err := gtk.HeaderBarNew()
	if err != nil {
		log.Fatal("Could not create header bar:", err)
	}
	header.SetShowCloseButton(true)
	header.SetTitle("Probster - REST Easy")
	header.SetSubtitle("API testing tool")

	// Create a new menu button
	mbtn, err := gtk.MenuButtonNew()
	if err != nil {
		log.Fatal("Could not create menu button:", err)
	}

	// Set up the menu model for the button
	menu := glib.MenuNew()
	if menu == nil {
		log.Fatal("Could not create menu (nil)")
	}
	// Actions with the prefix 'app' reference actions on the application
	// Actions with the prefix 'win' reference actions on the current window (specific to ApplicationWindow)
	// Other prefixes can be added to widgets via InsertActionGroup
	menu.Append("New Request", "win.new-request")
	menu.Append("Clear history", "win.clear-history")
	menu.Append("Quit", "app.quit")

	// Create the action "win.close"
	aClearHistory := glib.SimpleActionNew("clear-history", nil)
	aClearHistory.Connect("activate", func() {
		confirmDiag.Confirm(fmt.Sprintf("This will delete your entire requests history.\nAre you really sure that you want to proceed?"), func(yes bool) {
			if yes {
				bus.Publish("history:clear")
			}
		})
	})
	win.AddAction(aClearHistory)

	// Create the action "win.new-request"
	aNewRequest := glib.SimpleActionNew("new-request", nil)
	aNewRequest.Connect("activate", func() {
		bus.Publish("request:new")
	})
	win.AddAction(aNewRequest)

	mbtn.SetMenuModel(&menu.MenuModel)

	// add the menu button to the header
	header.PackStart(mbtn)
	win.SetTitlebar(header)

}
