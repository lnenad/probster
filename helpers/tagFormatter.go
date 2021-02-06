package helpers

import (
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
)

// TextTagList yo
var TextTagList = make(map[string]*gtk.TextTag)

type formatterFunc func(w io.Writer, style *chroma.Style, it chroma.Iterator) error

func tagFormatter(buff *gtk.TextBuffer) formatterFunc {
	return func(w io.Writer, style *chroma.Style, it chroma.Iterator) error {
		for tkn := it(); tkn != chroma.EOF; tkn = it() {
			tagName := strings.ToLower(tkn.Type.String())
			entry := style.Get(tkn.Type)
			startIter := buff.GetEndIter()
			if !entry.IsZero() {
				if _, ok := TextTagList[tagName]; !ok {
					tagProps := buildTagProps(buff, tagName, &entry)
					TextTagList[tagName] = buff.CreateTag(tagName, tagProps)
				}
				buff.InsertWithTag(startIter, tkn.Value, TextTagList[tagName])
			} else {
				buff.Insert(startIter, tkn.Value)
			}
		}
		return nil
	}
}

func buildTagProps(buff *gtk.TextBuffer, name string, styleEntry *chroma.StyleEntry) map[string]interface{} {
	tagProp := map[string]interface{}{}
	if styleEntry.Bold == chroma.Yes {
		tagProp["weight"] = pango.WEIGHT_BOLD
	}
	if styleEntry.Italic == chroma.Yes {
		tagProp["style"] = pango.STYLE_ITALIC
	}
	if styleEntry.Underline == chroma.Yes {
		tagProp["underline"] = pango.UNDERLINE_SINGLE
	}
	if styleEntry.Colour.IsSet() {
		tagProp["foreground"] = fmt.Sprintf("#%02X%02X%02X",
			styleEntry.Colour.Red(),
			styleEntry.Colour.Green(),
			styleEntry.Colour.Blue())
	}
	if styleEntry.Background.IsSet() {
		tagProp["background"] = fmt.Sprintf("#%02X%02X%02X",
			styleEntry.Background.Red(),
			styleEntry.Background.Green(),
			styleEntry.Background.Blue())
	}
	if styleEntry.Border.IsSet() {
		tagProp["background"] = fmt.Sprintf("#%02X%02X%02X",
			styleEntry.Border.Red(),
			styleEntry.Border.Green(),
			styleEntry.Border.Blue())
	}
	return tagProp
}
