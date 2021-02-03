package window

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
)

func GetSidebar() (*gtk.Grid, *gtk.ListBox) {
	sideGrid, _ := gtk.GridNew()
	sideGrid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	sideGrid.SetVExpand(true)
	sideGrid.SetHExpand(true)

	listView, _ := gtk.ListBoxNew()
	listView.SetVExpand(true)
	listView.SetHExpand(true)

	addHistoryRow(listView, "GET", "https://httpbin.org/get", 200)
	addHistoryRow(listView, "POST", "https://httpbin.org/post", 201)
	addHistoryRow(listView, "POST", "https://httpbin.org/post", 405)
	addHistoryRow(listView, "POST", "https://httpbin.org/post", 303)

	listView.Connect("row_selected", func(lb *gtk.ListBox, row *gtk.ListBoxRow) {
		fmt.Printf("%#v\n", row.GetIndex())
	})

	historyLbl, _ := gtk.LabelNew("")
	historyLbl.SetMarkup("<span size='large'>Request History</span>")
	historyLbl.SetHAlign(gtk.ALIGN_START)
	setMargins(historyLbl, 10, 10, 10, 10)

	historySep, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)

	sideGrid.Add(historyLbl)
	sideGrid.Add(historySep)
	sideGrid.Add(listView)

	return sideGrid, listView
}
