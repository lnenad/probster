package window

import (
	"fmt"
	"net/url"
	"time"

	evbus "github.com/asaskevich/EventBus"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/lnenad/probster/communication"
	"github.com/lnenad/probster/storage"
	log "github.com/sirupsen/logrus"
)

func getPathGrid(
	h *storage.History,
	bus evbus.Bus,
	errorDiag *ErrorDialog,
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

	performRequest := func() {
		path, _ := pathInput.GetText()
		method := pathMethod.GetActiveText()

		res, err := url.Parse(path)
		if err != nil {
			errorDiag.ShowError(fmt.Sprintf("Invalid URL provided. %s", err))
			return
		}
		if res.Scheme != "http" && res.Scheme != "https" {
			errorDiag.ShowError(fmt.Sprintf("Invalid URL Scheme provided.\nPlease start the url with http:// or https://"))
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
			response, responseBody, err := communication.Send(path, method, requestHeaders, requestBody)
			if err != nil {
				glib.IdleAdd(func() {
					errorDiag.ShowError(fmt.Sprintf("Error while performing request.\n%s", err))
					return
				})
				return
			}
			log.Printf("Response: %#v\n", response)

			glib.IdleAdd(func(reqRes storage.RequestResponse) {
				bus.Publish("request:completed", reqRes)
				sendRequestBtn.SetSensitive(true)
			}, storage.RequestResponse{
				Request: storage.RequestInput{
					Body:    requestBody,
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
	}

	pathInput.Connect("activate", performRequest)

	sendRequestBtn.Connect("clicked", performRequest)

	// Assemble the window
	pathGrid.Add(pathMethod)
	pathGrid.Add(pathInput)
	pathGrid.Add(sendRequestBtn)

	pathGrid.SetHExpand(true)
	return pathGrid, pathInput, pathMethod
}
