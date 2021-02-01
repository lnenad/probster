package main

import (
	"fmt"
	"log"

	"github.com/lnenad/probester/communication"
	"github.com/lnenad/probester/helpers"

	"os"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const appID = "com.mockadillo.probester"

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

	win.SetTitle("Probester")

	// Create a header bar
	header, err := gtk.HeaderBarNew()
	if err != nil {
		log.Fatal("Could not create header bar:", err)
	}
	header.SetShowCloseButton(true)
	header.SetTitle("Probester")
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

	helpers.LoadAndDisplaySource(requestText, "test.json")

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

	requestHeadersButtonBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		log.Fatal("Unable to create requestHeadersButtonBox:", err)
	}

	requestTreeView, requestStore := setupTreeView(true)
	deleteRequestHeaderBtn, _ := gtk.ButtonNewWithLabel("Delete selected header")
	addRequestHeaderBtn, _ := gtk.ButtonNewWithLabel("Add a new header")

	addRequestHeaderBtn.Connect("clicked", func() {
		addRow(requestStore, "", "")
	})

	deleteRequestHeaderBtn.Connect("clicked", func() {
		// requestTreeView.getAc
	})

	requestHeaders.SetVExpand(true)
	requestHeaders.Add(requestTreeView)
	requestHeadersButtonBox.SetVAlign(gtk.ALIGN_END)
	requestHeadersButtonBox.PackEnd(deleteRequestHeaderBtn, false, false, 5)
	requestHeadersButtonBox.PackEnd(addRequestHeaderBtn, false, false, 5)
	requestHeaders.Add(requestHeadersButtonBox)

	requestNotebook.AppendPage(requestBodyWindow, requestNotebookBodyLbl)
	requestNotebook.AppendPage(requestHeaders, requestNotebookHeadersLbl)
	requestFrame.Add(requestNotebook)

	pane.Add(requestFrame)

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

	responseTreeView, responseStore := setupTreeView(false)
	responseHeaders.Add(responseTreeView)

	responseNotebook.AppendPage(responseBodyWindow, responseNotebookBodyLbl)
	responseNotebook.AppendPage(responseHeaders, responseNotebookHeadersLbl)
	responseFrame.Add(responseNotebook)
	pane.Add(responseFrame)

	pathGrid := getPathGrid(requestText, responseText, requestStore, responseStore)

	mainGrid.Add(pathGrid)
	mainGrid.Add(pane)

	win.Add(mainGrid)
	win.SetTitlebar(header)
	win.SetPosition(gtk.WIN_POS_MOUSE)
	win.SetDefaultSize(600, 700)

	win.ShowAll()
	//requestBodyWindow.SetVisible(false)

	return win
}

func getPathGrid(requestText, responseText *gtk.TextView, requestStore, responseStore *gtk.ListStore) *gtk.Grid {
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
			//requestFrame.SetVisible(false)
		} else {
			//requestFrame.SetVisible(true)
		}
	})
	sendRequestBtn.Connect("clicked", func() {
		url, _ := pathInput.GetText()
		method := pathMethod.GetActiveText()
		requestBody, err := getText(requestText)
		if err != nil {
			log.Fatal("Unable to retrieve text from requestTextView:", err)
		}
		response, responseBody := communication.Send(url, method, nil, requestBody)
		fmt.Printf("Response: %#v\n", response)
		helpers.DisplaySource(responseText, string(responseBody))
		responseStore.Clear()
		for name, values := range response.Header {
			// Loop over all values for the name.
			for _, value := range values {
				fmt.Println(name, value)
				addRow(responseStore, name, value)
			}
		}
	})

	// Assemble the window
	pathGrid.Add(pathMethod)
	pathGrid.Add(pathInput)
	pathGrid.Add(sendRequestBtn)

	pathGrid.SetHExpand(true)
	return pathGrid
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
func createColumn(title string, id int, editable bool) *gtk.TreeViewColumn {
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
		})
	}

	return column
}

// Creates a tree view and the list store that holds its data
func setupTreeView(editable bool) (*gtk.TreeView, *gtk.ListStore) {
	treeView, err := gtk.TreeViewNew()
	if err != nil {
		log.Fatal("Unable to create tree view:", err)
	}

	treeView.SetHExpand(true)

	treeView.AppendColumn(createColumn("Header Name", ColumnKey, editable))
	treeView.AppendColumn(createColumn("Header Value", ColumnValue, editable))

	// Creating a list store. This is what holds the data that will be shown on our tree view.
	listStore, err := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING)
	if err != nil {
		log.Fatal("Unable to create list store:", err)
	}
	treeView.SetModel(listStore)

	return treeView, listStore
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
