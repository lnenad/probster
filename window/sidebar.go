package window

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gtk"
	"github.com/lnenad/probster/storage"
)

func GetSidebar(h *storage.History) (*gtk.Grid, *gtk.ListBox) {
	sideGrid, _ := gtk.GridNew()
	sideGrid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	sideGrid.SetVExpand(true)
	sideGrid.SetHExpand(true)

	listView, _ := gtk.ListBoxNew()
	listView.SetVExpand(true)
	listView.SetHExpand(true)

	requestHistory := h.GetAllRequests()
	for key, entry := range requestHistory {
		AddHistoryRow(h, listView, key, entry.Request.Method, entry.Request.Path, entry.Response.StatusCode)
	}

	listView.Connect("row_selected", func(lb *gtk.ListBox, row *gtk.ListBoxRow) {
		if lb.GetChildren().Length() > 0 {
			log.Printf("%#v\n", row.GetIndex())
		}
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

func AddHistoryRow(h *storage.History, historyListbox *gtk.ListBox, key string, method, path string, statusCode int) *gtk.ListBoxRow {
	listRow, _ := gtk.ListBoxRowNew()
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	btn, _ := gtk.ButtonNewFromIconName("edit-delete-symbolic", gtk.ICON_SIZE_BUTTON)
	btn.SetHAlign(gtk.ALIGN_END)
	btn.SetTooltipText("Remove this history entry")

	lblMethod, _ := gtk.LabelNew("")
	lblMethod.SetHExpand(true)
	if statusCode <= 299 {
		lblMethod.SetMarkup(fmt.Sprintf(`<span size='large' foreground='green'>%s</span>`, method))
	} else if statusCode > 299 && statusCode < 399 {
		lblMethod.SetMarkup(fmt.Sprintf(`<span size='large' foreground='orange'>%s</span>`, method))
	} else {
		lblMethod.SetMarkup(fmt.Sprintf(`<span size='large' foreground='red'>%s</span>`, method))
	}

	sep, _ := gtk.SeparatorMenuItemNew()

	lblPath, _ := gtk.LabelNew(path)
	lblPath.SetHExpand(true)

	box.Add(lblMethod)
	box.Add(sep)
	box.Add(lblPath)
	box.Add(btn)

	listRow.Add(box)
	listRow.SetHExpand(true)

	btn.Connect("clicked", func() {
		historyListbox.Remove(listRow)
		h.RemoveEntry(key)
	})

	historyListbox.Prepend(listRow)
	historyListbox.ShowAll()

	return listRow
}
