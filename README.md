<h1 align="center">
  <br>
	<img src="https://user-images.githubusercontent.com/2420543/144679248-1f6e4c10-a558-424c-b6f5-b3695269c906.png" width=128
         alt="logo"><br>
    EdgeVPN GUI

<br>
</h1>

<h3 align="center">Create Decentralized private networks </h3>
<p align="center">
  <a href="https://opensource.org/licenses/">
    <img src="https://img.shields.io/badge/licence-GPL3-brightgreen"
         alt="license">
  </a>
  <a href="https://github.com/mudler/edgevpn-gui/issues"><img src="https://img.shields.io/github/issues/mudler/edgevpn-gui"></a>
  <img src="https://img.shields.io/badge/made%20with-Go-blue">
  <img src="https://goreportcard.com/badge/github.com/mudler/edgevpn-gui" alt="go report card" />
</p>

A simple GUI for [EdgeVPN](https://github.com/mudler/edgevpn) built with [fyne](https://github.com/fyne-io/fyne).

# :wrench: Features

- Manage EdgeVPN versions locally from the GUI. No system install needed
- Generate, Export, Import and Add VPN connections
- Start/Stop VPN connections, manage connection details and allows to associate versions of EdgeVPN to specific connections if necessary
- Works in any Desktop environment (GNOME, KDE, etc. ), built with [fyne](https://github.com/fyne-io/fyne). Does not depend on NetworkManager, or any other connection manager

# :camera: Screenshots

Dashboard            |  Connections index
:-------------------------:|:-------------------------:
![edgevpn-gui-2](https://user-images.githubusercontent.com/2420543/147854909-a223a7c1-5caa-4e90-b0ac-0ae04dc0949d.png) | ![edgevpn-3](https://user-images.githubusercontent.com/2420543/147854904-09d96991-8752-421a-a301-8f0bdd9d5542.png)
![edgevpn-gui](https://user-images.githubusercontent.com/2420543/147854907-1e4a4715-3181-4dc2-8bc0-d052b3bf46d3.png) | 

# :running: Installation

Download the pre-compiled bundle from the release page, extract it and run `make install`.

At the moment builds are available only for Linux.

# :ledger: State

This GUI is a work in progress. It is able to manage edgevpn connections so far, but still has few graphical glitches that needs to be fixed.