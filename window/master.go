package window

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	evbus "github.com/asaskevich/EventBus"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	gv "github.com/hashicorp/go-version"
	"github.com/lnenad/probster/storage"
	"github.com/lnenad/probster/update"
)

var headerRegex = regexp.MustCompile(`^[\w-]+$`)

// IDs to access the tree view columns by
const (
	ColumnKey = iota
	ColumnValue
)

var supportedMethods = []string{
	"GET",
	"POST",
	"PUT",
	"PATCH",
	"DELETE",
	"HEAD",
}

// BuildWindow is used to build main app window
func BuildWindow(
	currentVersion *gv.Version,
	settings *storage.Settings,
	application *gtk.Application,
	h *storage.HistoryStorage,
	st *storage.SettingsStorage,
	bus evbus.Bus,
) *gtk.ApplicationWindow {
	win, err := gtk.ApplicationWindowNew(application)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	win.SetTitle("Probster")

	errorDiag := getErrorDialog(win)
	confirmDiag := getConfirmDialog(win)
	nDiag := getNotificationDialog(win)
	aDiag := getAboutDialog(currentVersion)
	sDiag := getSettingsDialog(win, settings, bus)

	if should, ok := (*settings)[storage.SettingCheckUpdates].(bool); ok && should {
		if shouldUpdate, newVersion := update.CheckVersion(currentVersion); shouldUpdate {
			nDiag.ShowNotification(fmt.Sprintf("An update is available.\nYou can download the latest version from the probster.com website\nLatest version is %s", newVersion))
		} else {
			log.Infof("Software is up to date. Version: %s", newVersion)
		}
	}

	// Register header bar with menu
	registerMenu(win, bus, confirmDiag, aDiag)

	//
	// START DRAWING
	// MAIN COMPONENTS
	//

	mainGrid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create mainGrid:", err)
	}
	mainGrid.SetOrientation(gtk.ORIENTATION_VERTICAL)

	requestBodyWindow, requestText := getScrollableTextView("Request")

	requestFrame, err := gtk.FrameNew("Request")
	if err != nil {
		log.Fatal("Unable to create Frame:", err)
	}
	setMargins(requestFrame, 10, 10, 10, 10)

	responseBodyWindow, responseText := getScrollableTextView("Response")

	responseFrame, err := gtk.FrameNew("Response")
	if err != nil {
		log.Fatal("Unable to create Frame:", err)
	}
	setMargins(responseFrame, 10, 10, 10, 10)

	responseText.SetEditable(false)

	pane, err := gtk.PanedNew(gtk.ORIENTATION_VERTICAL)
	if err != nil {
		log.Fatal("Unable to create paned:", err)
	}

	requestNotebook, err := gtk.NotebookNew()
	if err != nil {
		log.Fatal("Unable to create notebook:", err)
	}
	requestNotebookBodyLbl, err := gtk.LabelNew("Body")
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}
	requestNotebookHeadersLbl, err := gtk.LabelNew("Headers")
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}
	requestHeaders, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create requestHeaders grid:", err)
	}
	requestHeaders.SetOrientation(gtk.ORIENTATION_VERTICAL)

	requestHeadersButtonBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	if err != nil {
		log.Fatal("Unable to create requestHeadersButtonBox:", err)
	}

	setMargins(requestHeadersButtonBox, 5, 5, 5, 0)

	requestTreeScroll, requestTreeView, requestStore := setupTreeView(errorDiag, true)
	deleteRequestHeaderBtn, _ := gtk.ButtonNewWithLabel("Delete selected header")
	addRequestHeaderBtn, _ := gtk.ButtonNewWithLabel("Add a new header")

	addRequestHeaderBtn.Connect("clicked", func() {
		AddRowToStore(requestStore, "Name", "Value")
	})

	deleteRequestHeaderBtn.Connect("clicked", func() {
		selection, err := requestTreeView.GetSelection()
		if err != nil {
			log.Fatal("Unable to get tree view selection:", err)
		}
		selected := selection.GetSelectedRows(&requestStore.TreeModel)
		if err != nil {
			log.Fatal("Unable to get tree view selected rows:", err)
		}
		log.Printf("%#v\n", selected)
		selected.Foreach(func(item interface{}) {
			log.Printf("%#v\n", item)
			iter, err := requestStore.GetIter(item.(*gtk.TreePath))
			if err != nil {
				log.Fatal("Unable to get tree view iter:", err)
			}
			requestStore.Remove(iter)
		})
	})

	requestHeaders.SetVExpand(true)
	requestTreeView.SetVExpand(true)
	requestTreeScroll.SetVExpand(true)
	requestHeaders.Add(requestTreeScroll)
	requestHeadersButtonBox.SetVAlign(gtk.ALIGN_END)
	requestHeadersButtonBox.PackEnd(deleteRequestHeaderBtn, false, false, 3)
	requestHeadersButtonBox.PackEnd(addRequestHeaderBtn, false, false, 3)

	sep, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)

	requestHeaders.Add(sep)
	requestHeaders.Add(requestHeadersButtonBox)

	requestNotebook.AppendPage(requestBodyWindow, requestNotebookBodyLbl)
	requestNotebook.AppendPage(requestHeaders, requestNotebookHeadersLbl)
	requestFrame.Add(requestNotebook)
	requestNotebook.SetVExpand(true)
	requestFrame.SetVExpand(true)

	pane.Add1(requestFrame)

	responseNotebook, err := gtk.NotebookNew()
	if err != nil {
		log.Fatal("Unable to create notebook:", err)
	}
	responseNotebookBodyLbl, err := gtk.LabelNew("Body")
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}
	responseNotebookHeadersLbl, err := gtk.LabelNew("Headers")
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}
	responseHeaders, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create responseHeaders grid:", err)
	}

	responseTreeScroll, _, responseStore := setupTreeView(errorDiag, false)
	responseHeaders.Add(responseTreeScroll)
	responseTreeScroll.SetVExpand(true)

	responseNotebook.AppendPage(responseBodyWindow, responseNotebookBodyLbl)
	responseNotebook.AppendPage(responseHeaders, responseNotebookHeadersLbl)

	responseFrame.Add(responseNotebook)
	pane.Add2(responseFrame)

	actionBar, highlightCheckbutton, responseStatusLbl, requestDurationLbl := GetActionbar()

	sideBar, historyListbox := GetSidebar(h, bus)

	reloadResponseBodyFn := reloadResponseBody(
		h,
		settings,
		highlightCheckbutton,
		responseText,
	)

	highlightCheckbutton.Connect("clicked", reloadResponseBodyFn)

	pathHeader, pathInput, pathMethod := getPathGrid(
		h,
		bus,
		errorDiag,
		requestText,
		requestStore,
		requestBodyWindow,
	)

	bus.Subscribe("request:completed", requestCompleted(
		h,
		settings,
		highlightCheckbutton,
		historyListbox,
		responseText,
		responseStore,
		responseStatusLbl,
		requestDurationLbl,
	))

	bus.Subscribe("request:loaded", requestLoaded(
		h,
		settings,
		highlightCheckbutton,
		pathInput,
		pathMethod,
		historyListbox,
		requestText,
		responseText,
		requestStore,
		responseStore,
		responseStatusLbl,
		requestDurationLbl,
	))

	bus.Subscribe("request:new", requestNew(
		h,
		settings,
		pathInput,
		pathMethod,
		historyListbox,
		responseText,
		requestStore,
		responseStore,
		responseStatusLbl,
		requestDurationLbl,
	))

	bus.Subscribe("history:clear", clearHistory(
		h,
		historyListbox,
	))

	bus.Subscribe("preferences:updated", settingsUpdated(
		st,
		reloadResponseBodyFn,
	))

	bus.Subscribe("preferences:show", func() {
		sDiag.Show()
	})

	mainGrid.Add(pathHeader)
	mainGrid.Add(pane)

	mainGrid.Add(actionBar)

	windowPane, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)

	sideBar.SetSizeRequest(150, 600)
	windowPane.Pack1(sideBar, true, false)
	windowPane.Pack2(mainGrid, true, true)
	win.Add(windowPane)
	win.SetPosition(gtk.WIN_POS_MOUSE)
	win.SetDefaultSize(900, 700)

	win.ShowAll()

	currMethod := pathMethod.GetActiveText()
	if currMethod == "GET" || currMethod == "HEAD" {
		requestBodyWindow.SetVisible(false)
	} else {
		requestBodyWindow.SetVisible(true)
	}

	return win
}

func setMargins(iw gtk.IWidget, top, right, bot, left int) {
	w := iw.ToWidget()
	w.SetMarginTop(top)
	w.SetMarginEnd(right)
	w.SetMarginBottom(bot)
	w.SetMarginStart(left)
}

func resolveResponseHeaders(headers http.Header) map[string][]string {
	responseHeaders := make(map[string][]string)
	for n, vals := range headers {
		name := strings.ToLower(n)
		responseHeaders[name] = []string{}
		for _, v := range vals {
			responseHeaders[name] = append(responseHeaders[name], v)
		}
	}
	return responseHeaders
}

func getListStoreContents(store *gtk.ListStore) map[string][]string {
	result := make(map[string][]string)
	iter, err := store.GetIterFirst()
	if err != true {
		return result
	}
	for {
		k, err := getStringValue(store, iter, ColumnKey)
		if err != nil {
			log.Fatal("Error getting value from store tree model: ", err)
		}
		v, err := getStringValue(store, iter, ColumnValue)
		if err != nil {
			log.Fatal("Error getting value from store tree model: ", err)
		}
		if _, ok := result[k]; ok {
			result[k] = append(result[k], v)
		} else {
			result[k] = []string{v}
		}
		if !store.IterNext(iter) {
			break
		}
	}
	return result
}

func getStringValue(store *gtk.ListStore, iter *gtk.TreeIter, column int) (string, error) {
	k, err := store.TreeModel.GetValue(iter, column)
	if err != nil {
		log.Fatal("Error getting value from store tree model: ", err)
	}
	return k.GetString()
}

func getText(textView *gtk.TextView) (string, error) {
	buffer, _ := textView.GetBuffer()
	start := buffer.GetStartIter()
	end := buffer.GetEndIter()
	return buffer.GetText(start, end, true)
}

func getScrollableTextView(frameLabel string) (*gtk.ScrolledWindow, *gtk.TextView) {
	// Label text in the window
	textView, err := gtk.TextViewNew()
	if err != nil {
		log.Fatal("Unable to create TextView:", err)
	}

	// Configure TextView
	textView.SetHExpand(true)
	textView.SetVExpand(true)
	textView.SetMarginStart(5)
	textView.SetWrapMode(gtk.WRAP_WORD)

	// Allow to scroll the text
	scrolledWindow, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatal("Unable to create ScrolledWindow:", err)
	}
	scrolledWindow.Add(textView)

	return scrolledWindow, textView
}

// Add a column to the tree view (during the initialization of the tree view)
func createColumn(errorDiag *ErrorDialog, title string, id int, editable bool, listStore *gtk.ListStore) *gtk.TreeViewColumn {
	cellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		log.Fatal("Unable to create text cell renderer:", err)
	}
	if editable {
		cellRenderer.SetProperty("editable", true)
	}
	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "text", id)
	if err != nil {
		log.Fatal("Unable to create cell column:", err)
	}
	column.SetResizable(true)

	if editable {
		cellRenderer.Connect("edited", func(crt *gtk.CellRendererText, row string, value string) {
			log.Printf("Edited: %#v %#v %#v %#v\n", title, id, row, value)
			rowIter, err := listStore.GetIterFromString(row)
			if err != nil {
				log.Fatal("Unable to get row iter:", err)
			}
			if id == ColumnKey && !headerRegex.Match([]byte(value)) {
				errorDiag.ShowError("Invalid Header Name. No spaces allowed")
				return
			}
			listStore.SetValue(rowIter, id, value)
		})
	}

	return column
}

// Creates a tree view and the list store that holds its data
func setupTreeView(errorDiag *ErrorDialog, editable bool) (*gtk.ScrolledWindow, *gtk.TreeView, *gtk.ListStore) {
	treeView, err := gtk.TreeViewNew()
	if err != nil {
		log.Fatal("Unable to create tree view:", err)
	}

	treeView.SetHExpand(true)

	// Creating a list store. This is what holds the data that will be shown on our tree view.
	listStore, err := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING)
	if err != nil {
		log.Fatal("Unable to create list store:", err)
	}
	treeView.SetModel(listStore)

	treeView.AppendColumn(createColumn(errorDiag, "Header Name", ColumnKey, editable, listStore))
	treeView.AppendColumn(createColumn(errorDiag, "Header Value", ColumnValue, editable, listStore))

	scrolledWindow, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatal("Unable to create ScrolledWindow:", err)
	}
	scrolledWindow.Add(treeView)

	return scrolledWindow, treeView, listStore
}

// AddRowToStore append a row to the list store for the tree view
func AddRowToStore(listStore *gtk.ListStore, key, value string) {
	// Get an iterator for a new row at the end of the list store
	iter := listStore.Append()

	// Set the contents of the list store row that the iterator represents
	err := listStore.Set(iter,
		[]int{ColumnKey, ColumnValue},
		[]interface{}{key, value})

	if err != nil {
		log.Fatal("Unable to add row:", err)
	}
}
