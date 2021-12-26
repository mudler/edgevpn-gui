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
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"fyne.io/fyne/v2"
)

type parent interface {
	Reload(app fyne.App)
}

type vpn struct {
	Name           string `json:"name"`
	Token          string `json:"token"`
	IP             string `json:"ip"`
	API            bool   `json:"api"`
	APIAddress     string `json:"api_address"`
	Interface      string `json:"interface"`
	RuntimeVersion string `json:"runtime_version"`

	stateDir string

	window fyne.Window
	parent parent
}

func (c *vpn) writeJSON(name string) error {
	if err := c.validate(); err != nil {
		return err
	}
	os.MkdirAll(filepath.Join(stateDir(), name), os.ModePerm)

	if c.RuntimeVersion == "system" {
		c.RuntimeVersion = ""
	}

	dat, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(stateDir(), name, "data"), dat, os.ModePerm)
}

func (c *vpn) loadJSON() *vpn {
	t, _ := ioutil.ReadFile(filepath.Join(c.stateDir, "data"))
	json.Unmarshal(t, c)
	return c
}

func generateToken(app fyne.App, w fyne.Window) string {
	available := availableVersions()
	if !isInstalled("edgevpn") && len(available) == 0 {
		errorWindow(
			fmt.Errorf("Can't generate a new token as edgeVPN is not installed, and no versions were downloaded"), w)
		return ""
	}

	bin := "edgevpn"

	if !isInstalled("edgevpn") {
		v := available[len(available)-1]
		bin = binaryVersion(v)
	}

	token, _ := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s -g -b", bin)).CombinedOutput()
	return string(token)
}

func (c *vpn) validate() error {
	_, _, err := net.ParseCIDR(c.IP)
	if err != nil {

		return err
	}
	return nil
}

func newVPN(p string, parent parent) *vpn {
	v := &vpn{
		parent:   parent,
		stateDir: p,
	}
	return v.loadJSON()
}
