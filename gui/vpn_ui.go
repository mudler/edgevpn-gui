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
	"fmt"
	"io"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/0xAX/notificator"
	process "github.com/mudler/go-processmanager"
)

func (c *vpn) isAlive() bool {
	return process.New(process.WithStateDir(c.processDir())).IsAlive()
}

func (c *vpn) processDir() string {
	return filepath.Join(c.stateDir, "vpn")
}

func (c *vpn) showDetails(w fyne.Window, app fyne.App) {

	c.loadJSON()

	name := widget.NewEntry()
	name.SetText(c.Name)
	name.Disable()
	ipE := widget.NewEntry()
	ipE.SetText(c.IP)
	token := widget.NewPasswordEntry()
	token.SetText(c.Token)

	runtimeVersion := widget.NewSelect(selectableVersions(), func(s string) {})
	runtimeForm := widget.NewFormItem("Runtime version", runtimeVersion)
	selected := c.RuntimeVersion
	if selected == "" {
		selected = "system"
	}
	runtimeVersion.SetSelected(selected)

	v := widget.NewFormItem("VPN Name", name)
	ip := widget.NewFormItem("IP", ipE)

	s := container.NewHSplit(token, c.tokenClipboardButton())
	s.SetOffset(0.9)
	tokenW := widget.NewFormItem("Token", s)

	saveDialog := dialog.NewFileSave(
		func(f fyne.URIWriteCloser, e error) {
			if e != nil {
				errorWindow(e, w)
				return
			}
			if f == nil {
				return
			}
			s, err := os.Open(filepath.Join(c.stateDir, "data"))
			if err != nil {
				errorWindow(err, w)
				return
			}
			io.Copy(f, s)
			f.Close()
			notify.Push("info", "File saved", "", notificator.UR_NORMAL)
		}, w)

	iff := widget.NewEntry()
	iff.SetText(c.Interface)
	apiText := widget.NewEntry()
	apiText.SetText(c.APIAddress)

	apiL := widget.NewFormItem("API Listen Address", apiText)
	apiB := widget.NewCheck("API", func(c bool) {
		if c {
			apiL.Widget.Show()
		} else {
			apiL.Widget.Hide()
		}
	})

	if !c.API {
		apiB.SetChecked(false)
	} else {
		apiB.SetChecked(true)
	}

	ifw := widget.NewFormItem("Interface", iff)
	api := widget.NewFormItem("API", apiB)

	form := widget.NewForm(
		v, ip, ifw, api, apiL, runtimeForm, tokenW,
	)

	buttons := []fyne.CanvasObject{
		c.deleteButton(app, w),
		widget.NewButtonWithIcon("Save",
			theme.DocumentSaveIcon(),
			func() {
				c.update(vpn{
					Token:          token.Text,
					IP:             ipE.Text,
					Name:           name.Text,
					Interface:      iff.Text,
					API:            apiB.Checked,
					APIAddress:     apiText.Text,
					RuntimeVersion: runtimeVersion.Selected,
				}, app, w)()
			},
		),
		widget.NewButtonWithIcon("Export",
			theme.UploadIcon(),
			func() {
				saveDialog.Show()
			}),
	}

	// if c.isAlive() {
	// 	buttons = append(buttons,
	// 		//		c.stopButton(app, w),
	// 		widget.NewButtonWithIcon("Open Logs",
	// 			theme.ListIcon(),
	// 			c.logs(app, processStateDir),
	// 		),
	// 	)
	// } else {
	// 	buttons = append(buttons) //		c.startButton(app, w, widget.HighImportance),

	// }

	// if c.isAlive() && c.API {
	// 	if l := c.apiLink(); l != nil {
	// 		buttons = append(buttons, l)
	// 	}
	// }

	w.SetContent(container.NewBorder(
		container.NewGridWithColumns(
			4,
			buttons...,
		),
		nil,
		nil,
		nil,
		container.NewGridWithColumns(1, form),
	))
	// w.SetContent(container.NewBorder(
	// 	nil,
	// 	container.NewGridWithColumns(4,
	// 		buttons...),
	// 	nil,
	// 	nil,
	// 	container.NewGridWithColumns(1, form),
	// ))
}

func (c *vpn) showUI(app fyne.App) {

	c.window = app.NewWindow(fmt.Sprintf("VPN %s", c.Name))

	c.showDetails(c.window, app)

	//c.window.Resize(fyne.NewSize(200, 300))
	c.window.Show()
}

func selectableVersions() []string {
	if isInstalled("edgevpn") {
		return append([]string{"system"}, availableVersions()...)
	}
	return availableVersions()
}

func (c *vpn) generateUI(app fyne.App, genToken bool) {
	c.window = app.NewWindow("VPN")
	name := widget.NewEntry()
	ipE := widget.NewEntry()
	v := widget.NewFormItem("VPN Name", name)
	ip := widget.NewFormItem("IP", ipE)
	iff := widget.NewEntry()

	runtimeVersion := widget.NewSelect(selectableVersions(), func(string) {})
	runtimeForm := widget.NewFormItem("Runtime version", runtimeVersion)

	apiText := widget.NewEntry()
	apiL := widget.NewFormItem("API Listen Address", apiText)
	apiText.Text = ":8080"

	apiL.Widget.Hide()
	apiB := widget.NewCheck("API", func(c bool) {
		if c {

			apiL.Widget.Show()
		} else {
			apiL.Widget.Hide()
		}
	})

	ifw := widget.NewFormItem("Interface", iff)
	api := widget.NewFormItem("API", apiB)

	iff.Text = "edgevpn0"
	token := widget.NewPasswordEntry()
	tk := widget.NewFormItem("Token", token)

	if genToken {
		token.Disable()
		token.SetText(generateToken(app, c.window))
	}

	form := widget.NewForm(
		v, ip, tk, ifw, api, apiL, runtimeForm,
	)

	form.OnCancel = func() {
		c.window.Close()
	}
	form.OnSubmit = func() {
		d := vpn{
			Token:          token.Text,
			IP:             ipE.Text,
			Name:           name.Text,
			Interface:      iff.Text,
			API:            apiB.Checked,
			APIAddress:     apiText.Text,
			RuntimeVersion: runtimeVersion.Selected,
		}
		if err := d.writeJSON(name.Text); err != nil {
			errorWindow(err, c.window)
			return
		}

		c.parent.Reload(app)
		c.window.Close()
	}

	c.window.SetContent(form)

	c.window.Resize(fyne.NewSize(300, 300))
	c.window.Show()
}

func (c *vpn) card(app fyne.App, w fyne.Window) fyne.CanvasObject {
	var objs []fyne.CanvasObject

	info := widget.NewButtonWithIcon("",
		theme.InfoIcon(),
		func() {
			c.showUI(app)
		},
	)
	info.Importance = widget.LowImportance

	objs = append(objs,
		info,
	)

	if c.isAlive() {
		objs = append(objs,
			c.stopButton(app, w),
			c.logButton(app),
		)
		if c.API {
			if l := c.apiLink(); l != nil {
				objs = append(objs, l)
			}
		}
	} else {
		objs = append(objs,
			c.startButton(app, nil, widget.LowImportance),
		)
	}

	return widget.NewCard(
		c.Name, c.IP,
		container.NewHBox(
			objs...,
		))
}
