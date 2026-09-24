package main

import (
	"embed"
	"fmt"
	"log"
	"strconv"

	"github.com/movsb/fbiw"
	"github.com/movsb/fbiw-widgets/markdown"
	"github.com/movsb/fbiw/input/sticks"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/goregular"
)

//go:embed main.html sample.md
var files embed.FS

func main() {
	app := fbiw.NewApp(
		fbiw.WithSystemFontData(goregular.TTF),
		fbiw.WithFontData(`monospace`, false, false, gomono.TTF),
	)
	defer app.Close()

	doc := app.NewDesktop(files, "main.html")
	viewport := doc.GetBoxByID[*fbiw.Scroll]("viewport")
	content := doc.GetBoxByID[fbiw.Box]("content")
	status := doc.GetBoxByID[*fbiw.Text]("status")

	source, err := files.ReadFile("sample.md")
	if err != nil {
		log.Fatal(err)
	}
	tree, err := markdown.Render(doc, source)
	if err != nil {
		log.Fatal(err)
	}
	content.Base().AppendChild(tree)

	fontSize := 20
	updateStatus := func() {
		status.SetText(fmt.Sprintf("Arrow keys: scroll   L1/R1: font size (%d)", fontSize))
	}
	updateStatus()
	viewport.Listen(fbiw.InputDownEvent, func(event *fbiw.Event) {
		if event.Input.Repeat {
			return
		}
		switch event.Input.Name {
		case sticks.L1:
			fontSize = max(14, fontSize-2)
		case sticks.R1:
			fontSize = min(34, fontSize+2)
		default:
			return
		}
		if err := content.SetProp("font-size", strconv.Itoa(fontSize)); err != nil {
			log.Fatal(err)
		}
		updateStatus()
		event.StopPropagation()
	})
	viewport.Activate()
	app.Run()
}
