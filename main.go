package main

import (
	"fmt"
	"log"
	"regexp"

	"github.com/lnenad/probster/communication"
	"github.com/lnenad/probster/helpers"

	"os"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const appID = "com.mockadillo.probster"

var headerRegex = regexp.MustCompile(`^[\w-]+$`)

// IDs to access the tree view columns by
const (
	ColumnKey = iota
	ColumnValue
)

func main() {
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal("Could not create application:", err)
	}

	application.Connect("activate", func() {
		buildWindow(application)

		aNew := glib.SimpleActionNew("new", nil)
		aNew.Connect("activate", func() {
			buildWindow(application).ShowAll()
		})
		application.AddAction(aNew)

		aQuit := glib.SimpleActionNew("quit", nil)
		aQuit.Connect("activate", func() {
			application.Quit()
		})
		application.AddAction(aQuit)
	})

	os.Exit(application.Run(os.Args))
}

func buildWindow(application *gtk.Application) *gtk.ApplicationWindow {
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
		fmt.Println("New request bro")
	})
	win.AddAction(nRequest)

	mbtn.SetMenuModel(&menu.MenuModel)

	// add the menu button to the header
	header.PackStart(mbtn)

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
	requestFrame.SetMarginTop(10)
	requestFrame.SetMarginEnd(10)
	requestFrame.SetMarginBottom(10)
	requestFrame.SetMarginStart(10)

	responseBodyWindow, responseText := getScrollableTextView("Response")

	responseFrame, err := gtk.FrameNew("Response")
	if err != nil {
		log.Fatal("Unable to create Frame:", err)
	}
	responseFrame.SetMarginTop(10)
	responseFrame.SetMarginEnd(10)
	responseFrame.SetMarginBottom(10)
	responseFrame.SetMarginStart(10)

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

	requestHeadersButtonBox.SetMarginBottom(5)
	requestHeadersButtonBox.SetMarginEnd(10)
	requestHeadersButtonBox.SetBorderWidth(10)

	requestTreeScroll, requestTreeView, requestStore := setupTreeView(errorDiag, true)
	deleteRequestHeaderBtn, _ := gtk.ButtonNewWithLabel("Delete selected header")
	addRequestHeaderBtn, _ := gtk.ButtonNewWithLabel("Add a new header")

	addRequestHeaderBtn.Connect("clicked", func() {
		addRow(requestStore, "Name", "Value")
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
		fmt.Printf("%#v\n", selected)
		selected.Foreach(func(item interface{}) {
			fmt.Printf("%#v\n", item)
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

	responseStatusLbl, _ := gtk.LabelNew("Status Code: ---")
	responseStatusLbl.SetMarginBottom(10)
	responseStatusLbl.SetMarginEnd(20)
	responseStatusLbl.SetHAlign(gtk.ALIGN_END)

	pathGrid := getPathGrid(requestText, responseText, requestStore, responseStore, responseStatusLbl, requestBodyWindow)

	mainGrid.Add(pathGrid)
	mainGrid.Add(pane)
	mainGrid.Add(responseStatusLbl)

	sideGrid, _ := gtk.GridNew()
	listView, _ := gtk.ListBoxNew()
	listView.SetVExpand(true)
	listView.SetHExpand(true)
	sideGrid.SetVExpand(true)
	sideGrid.SetHExpand(true)
	listRow := getHistoryRow("Item name")

	listView.Add(listRow)
	sideGrid.Add(listView)

	windowPane, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)

	windowPane.Add(sideGrid)
	windowPane.Add(mainGrid)
	win.Add(windowPane)
	win.SetTitlebar(header)
	win.SetPosition(gtk.WIN_POS_MOUSE)
	win.SetDefaultSize(900, 700)

	win.ShowAll()

	requestBodyWindow.SetVisible(false)

	return win
}

func getHistoryRow(itemName string) *gtk.ListBoxRow {
	listRow, _ := gtk.ListBoxRowNew()
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	btn, _ := gtk.ButtonNew()
	btn.SetHAlign(gtk.ALIGN_END)
	lbl, _ := gtk.LabelNew(itemName)
	lbl.SetHExpand(true)
	btn.SetLabel("Delete")
	box.Add(lbl)
	box.Add(btn)
	listRow.Add(box)
	listRow.SetHExpand(true)
	return listRow
}

func getPathGrid(
	requestText, responseText *gtk.TextView,
	requestStore, responseStore *gtk.ListStore,
	responseStatusLbl *gtk.Label,
	requestBodyWindow *gtk.ScrolledWindow,
) *gtk.Grid {
	pathGrid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create pathGrid:", err)
	}
	pathGrid.SetMarginTop(10)
	pathGrid.SetMarginEnd(10)
	pathGrid.SetMarginBottom(10)
	pathGrid.SetMarginStart(10)

	pathMethod, err := gtk.ComboBoxTextNew()
	if err != nil {
		log.Fatal("Unable to create pathMethod:", err)
	}
	pathMethod.AppendText("GET")
	pathMethod.AppendText("POST")
	pathMethod.AppendText("PUT")
	pathMethod.AppendText("PATCH")
	pathMethod.AppendText("DELETE")
	pathMethod.AppendText("HEAD")
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
		url, _ := pathInput.GetText()
		method := pathMethod.GetActiveText()
		requestBody, err := getText(requestText)
		if err != nil {
			log.Fatal("Unable to retrieve text from requestTextView:", err)
		}
		requstHeaders := getListStoreContents(requestStore)
		response, responseBody := communication.Send(url, method, requstHeaders, requestBody)
		fmt.Printf("Response: %#v\n", response)
		contentType := response.Header.Get("content-type")
		helpers.DisplaySource(contentType, responseText, string(responseBody))
		responseStore.Clear()
		for name, values := range response.Header {
			// Loop over all values for the name.
			for _, value := range values {
				fmt.Println(name, value)
				addRow(responseStore, name, value)
			}
		}
		responseStatusLbl.SetText(fmt.Sprintf("Status Code: %d", response.StatusCode))
	})

	// Assemble the window
	pathGrid.Add(pathMethod)
	pathGrid.Add(pathInput)
	pathGrid.Add(sendRequestBtn)

	pathGrid.SetHExpand(true)
	return pathGrid
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
			fmt.Printf("Edited: %#v %#v %#v %#v\n", title, id, row, value)
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
func addRow(listStore *gtk.ListStore, key, value string) {
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
