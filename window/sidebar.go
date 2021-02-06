package window

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	evbus "github.com/asaskevich/EventBus"
	"github.com/gotk3/gotk3/gtk"
	"github.com/lnenad/probster/storage"
)

func GetSidebar(h *storage.History, bus evbus.Bus) (*gtk.Grid, *gtk.ListBox) {
	sideGrid, _ := gtk.GridNew()
	sideGrid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	sideGrid.SetVExpand(true)
	sideGrid.SetHExpand(true)

	// Allow to scroll the entries
	scrolledWindow, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatal("Unable to create ScrolledWindow:", err)
	}

	listView, _ := gtk.ListBoxNew()

	scrolledWindow.SetVExpand(true)
	scrolledWindow.Add(listView)

	requestHistory := h.GetAllRequests()
	for _, entry := range requestHistory {
		AddHistoryRow(h, listView, entry.Key, entry.RR)
	}

	listView.Connect("row_selected", func(lb *gtk.ListBox, row *gtk.ListBoxRow) {
		if lb.GetSelectedRows().Length() > 0 {
			id, err := row.GetName()
			if err != nil {
				log.Printf("Error getting row id: %s", err)
				listView.UnselectAll()
			} else {
				entry := h.GetEntry(id)
				bus.Publish("request:loaded", entry.RR)
			}
		}
	})

	historyLbl, _ := gtk.LabelNew("")
	historyLbl.SetMarkup("<span size='large'>Request History</span>")
	historyLbl.SetHAlign(gtk.ALIGN_START)
	setMargins(historyLbl, 10, 10, 10, 10)

	historySep, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)

	sideGrid.Add(historyLbl)
	sideGrid.Add(historySep)
	sideGrid.Add(scrolledWindow)

	return sideGrid, listView
}

func AddHistoryRow(
	h *storage.History,
	historyListbox *gtk.ListBox,
	key string,
	reqRes storage.RequestResponse,
) *gtk.ListBoxRow {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	btn, _ := gtk.ButtonNewFromIconName("edit-delete-symbolic", gtk.ICON_SIZE_BUTTON)
	btn.SetHAlign(gtk.ALIGN_START)
	btn.SetTooltipText("Remove this history entry")

	lblMethod, _ := gtk.LabelNew("")
	//lblMethod.SetHExpand(true)
	lblMethod.SetWidthChars(11)
	if reqRes.Response.StatusCode <= 299 {
		lblMethod.SetMarkup(fmt.Sprintf(`<span size='large' foreground='green'>%s</span>`, reqRes.Request.Method))
	} else if reqRes.Response.StatusCode > 299 && reqRes.Response.StatusCode < 399 {
		lblMethod.SetMarkup(fmt.Sprintf(`<span size='large' foreground='orange'>%s</span>`, reqRes.Request.Method))
	} else {
		lblMethod.SetMarkup(fmt.Sprintf(`<span size='large' foreground='red'>%s</span>`, reqRes.Request.Method))
	}

	sep, _ := gtk.SeparatorMenuItemNew()
	sep2, _ := gtk.SeparatorMenuItemNew()

	lblPath, _ := gtk.LabelNew(reqRes.Request.Path)
	lblPath.SetHAlign(gtk.ALIGN_START)
	lblPath.SetMarginStart(10)
	lblPath.SetMarginEnd(20)

	box.Add(btn)
	box.Add(sep2)
	box.Add(lblMethod)
	box.Add(sep)
	box.Add(lblPath)

	listRow, _ := gtk.ListBoxRowNew()
	listRow.Add(box)
	listRow.SetHExpand(true)
	listRow.SetName(key)
	listRow.SetTooltipText("Load this request")

	btn.Connect("clicked", func() {
		historyListbox.Remove(listRow)
		go h.RemoveEntry(key)
	})

	historyListbox.Prepend(listRow)
	historyListbox.ShowAll()

	return listRow
}
