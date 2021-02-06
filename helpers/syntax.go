package helpers

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/quick"
	log "github.com/sirupsen/logrus"

	"github.com/gotk3/gotk3/gtk"
)

const HTMLContentType = "text/html"

// LoadAndDisplaySource
func LoadAndDisplaySource(contentType string, textView *gtk.TextView, filename string) {
	text, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Unable to load file:", err)
	}
	DisplaySource(contentType, textView, string(text), true)
}

// DisplaySource used to load text and perform highlighting
func DisplaySource(contentType string, textView *gtk.TextView, text string, highlight bool) {
	formatter := "tag"
	// Get source formatted using pango markup format
	if highlight {

		// fill TextµBuffer with formatted text
		buff, err := textView.GetBuffer()
		if err != nil {
			log.Fatal("Unable to retrieve TextBuffer:", err)
		}

		textView.SetBuffer(nil)
		buff.Delete(buff.GetStartIter(), buff.GetEndIter())

		formattedSource, err := chromaHighlight(buff, contentType, text, formatter)
		if err != nil {
			log.Fatal("Unable to perform highlighting:", err)
		}

		if formatter == "pango" {
			buff.InsertMarkup(buff.GetStartIter(), formattedSource)
		}

		textView.SetBuffer(buff)
	} else {
		// fill TextµBuffer with formatted text
		buff, err := textView.GetBuffer()
		if err != nil {
			log.Fatal("Unable to retrieve TextBuffer:", err)
		}
		textView.SetBuffer(nil)
		buff.SetText(text)
		textView.SetBuffer(buff)
	}
}

func chromaHighlight(tbuff *gtk.TextBuffer, contentType, inputString, formatter string) (out string, err error) {
	buff := new(bytes.Buffer)
	writer := bufio.NewWriter(buff)

	// Registrering pango formatter
	formatters.Register("tag", chroma.FormatterFunc(tagFormatter(tbuff)))
	formatters.Register("pango", chroma.FormatterFunc(pangoFormatter))

	if err = quick.Highlight(writer, inputString, getLanguage(contentType), formatter, "github"); err != nil {
		return "", err
	}
	writer.Flush()
	return string(buff.Bytes()), err
}

func getLanguage(contentType string) string {
	var language string
	switch contentType {
	case "application/json":
		language = "json"
	case HTMLContentType:
		language = "html"
	case "text/xml":
		fallthrough
	case "application/xml":
		fallthrough
	case "image/svg+xml":
		language = "xml"
	}
	if strings.Contains(contentType, HTMLContentType) {
		language = "html"
	}
	return language
}
