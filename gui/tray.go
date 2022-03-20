package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

func makeTray(a fyne.App, d *dashboard) {
	visible := true
	if desk, ok := a.(desktop.App); ok {
		menu := fyne.NewMenu("EdgeVPN",
			fyne.NewMenuItem("Show/Hide Dashboard", func() {
				if visible {
					d.window.Hide()
					visible = false
				} else {
					d.window.Show()
					visible = true
				}
			}),
			fyne.NewMenuItem("About", func() {
				about(a)
			}),
		)
		desk.SetSystemTrayMenu(menu)
	}
}
