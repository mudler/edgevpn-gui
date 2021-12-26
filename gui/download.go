package gui

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/0xAX/notificator"
	"github.com/cavaliercoder/grab"
	"github.com/mholt/archiver/v3"
	"github.com/otiai10/copy"
)

const repoSlug = "mudler/edgevpn"

func Download(app fyne.App, url, dst string) string {
	// create client
	client := grab.NewClient()
	req, _ := grab.NewRequest(dst, url)

	w := app.NewWindow(fmt.Sprintf("Download %s", url))

	tt := widget.NewLabel(fmt.Sprintf("Downloading %s in %s", url, dst))
	progress := widget.NewProgressBar()
	w.SetContent(container.NewVBox(tt, progress))

	//c.window.Resize(fyne.NewSize(200, 300))
	w.Show()

	// start download
	resp := client.Do(req)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			progress.SetValue(resp.Progress())
		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		errorWindow(err, w)
	}
	w.Close()
	notify.Push("info", fmt.Sprintf("Download saved to ./%v", resp.Filename), "", notificator.UR_NORMAL)
	return resp.Filename
}

func DownloadEdgeVPN(url, dstfile string, app fyne.App, w fyne.Window) {
	tmpdir, err := ioutil.TempDir("", "edgevpn-gui")
	if err != nil {
		errorWindow(err, w)
		return
	}
	defer os.RemoveAll(tmpdir)

	dst := Download(app, url, tmpdir)

	err = archiver.Unarchive(dst, tmpdir)
	if err != nil {
		errorWindow(err, w)
		return
	}

	err = copy.Copy(filepath.Join(tmpdir, "edgevpn"), dstfile)
	if err != nil {
		errorWindow(err, w)
		return
	}
}

func DownloadAndInstall(app fyne.App, w fyne.Window, version string) {
	f := newReleaseFinder(context.Background(), "")

	rel, ass, err := f.find(repoSlug, version)
	if err != nil {
		errorWindow(err, w)
		return
	}
	fmt.Println("Found", *ass.Name, *ass.BrowserDownloadURL)
	DownloadEdgeVPN(*ass.BrowserDownloadURL, binaryVersion(rel.GetName()), app, w)
}

type VersionsManager struct {
	window fyne.Window
}

func inSlice(s string, ss []string) bool {
	for _, f := range ss {
		if f == s {
			return true
		}
	}
	return false
}

func (m *VersionsManager) showUI(app fyne.App) {

	if m.window == nil {
		m.window = app.NewWindow("Version manager")
	}
	f := newReleaseFinder(context.Background(), "")
	versions, _ := f.findAll(repoSlug)

	cards := []fyne.CanvasObject{}

	available := availableVersions()

	for i := range versions {
		v := versions[i]
		var b *widget.Button
		if inSlice(v, available) {
			b = widget.NewButton(
				"Remove",
				func() {
					dialog.NewConfirm(
						"Delete",
						fmt.Sprintf("Are you sure you want to delete version %v?",
							v),
						func(b bool) {
							if b {
								os.RemoveAll(binaryVersion(v))
								m.showUI(app)
							}
						}, m.window).Show()
				},
			)
		} else {
			b = widget.NewButton(
				"Download",
				func() {
					DownloadAndInstall(app, m.window, v)
					m.showUI(app)
				},
			)
		}

		cards = append(cards,
			widget.NewCard(v, "",
				b,
			),
		)
	}

	m.window.SetContent(
		container.NewBorder(
			nil,
			nil,
			nil,
			nil,
			container.NewVScroll(container.NewVBox(cards...)),
		),
	)

	m.window.Resize(fyne.NewSize(300, 300))
	m.window.Show()
}
