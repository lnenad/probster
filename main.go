package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/quick"
	"github.com/lnenad/probester/communication"

	"os"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const appID = "com.mockadillo.probester"

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
	pathGrid, _ := gtk.GridNew()
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
	// Assemble the window
	pathGrid.Add(pathMethod)
	pathGrid.Add(pathInput)
	pathGrid.Add(sendRequestBtn)

	pathGrid.SetHExpand(true)

	requestBodyWindow, requestText := getJSONFrame("Request")

	requestFrame, err := gtk.FrameNew("Request")
	if err != nil {
		log.Fatal("Unable to create Frame:", err)
	}
	requestFrame.SetMarginTop(10)
	requestFrame.SetMarginEnd(10)
	requestFrame.SetMarginBottom(10)
	requestFrame.SetMarginStart(10)

	responseWindow, responseText := getJSONFrame("Response")

	responseFrame, err := gtk.FrameNew("Response")
	if err != nil {
		log.Fatal("Unable to create Frame:", err)
	}
	responseFrame.SetMarginTop(10)
	responseFrame.SetMarginEnd(10)
	responseFrame.SetMarginBottom(10)
	responseFrame.SetMarginStart(10)
	responseFrame.Add(responseText)

	responseText.SetEditable(false)

	pane, err := gtk.PanedNew(gtk.ORIENTATION_VERTICAL)
	if err != nil {
		log.Fatal("Unable to create paned:", err)
	}

	pathMethod.Connect("changed", func() {
		method := pathMethod.GetActiveText()
		if method == "GET" || method == "HEAD" {
			requestFrame.SetVisible(false)
		} else {
			requestFrame.SetVisible(true)
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
		displaySource(responseText, string(responseBody))
	})

	loadAndDispSource(requestText, "test.json")

	requestNotebook, err := gtk.NotebookNew()
	if err != nil {
		log.Fatal("Unable to create notebook:", err)
	}
	requestNotebookBodyLbl, err := gtk.LabelNew("Request Body")
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}
	requestNotebookHeadersLbl, err := gtk.LabelNew("Request Headers")
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}
	requestHeaders, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create requestHeaders grid:", err)
	}
	requestNotebook.AppendPage(requestBodyWindow, requestNotebookBodyLbl)
	requestNotebook.AppendPage(requestHeaders, requestNotebookHeadersLbl)
	requestFrame.Add(requestNotebook)
	pane.Add(requestFrame)

	pane.Add(responseFrame)
	mainGrid.Add(pathGrid)
	mainGrid.Add(pane)

	win.Add(mainGrid)
	win.SetTitlebar(header)
	win.SetPosition(gtk.WIN_POS_MOUSE)
	win.SetDefaultSize(600, 700)

	win.ShowAll()
	requestBodyWindow.SetVisible(false)

	return win
}

func getText(textView *gtk.TextView) (string, error) {
	buffer, _ := textView.GetBuffer()
	start := buffer.GetStartIter()
	end := buffer.GetEndIter()
	return buffer.GetText(start, end, true)
}

func getJSONFrame(frameLabel string) (*gtk.ScrolledWindow, *gtk.TextView) {
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

// loadAndDispSource:
func loadAndDispSource(textView *gtk.TextView, filename string) {
	text, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Unable to load file:", err)
	}
	displaySource(textView, string(text))
}

func displaySource(textView *gtk.TextView, text string) {
	// Get source formatted using pango markup format
	formattedSource, err := ChromaHighlight(text)

	// fill TextµBuffer with formatted text
	buff, err := textView.GetBuffer()
	if err != nil {
		log.Fatal("Unable to retrieve TextBuffer:", err)
	}
	// Clean text window before fill it
	buff.Delete(buff.GetStartIter(), buff.GetEndIter())

	// insert markup to the TextBuffer
	buff.InsertMarkup(buff.GetStartIter(), formattedSource)
}

// ChromaHighlight Syntax highlighter using Chroma syntax
// highlighter: "github.com/alecthomas/chroma"
// informations above
func ChromaHighlight(inputString string) (out string, err error) {
	buff := new(bytes.Buffer)
	writer := bufio.NewWriter(buff)

	// Registrering pango formatter
	formatters.Register("pango", chroma.FormatterFunc(pangoFormatter))

	// Doing the job (io.Writer, SourceText, language(go), Lexer(pango), style(pygments))
	if err = quick.Highlight(writer, inputString, "json", "pango", "pygments"); err != nil {
		return
	}
	writer.Flush()
	return string(buff.Bytes()), err
}

// pangoFormatter: is a part of "ChromaHighlight" library
// This is the Pango version, wich not use tags functionality
// but only Pango markup style. The complete libray include
// more functionalities and speed improvement of 80% using
// Tags and TextBuffer capabilities.
func pangoFormatter(w io.Writer, style *chroma.Style, it chroma.Iterator) error {
	var r, g, b uint8
	var closer, out string

	var getColour = func(color chroma.Colour) string {
		r, g, b = color.Red(), color.Green(), color.Blue()
		return fmt.Sprintf("#%02X%02X%02X", r, g, b)
	}

	for tkn := it(); tkn != chroma.EOF; tkn = it() {

		entry := style.Get(tkn.Type)
		if !entry.IsZero() {
			if entry.Bold == chroma.Yes {
				out = `<b>`
				closer = `</b>`
			}
			if entry.Underline == chroma.Yes {
				out += `<u>`
				closer = `</u>` + closer
			}
			if entry.Italic == chroma.Yes {
				out += `<i>`
				closer = `</i>` + closer
			}
			if entry.Colour.IsSet() {
				out += `<span foreground="` + getColour(entry.Colour) + `">`
				closer = `</span>` + closer
			}
			if entry.Background.IsSet() {
				out += `<span background="` + getColour(entry.Background) + `">`
				closer = `</span>` + closer
			}
			if entry.Border.IsSet() {
				out += `<span background="` + getColour(entry.Border) + `">`
				closer = `</span>` + closer
			}
			fmt.Fprint(w, out)
		}
		fmt.Fprint(w, html.EscapeString(tkn.Value))
		if !entry.IsZero() {
			fmt.Fprint(w, closer)
		}
		closer, out = "", ""
	}
	return nil
}
