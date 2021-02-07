package window

import (
	"fmt"
	"time"

	"github.com/gotk3/gotk3/gtk"
	"github.com/lnenad/probster/helpers"
	"github.com/lnenad/probster/storage"
)

func settingsUpdated(
	st *storage.SettingsStorage,
	reloadResponseBodyFn func() error,
) func(storage.Settings) error {
	return func(newSettings storage.Settings) error {
		for k, v := range newSettings {
			st.UpdateSetting(k, v)
		}
		return reloadResponseBodyFn()
	}
}

func clearHistory(
	h *storage.HistoryStorage,
	historyListbox *gtk.ListBox,
) func() error {
	return func() error {
		chl := historyListbox.GetChildren()
		chl.Foreach(func(ch interface{}) {
			historyListbox.Remove(ch.(*gtk.Widget))
		})
		h.RemoveAll()
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

func requestCompleted(
	h *storage.HistoryStorage,
	settings *storage.Settings,
	highlightCheckbutton *gtk.CheckButton,
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
			highlightCheckbutton.GetActive(),
			settings,
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
		historyListbox.UnselectAll()

		h.RequestCompleted(key, reqRes)
		h.SetActiveRecord(&reqRes)

		return nil
	}
}

func requestLoaded(
	h *storage.HistoryStorage,
	settings *storage.Settings,
	highlightCheckbutton *gtk.CheckButton,
	pathInput *gtk.Entry,
	pathMethod *gtk.ComboBoxText,
	historyListbox *gtk.ListBox,
	requestText *gtk.TextView,
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
			highlightCheckbutton.GetActive(),
			settings,
		)
		rqTxtBuff, _ := requestText.GetBuffer()
		rqTxtBuff.SetText(reqRes.Request.Body)
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
		h.SetActiveRecord(&reqRes)

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

func reloadResponseBody(
	h *storage.HistoryStorage,
	settings *storage.Settings,
	highlightCheckbutton *gtk.CheckButton,
	responseText *gtk.TextView,
) func() error {
	return func() error {
		reqRes := h.GetActiveRecord()
		helpers.DisplaySource(
			resolveContentType(reqRes.Response.Headers),
			responseText,
			string(reqRes.Response.ResponseBody),
			highlightCheckbutton.GetActive(),
			settings,
		)

		return nil
	}
}

func requestNew(
	h *storage.HistoryStorage,
	settings *storage.Settings,
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
			false,
			settings,
		)
		requestStore.Clear()
		responseStore.Clear()
		historyListbox.UnselectAll()
		responseStatusLbl.SetText("Status Code: ---")
		requestDurationLbl.SetText("Request Duration: --- ms")

		pathInput.SetText("https://")
		pathMethod.SetActive(0)
		h.SetActiveRecord(nil)

		return nil
	}
}
