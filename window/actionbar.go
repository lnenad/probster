package window

import "github.com/gotk3/gotk3/gtk"

func GetActionbar() (*gtk.ActionBar, *gtk.Label, *gtk.Label) {
	actionBar, _ := gtk.ActionBarNew()

	responseStatusLbl, _ := gtk.LabelNew("Status Code: ---")
	responseStatusLbl.SetMarginTop(10)
	responseStatusLbl.SetMarginBottom(10)
	responseStatusLbl.SetMarginEnd(10)
	responseStatusLbl.SetHAlign(gtk.ALIGN_END)

	requestDurationLbl, _ := gtk.LabelNew("Request Duration: --- ms")
	requestDurationLbl.SetMarginTop(10)
	requestDurationLbl.SetMarginBottom(10)
	requestDurationLbl.SetMarginEnd(20)
	requestDurationLbl.SetHAlign(gtk.ALIGN_END)
	actionBar.PackEnd(responseStatusLbl)
	actionBar.PackEnd(requestDurationLbl)

	return actionBar, responseStatusLbl, requestDurationLbl
}
