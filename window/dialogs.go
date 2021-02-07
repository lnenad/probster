package window

import (
	"fmt"

	"github.com/alecthomas/chroma/styles"
	evbus "github.com/asaskevich/EventBus"
	gv "github.com/hashicorp/go-version"
	"github.com/lnenad/probster/storage"
	log "github.com/sirupsen/logrus"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type ConfirmationDialog struct {
	widget *gtk.MessageDialog
	yesno  *bool
}

type SettingsDialog struct {
	widget *gtk.Window
	Show   func()
}

type ErrorDialog struct {
	widget *gtk.MessageDialog
}

type NotificationDialog struct {
	widget *gtk.MessageDialog
}
type AboutDialog struct {
	widget *gtk.AboutDialog
}

func (cd *ConfirmationDialog) Confirm(question string, cb func(result bool)) {
	cd.widget.FormatSecondaryText(question)
	cd.widget.Run()
	cb(*cd.yesno)
}

func (ed *ErrorDialog) ShowError(message string) {
	ed.widget.FormatSecondaryText(message)
	ed.widget.Run()
}

func (nd *NotificationDialog) ShowNotification(message string) {
	nd.widget.FormatSecondaryText(message)
	nd.widget.Run()
}

func (nd *AboutDialog) ShowAbout() {
	nd.widget.Run()
}

func getErrorDialog(win *gtk.ApplicationWindow) *ErrorDialog {
	errorDiag := gtk.MessageDialogNew(
		win,
		gtk.DIALOG_MODAL,
		gtk.MESSAGE_ERROR,
		gtk.BUTTONS_CLOSE,
		"",
	)

	errorDiag.Connect("response", func() bool {
		errorDiag.Hide()
		return true
	})
	errorDiag.SetTitle("Error")

	return &ErrorDialog{errorDiag}
}

func getConfirmDialog(win *gtk.ApplicationWindow) *ConfirmationDialog {
	confirmDiag := gtk.MessageDialogNew(
		win,
		gtk.DIALOG_MODAL,
		gtk.MESSAGE_ERROR,
		gtk.BUTTONS_OK_CANCEL,
		"",
	)

	var yesno bool

	confirmDiag.Connect("response", func(msd *gtk.MessageDialog, result int) bool {
		confirmDiag.Hide()
		switch result {
		case -5:
			yesno = true
		case -6:
			yesno = false
		default:
			log.Fatal("Invalid result from confirm dialog")
		}
		return true
	})
	confirmDiag.SetTitle("Please confirm")

	return &ConfirmationDialog{confirmDiag, &yesno}
}

func getSettingsDialog(win *gtk.ApplicationWindow, settings *storage.Settings, bus evbus.Bus) *SettingsDialog {
	settingsDiag, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)

	settingsDiag.SetTitle("Preferences")
	settingsDiag.SetPosition(gtk.WIN_POS_MOUSE)
	settingsDiag.SetKeepAbove(true)
	settingsDiag.SetResizable(false)

	b, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 20)
	setMargins(b, 10, 30, 15, 30)

	l, _ := gtk.LabelNew("")
	l.SetMarkup(fmt.Sprintf("<span size='large' weight='bold'>Preferences</span>"))
	setMargins(l, 0, 0, 20, 0)

	updates, _ := gtk.CheckButtonNewWithLabel("Check for updates on startup")
	setMargins(updates, 0, 0, 40, 0)

	ltheme, _ := gtk.LabelNew("Syntax highlight theme")
	theme, _ := gtk.ComboBoxTextNew()
	setMargins(theme, 0, 0, 40, 0)

	bbox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 20)

	bs, _ := gtk.ButtonNewWithLabel("Save")
	bc, _ := gtk.ButtonNewWithLabel("Close")

	bbox.Add(bs)
	bbox.Add(bc)
	bbox.SetHAlign(gtk.ALIGN_CENTER)

	applyCurrentValues := func() {
		// Set current values
		if val, ok := (*settings)[storage.SettingCheckUpdates].(bool); ok {
			updates.SetActive(val)
		}

		var chosenTheme string
		if val, ok := (*settings)[storage.SettingTheme].(string); ok {
			chosenTheme = val
		}

		for k, v := range styles.Names() {
			theme.AppendText(v)
			if v == chosenTheme {
				theme.SetActive(k)
			}
		}
	}

	applyCurrentValues()

	b.Add(l)
	b.Add(updates)
	b.Add(ltheme)
	b.Add(theme)
	b.Add(bbox)

	settingsDiag.Add(b)

	bs.Connect("clicked", func() {
		newSettings := storage.Settings{
			storage.SettingTheme:        theme.GetActiveText(),
			storage.SettingCheckUpdates: updates.GetActive(),
		}
		*settings = newSettings
		bus.Publish("preferences:updated", newSettings)
		settingsDiag.Hide()
	})
	bc.Connect("clicked", func() {
		settingsDiag.Hide()
	})

	showFunc := func() {
		applyCurrentValues()
		settingsDiag.ShowAll()
	}

	return &SettingsDialog{settingsDiag, showFunc}
}

func getNotificationDialog(win *gtk.ApplicationWindow) *NotificationDialog {
	notificationDiag := gtk.MessageDialogNew(
		win,
		gtk.DIALOG_MODAL,
		gtk.MESSAGE_INFO,
		gtk.BUTTONS_CLOSE,
		"",
	)

	notificationDiag.Connect("response", func() bool {
		notificationDiag.Hide()
		return true
	})
	notificationDiag.SetTitle("Error")

	return &NotificationDialog{notificationDiag}
}

func getAboutDialog(currentVersion *gv.Version) *AboutDialog {
	ad, _ := gtk.AboutDialogNew()
	ad.SetPosition(gtk.WIN_POS_MOUSE)
	ad.SetProgramName("Probster - REST Easy")
	ld, _ := gdk.PixbufNewFromFile("icon.png")
	ad.SetLogo(ld)
	ad.SetAuthors([]string{"Nenad Lukic"})
	ad.SetVersion(currentVersion.String())
	ad.SetWebsite("https://probster.com")
	ad.SetComments("An alternative to all of the heavy, resource intensive Electron based apps. Simple to use.")

	ad.Connect("response", func() {
		ad.Hide()
	})
	return &AboutDialog{
		ad,
	}
}
