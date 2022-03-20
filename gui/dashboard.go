// Copyright © 2021 Ettore Di Giacinto <mudler@mocaccino.org>
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, see <http://www.gnu.org/licenses/>.

package gui

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/mudler/edgevpn-gui/resources"
)

type dashboard struct {
	window fyne.Window
}

const welcomeMessage string = `
# Welcome

Welcome to the EdgeVPN gui. This is a simple utility to control EdgeVPN instances in your system.

This application can be safely closed. VPN connection will keep running in the background.
`

func (c *dashboard) Reload(app fyne.App) {
	state := stateDir()
	os.MkdirAll(state, os.ModePerm)

	readVpn := func() (foundVpns []*vpn) {
		files, err := ioutil.ReadDir(state)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			if f.IsDir() {
				if _, err := os.Stat(filepath.Join(state, f.Name(), "data")); err == nil {
					foundVpns = append(foundVpns, newVPN(filepath.Join(state, f.Name()), c))
				}
			}
		}
		return
	}

	genCards := func() (cards []fyne.CanvasObject) {
		foundVpns := readVpn()
		for _, v := range foundVpns {
			cards = append(cards, v.card(app, c.window))
		}
		return
	}

	welcomeText := widget.NewRichTextFromMarkdown(welcomeMessage)

	addVPN := func() *widget.Button {
		b := widget.NewButtonWithIcon("Add VPN",
			theme.ContentAddIcon(),
			func() {
				newVPN("", c).generateUI(app, false)
			})
		b.Importance = widget.HighImportance
		return b
	}

	generateVPN := func() *widget.Button {
		return widget.NewButtonWithIcon("Generate new VPN",
			theme.DocumentCreateIcon(),
			func() {
				newVPN("", c).generateUI(app, true)
			})
	}

	importVPN := func() *widget.Button {
		return widget.NewButtonWithIcon("Import new VPN",
			theme.DownloadIcon(),
			func() {
				d := dialog.NewFileOpen(
					func(f fyne.URIReadCloser, e error) {
						if e != nil {
							errorWindow(e, c.window)
							return
						}

						if f == nil {
							return
						}
						v := &vpn{}
						err := json.NewDecoder(f).Decode(v)
						if err != nil {
							errorWindow(e, c.window)
							return
						}
						f.Close()

						err = v.writeJSON(v.Name)
						if err != nil {
							errorWindow(e, c.window)
							return
						}
						c.Reload(app)
						app.SendNotification(fyne.NewNotification("info", "File saved"))
					}, c.window)
				d.Show()
			})
	}
	downloadEdgeVPN := func() *widget.Button {
		return widget.NewButtonWithIcon("Manage EdgeVPN versions",
			resources.GetResource(resources.EdgeVPNIcon, "manage"),
			func() {
				m := &VersionsManager{}
				m.showUI(app)
			})
	}
	noVPN := widget.NewLabel("No VPN found in the system!")

	aboutButton := widget.NewButtonWithIcon("About",
		theme.InfoIcon(),
		func() {
			about(app)
		})

	if len(readVpn()) == 0 {
		c.window.SetContent(
			container.NewBorder(
				welcomeText,
				nil,
				nil,
				nil,
				container.NewCenter(container.NewGridWithColumns(
					1,
					noVPN, addVPN(), generateVPN(), importVPN(), downloadEdgeVPN(),
					aboutButton,
				)),
			),
		)
		c.window.Resize(fyne.NewSize(640, 640))
	} else {
		cards := genCards()
		grid := container.NewAdaptiveGrid(3, cards...)
		acc := widget.NewAccordion(
			widget.NewAccordionItem(
				"Create, Import ...",
				container.NewGridWithColumns(
					4,
					addVPN(), generateVPN(), downloadEdgeVPN(), importVPN(),
				),
			),
		)
		c.window.Resize(fyne.NewSize(640, 640))

		//c.window.Resize(grid.Layout.MinSize(append(cards, acc, welcomeText, layout.NewSpacer())))
		b := container.NewVScroll(container.NewVBox(grid))

		c.window.SetContent(
			container.NewBorder(
				welcomeText,
				acc,
				nil,
				nil,
				b,
			),
		)
	}

}

func (c *dashboard) loadUI(app fyne.App) {
	c.window = app.NewWindow("EdgeVPN")
	c.Reload(app)
	c.window.SetPadded(true)
	c.window.CenterOnScreen()
	c.window.Show()

	if !isInstalled("edgevpn") {
		if len(availableVersions()) != 0 {
			return
		}
		dialog.NewConfirm(
			"Download",
			"EdgeVPN was not found in the system, proceed to download?",
			func(b bool) {
				if b {
					DownloadAndInstall(app, c.window, "")
				}
			},
			c.window,
		).Show()
	}
}

func newDashboard() *dashboard {
	return &dashboard{}
}
