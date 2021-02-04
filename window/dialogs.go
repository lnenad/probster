package window

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/gotk3/gotk3/gtk"
)

type ConfirmationDialog struct {
	widget *gtk.MessageDialog
	yesno  *bool
}

type ErrorDialog struct {
	widget *gtk.MessageDialog
}

func (cd *ConfirmationDialog) Confirm(question string, cb func(result bool)) {
	cd.widget.FormatSecondaryText(question)
	cd.widget.Run()
	cb(*cd.yesno)
}

func (cd *ErrorDialog) ShowError(message string) {
	cd.widget.FormatSecondaryText(message)
	cd.widget.Run()
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
		fmt.Printf("Confirm %#v\n", result)
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
