package helpers

import (
	"fmt"
	"html"
	"io"

	"github.com/alecthomas/chroma"
)

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
