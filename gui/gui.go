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
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	process "github.com/mudler/go-processmanager"
	"github.com/nxadm/tail"
)

//go:generate fyne bundle -package gui -o data.go ../Icon.png

func Run() {

	app := app.New()
	app.Settings().SetTheme(theme.LightTheme())
	app.SetIcon(resourceIconPng)

	c := newDashboard()
	c.loadUI(app)
	makeTray(app, c)
	app.Run()
}

func errorWindow(err error, w fyne.Window) {
	dialog.NewError(err, w).Show()
}

func stateDir() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	return filepath.Join(dirname, ".edgevpn")
}

func tailProcess(ctx context.Context, pr *process.Process, c chan string) {

	go func() {
		t, _ := tail.TailFile(pr.StdoutPath(), tail.Config{Follow: true})

		for {
			select {
			case line := <-t.Lines: // Every 100ms increate number of ticks and update UI
				c <- line.Text
			case <-ctx.Done():
				return
			}
		}

	}()
	go func() {
		t, _ := tail.TailFile(pr.StderrPath(), tail.Config{Follow: true})
		for {
			select {
			case line := <-t.Lines: // Every 100ms increate number of ticks and update UI
				c <- line.Text
			case <-ctx.Done():
				return
			}
		}
	}()
}

func newTail(t *widget.Label, w *container.Scroll) chan string {
	s := make(chan string)

	go func() {
		for ss := range s {
			curr := t.Text
			// Prevent too much log to be scrolled
			if len(curr) > 1000 {
				curr = ""
			}
			curr += ss + "\n"
			t.SetText(curr)
			w.ScrollToBottom()
		}
	}()
	return s
}
