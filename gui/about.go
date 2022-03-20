package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func about(app fyne.App) {
	a := app.NewWindow("About")
	a.SetContent(widget.NewRichTextFromMarkdown(
		`

This GUI was written by Ettore Di Giacinto and is released under GPL-3.0 License.

The source code of the GUI is available at GitHub [https://github.com/mudler/edgevpn-gui](https://github.com/mudler/edgevpn-gui).
`))
	a.CenterOnScreen()
	a.Show()
}
