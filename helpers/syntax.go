package helpers

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
	"github.com/gotk3/gotk3/gtk"
)

// LoadAndDisplaySource
func LoadAndDisplaySource(textView *gtk.TextView, filename string) {
	text, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Unable to load file:", err)
	}
	DisplaySource(textView, string(text))
}

// DisplaySource
func DisplaySource(textView *gtk.TextView, text string) {
	// Get source formatted using pango markup format
	formattedSource, err := chromaHighlight(text)

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

// chromaHighlight Syntax highlighter using Chroma syntax
// highlighter: "github.com/alecthomas/chroma"
// informations above
func chromaHighlight(inputString string) (out string, err error) {
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

// pangoFormatter: is a part of "chromaHighlight" library
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
