package window

import (
	gv "github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type ConfirmationDialog struct {
	widget *gtk.MessageDialog
	yesno  *bool
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
