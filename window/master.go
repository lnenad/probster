package window

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	evbus "github.com/asaskevich/EventBus"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/lnenad/probster/communication"
	"github.com/lnenad/probster/helpers"
	"github.com/lnenad/probster/storage"
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
func BuildWindow(application *gtk.Application, h *storage.History, bus evbus.Bus) *gtk.ApplicationWindow {
	win, err := gtk.ApplicationWindowNew(application)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	win.SetTitle("Probster")

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
	menu.Append("Close Window", "win.close")
	menu.Append("Quit", "app.quit")

	// Create the action "win.close"
	aClose := glib.SimpleActionNew("close", nil)
	aClose.Connect("activate", func() {
		win.Close()
	})
	win.AddAction(aClose)

	// Create the action "win.new-request"
	nRequest := glib.SimpleActionNew("new-request", nil)
	nRequest.Connect("activate", func() {
		bus.Publish("request:new")
	})
	win.AddAction(nRequest)

	mbtn.SetMenuModel(&menu.MenuModel)

	// add the menu button to the header
	header.PackStart(mbtn)

	// START DRAWING MAIN COMPONENTS

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

	helpers.LoadAndDisplaySource("application/json", requestText, "test.json")

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

	actionBar, responseStatusLbl, requestDurationLbl := GetActionbar()

	sideBar, historyListbox := GetSidebar(h, bus)

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
		historyListbox,
		responseText,
		responseStore,
		responseStatusLbl,
		requestDurationLbl,
	))

	bus.Subscribe("request:loaded", requestLoaded(
		h,
		pathInput,
		pathMethod,
		historyListbox,
		responseText,
		requestStore,
		responseStore,
		responseStatusLbl,
		requestDurationLbl,
	))

	bus.Subscribe("request:new", requestNew(
		pathInput,
		pathMethod,
		historyListbox,
		responseText,
		requestStore,
		responseStore,
		responseStatusLbl,
		requestDurationLbl,
	))

	mainGrid.Add(pathHeader)
	mainGrid.Add(pane)

	mainGrid.Add(actionBar)

	windowPane, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)

	windowPane.Add(sideBar)
	windowPane.Add(mainGrid)
	win.Add(windowPane)
	win.SetTitlebar(header)
	win.SetPosition(gtk.WIN_POS_MOUSE)
	win.SetDefaultSize(900, 700)

	win.ShowAll()

	requestBodyWindow.SetVisible(false)

	return win
}

func requestCompleted(
	h *storage.History,
	historyListbox *gtk.ListBox,
	responseText *gtk.TextView,
	responseStore *gtk.ListStore,
	responseStatusLbl *gtk.Label,
	requestDurationLbl *gtk.Label,
) func(reqRes storage.RequestResponse) error {
	return func(reqRes storage.RequestResponse) error {
		helpers.DisplaySource(
			resolveContentType(reqRes.Response.Headers),
			responseText,
			string(reqRes.Response.ResponseBody),
		)
		responseStore.Clear()
		for name, values := range reqRes.Response.Headers {
			for _, value := range values {
				AddRowToStore(responseStore, name, value)
			}
		}
		responseStatusLbl.SetText(fmt.Sprintf("Status Code: %d", reqRes.Response.StatusCode))
		requestDurationLbl.SetText(fmt.Sprintf("Request Duration: %d ms", reqRes.Response.Dur.Milliseconds()))
		key := []byte(time.Now().Format(storage.HistoryKeyFormat))
		AddHistoryRow(
			h,
			historyListbox,
			string(key),
			reqRes,
		)

		h.RequestCompleted(key, reqRes)
		return nil
	}
}

func requestLoaded(
	h *storage.History,
	pathInput *gtk.Entry,
	pathMethod *gtk.ComboBoxText,
	historyListbox *gtk.ListBox,
	responseText *gtk.TextView,
	requestStore *gtk.ListStore,
	responseStore *gtk.ListStore,
	responseStatusLbl *gtk.Label,
	requestDurationLbl *gtk.Label,
) func(reqRes storage.RequestResponse) error {
	return func(reqRes storage.RequestResponse) error {
		helpers.DisplaySource(
			resolveContentType(reqRes.Response.Headers),
			responseText,
			string(reqRes.Response.ResponseBody),
		)
		requestStore.Clear()
		responseStore.Clear()
		for name, values := range reqRes.Response.Headers {
			for _, value := range values {
				AddRowToStore(responseStore, name, value)
			}
		}
		for name, values := range reqRes.Request.Headers {
			for _, value := range values {
				AddRowToStore(requestStore, name, value)
			}
		}
		responseStatusLbl.SetText(fmt.Sprintf("Status Code: %d", reqRes.Response.StatusCode))
		requestDurationLbl.SetText(fmt.Sprintf("Request Duration: %d ms", reqRes.Response.Dur.Milliseconds()))

		pathInput.SetText(reqRes.Request.Path)
		for idx, v := range supportedMethods {
			if v == reqRes.Request.Method {
				pathMethod.SetActive(idx)
				return nil
			}
		}
		pathMethod.SetActive(0)

		return nil
	}
}

func requestNew(
	pathInput *gtk.Entry,
	pathMethod *gtk.ComboBoxText,
	historyListbox *gtk.ListBox,
	responseText *gtk.TextView,
	requestStore *gtk.ListStore,
	responseStore *gtk.ListStore,
	responseStatusLbl *gtk.Label,
	requestDurationLbl *gtk.Label,
) func() error {
	return func() error {
		helpers.DisplaySource(
			"",
			responseText,
			"",
		)
		requestStore.Clear()
		responseStore.Clear()
		historyListbox.UnselectAll()
		responseStatusLbl.SetText("Status Code: ---")
		requestDurationLbl.SetText("Request Duration: --- ms")

		pathInput.SetText("https://")
		pathMethod.SetActive(0)

		return nil
	}
}

func resolveContentType(headers map[string][]string) string {
	var contentType string
	if val, ok := headers["content-type"]; ok {
		contentType = val[0]
	} else {
		contentType = ""
	}
	return contentType
}

func setMargins(iw gtk.IWidget, top, right, bot, left int) {
	w := iw.ToWidget()
	w.SetMarginTop(top)
	w.SetMarginEnd(right)
	w.SetMarginBottom(bot)
	w.SetMarginStart(left)
}

func getPathGrid(
	h *storage.History,
	bus evbus.Bus,
	errorDiag *gtk.MessageDialog,
	requestText *gtk.TextView,
	requestStore *gtk.ListStore,
	requestBodyWindow *gtk.ScrolledWindow,
) (*gtk.Grid, *gtk.Entry, *gtk.ComboBoxText) {
	pathGrid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create pathGrid:", err)
	}
	setMargins(pathGrid, 10, 10, 10, 10)

	pathMethod, err := gtk.ComboBoxTextNew()
	if err != nil {
		log.Fatal("Unable to create pathMethod:", err)
	}
	for _, method := range supportedMethods {
		pathMethod.AppendText(method)
	}
	pathMethod.SetActive(0)
	pathMethod.SetTooltipText("Select the request method")

	pathInput, err := gtk.EntryNew()
	if err != nil {
		log.Fatal("Unable to create pathInput:", err)
	}
	pathInput.SetPlaceholderText("https://google.com")
	pathInput.SetHExpand(true)

	sendRequestBtn, err := gtk.ButtonNewWithLabel("SEND")
	if err != nil {
		log.Fatal("Unable to create Button:", err)
	}

	pathMethod.Connect("changed", func() {
		method := pathMethod.GetActiveText()
		if method == "GET" || method == "HEAD" {
			requestBodyWindow.SetVisible(false)
		} else {
			requestBodyWindow.SetVisible(true)
		}
	})
	sendRequestBtn.Connect("clicked", func() {
		path, _ := pathInput.GetText()
		method := pathMethod.GetActiveText()

		res, err := url.Parse(path)
		if err != nil {
			errorDiag.FormatSecondaryText(fmt.Sprintf("Invalid URL provided. %s", err))
			errorDiag.Run()
			return
		}
		if res.Scheme != "http" && res.Scheme != "https" {
			errorDiag.FormatSecondaryText(fmt.Sprintf("Invalid URL Scheme provided.\nPlease start the url with http:// or https://"))
			errorDiag.Run()
			return
		}
		sendRequestBtn.SetSensitive(false)

		go func() {
			requestBody, err := getText(requestText)
			if err != nil {
				log.Fatal("Unable to retrieve text from requestTextView:", err)
			}
			requestHeaders := getListStoreContents(requestStore)
			start := time.Now()
			response, responseBody := communication.Send(path, method, requestHeaders, requestBody)
			log.Printf("Response: %#v\n", response)

			glib.IdleAdd(func(reqRes storage.RequestResponse) {
				bus.Publish("request:completed", reqRes)
				sendRequestBtn.SetSensitive(true)
			}, storage.RequestResponse{
				Request: storage.RequestInput{
					Path:    path,
					Method:  method,
					Headers: requestHeaders,
				},
				Response: storage.RequestResult{
					StatusCode:   response.StatusCode,
					Headers:      resolveResponseHeaders(response.Header),
					ResponseBody: responseBody,
					Dur:          time.Now().Sub(start),
				},
			})
		}()
	})

	// Assemble the window
	pathGrid.Add(pathMethod)
	pathGrid.Add(pathInput)
	pathGrid.Add(sendRequestBtn)

	pathGrid.SetHExpand(true)
	return pathGrid, pathInput, pathMethod
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

	// Allow to scroll the text
	scrolledWindow, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatal("Unable to create ScrolledWindow:", err)
	}
	scrolledWindow.Add(textView)

	return scrolledWindow, textView
}

// Add a column to the tree view (during the initialization of the tree view)
func createColumn(errorDiag *gtk.MessageDialog, title string, id int, editable bool, listStore *gtk.ListStore) *gtk.TreeViewColumn {
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
				errorDiag.FormatSecondaryText("Invalid Header Name. No spaces allowed")
				errorDiag.Run()
				return
			}
			listStore.SetValue(rowIter, id, value)
		})
	}

	return column
}

// Creates a tree view and the list store that holds its data
func setupTreeView(errorDiag *gtk.MessageDialog, editable bool) (*gtk.ScrolledWindow, *gtk.TreeView, *gtk.ListStore) {
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

// Append a row to the list store for the tree view
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
