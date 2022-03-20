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
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/go-vgo/robotgo/clipboard"
	process "github.com/mudler/go-processmanager"
)

func (c *vpn) logs(app fyne.App, processStateDir string) func() {
	return func() {
		w := app.NewWindow("Logs")

		tt := widget.NewLabel("")
		accContent := container.NewVScroll(
			tt,
		)

		tt.Wrapping = fyne.TextWrapWord

		pr := process.New(
			process.WithStateDir(processStateDir),
		)

		tt.Resize(fyne.NewSize(100, 100))
		accContent.Resize(fyne.NewSize(100, 100))

		w.SetContent(accContent)
		w.Resize(fyne.NewSize(640, 480))
		w.Show()

		ss := newTail(tt, accContent)

		c, cancel := context.WithCancel(context.Background())
		tailProcess(c, pr, ss)
		w.SetOnClosed(func() {
			close(ss)
			cancel()
		})
	}
}

func (c *vpn) stopButton(app fyne.App, w fyne.Window) *widget.Button {
	s := widget.NewButtonWithIcon("Stop",
		theme.MediaStopIcon(),
		c.stop(app, w),
	)
	s.Importance = widget.LowImportance
	return s
}

func (c *vpn) tokenClipboardButton(app fyne.App) *widget.Button {
	return widget.NewButtonWithIcon(
		"",
		theme.ContentCopyIcon(),
		c.clipboard(app),
	)
}

func (c *vpn) startButton(app fyne.App, w fyne.Window, i widget.ButtonImportance) *widget.Button {
	s := widget.NewButtonWithIcon("Start",
		theme.MediaPlayIcon(),
		//	resources.GetResource(resources.StartIcon, "start"),
		c.start(app, w),
	)
	s.Importance = i
	return s
}

func (c *vpn) start(app fyne.App, w fyne.Window) func() {
	var apiCmd string
	if c.API {
		apiCmd = fmt.Sprintf("--api --api-listen %s", c.APIAddress)
	}

	processStateDir := filepath.Join(c.stateDir, "vpn")

	return func() {
		bin := "edgevpn"
		if c.RuntimeVersion != "" {
			vers := availableVersions()
			match := false
			for _, v := range vers {
				if v == c.RuntimeVersion {
					match = true
				}
			}
			if !match {
				errorWindow(fmt.Errorf("No version found for '%s'", c.RuntimeVersion), w)
			} else {
				bin = binaryVersion(c.RuntimeVersion)
			}
		}

		if c.RuntimeVersion == "" && !isInstalled("edgevpn") {
			errorWindow(fmt.Errorf("edgeVPN is not installed and no versions were downloaded"), w)
			return
		}

		os.MkdirAll(processStateDir, os.ModePerm)
		vpnP := process.New(
			process.WithName("/usr/bin/pkexec"),
			process.WithArgs("/bin/sh", "-c",
				fmt.Sprintf(
					"EDGEVPNTOKEN=%s %s --address %s --interface %s %s",
					c.Token,
					bin,
					c.IP,
					c.Interface,
					apiCmd,
				),
			),
			process.WithStateDir(processStateDir),
		)
		err := vpnP.Run()
		if err != nil {
			errorWindow(err, w)
			vpnP.Stop()
			os.RemoveAll(processStateDir)
		}

		go func() {
			time.Sleep(2 * time.Second)
			c.parent.Reload(app)
			if w != nil {
				c.showDetails(w, app)
			}
			if vpnP.IsAlive() {
				app.SendNotification(
					fyne.NewNotification(
						"connection successful",
						fmt.Sprintf("Network '%s' started on interface '%s'", c.Name, c.Interface)))
			} else {
				app.SendNotification(
					fyne.NewNotification(
						"connection failed",
						fmt.Sprintf("failed starting VPN '%s'", c.Name),
					))
				vpnP.Stop()
				os.RemoveAll(processStateDir)
			}
		}()
	}
}

func (c *vpn) stop(app fyne.App, w fyne.Window) func() {
	processStateDir := filepath.Join(c.stateDir, "vpn")
	return func() {
		dialog.NewConfirm(
			"Stop",
			"Are you sure you want to stop the VPN?",
			func(b bool) {
				if b {
					vpnP := process.New(
						process.WithStateDir(processStateDir),
					)
					vpnP.Stop()
					if vpnP.IsAlive() {
						exec.Command("/usr/bin/pkexec", "kill", "-9", vpnP.PID).CombinedOutput()
					}
					os.RemoveAll(processStateDir)
					go func() {
						time.Sleep(2 * time.Second)
						c.parent.Reload(app)
					}()
				}
			}, w,
		).Show()
	}
}

func (c *vpn) logButton(app fyne.App) *widget.Button {
	b := widget.NewButtonWithIcon("Logs",
		theme.FileTextIcon(),
		c.logs(app, c.processDir()),
	)
	b.Importance = widget.LowImportance
	return b
}

func (c *vpn) apiLink() *widget.Button {
	a := c.APIAddress
	if strings.HasPrefix(a, ":") {
		a = fmt.Sprintf("http://127.0.0.1%s", a)
	}
	if u, err := url.Parse(a); err == nil {
		b := widget.NewButtonWithIcon("API",
			theme.ComputerIcon(),
			func() {
				fyne.CurrentApp().OpenURL(u)
			},
		)
		b.Importance = widget.LowImportance
		return b
		//	return widget.NewHyperlink("API", u)
	}

	return nil
}

func (c *vpn) deleteButton(app fyne.App, w fyne.Window) *widget.Button {
	return widget.NewButtonWithIcon(
		"Delete",
		theme.DeleteIcon(),
		c.delete(app, w),
	)
}

func (c *vpn) delete(app fyne.App, p fyne.Window) func() {
	return func() {
		dialog.NewConfirm(
			"Delete",
			"Are you sure you want to delete the VPN?",
			func(b bool) {
				if b {
					os.RemoveAll(c.stateDir)
					c.parent.Reload(app)
					p.Close()
				}
			}, p,
		).Show()
	}
}

func (c *vpn) clipboard(app fyne.App) func() {
	return func() {
		clipboard.WriteAll(c.Token)
		app.SendNotification(fyne.NewNotification("info", "Token copied to clipboard"))
	}
}

func (c *vpn) update(dat vpn, app fyne.App, w fyne.Window) func() {
	return func() {
		dialog.NewConfirm(
			"Update",
			"Are you sure you want to update the VPN?",
			func(b bool) {
				if b {
					if err := dat.writeJSON(dat.Name); err != nil {
						errorWindow(err, w)
						return
					}

					c.parent.Reload(app)
					c.showDetails(w, app)
				}
			}, w,
		).Show()
	}
}
