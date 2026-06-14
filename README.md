# gonfc — Go demo for linux_libnfc-nci

A small Go port of the C `nfcDemoApp`. All C interaction lives in the
`nfc` package, which wraps every function in `linux_nfc_api.h`. The demo
(`main.go`) is pure Go.

This project is based on
[linux_libnfc-nci (NCI2.0_PN7160)](https://github.com/nehmeroumani/linux_libnfc-nci/tree/NCI2.0_PN7160),
a Linux NFC stack for the PN7160 controller (64-bit and libgpiod support for
Raspberry Pi OS Bookworm).

## Layout

```
goapp/
├── main.go          # CLI demo (poll / read / write) — no cgo
├── nfc/             # Go bindings for linux_nfc_api.h
│   ├── types.go     # types, constants, DiscoveryOptions
│   ├── stack.go     # Open, Close, StartDiscovery, …
│   ├── tag.go       # Tag methods, New*/Parse* records
│   ├── snep.go      # SNEP client/server
│   ├── hce.go       # host card emulation
│   ├── handover.go  # handover
│   ├── llcp.go      # connectionless LLCP
│   ├── manager.go   # NCI config, T4T wired mode
│   ├── callbacks.go # Go callback trampolines
│   ├── bridge.h     # C bridge declarations
│   └── bridge.c     # C callback wiring
└── go.mod
```

## Requirements

- The C library built and installed — see the
  [linux_libnfc-nci README](https://github.com/nehmeroumani/linux_libnfc-nci/tree/NCI2.0_PN7160#install).
- A C toolchain + `pkg-config`.
- Go 1.21+. Build **natively on the target** (e.g. the Pi).

## Installing Go on Raspberry Pi

```bash
sudo apt update
sudo apt install golang-go
go version
```

The package is named `golang-go` (not `golang`) on Raspberry Pi OS.

## Build & run

Build on the Pi (or other target with the library installed):

```bash
go build -o gonfc .
```

If pkg-config can't find `libnfc-nci.pc`:

```bash
PKG_CONFIG_PATH=/usr/local/lib/pkgconfig go build -o gonfc .
```

Run with `sudo` (the stack needs access to the NFC device):

```bash
sudo ./gonfc poll
sudo ./gonfc read
sudo ./gonfc write text "Hello" en
sudo ./gonfc write uri https://www.nxp.com
```

Pass `-debug` before the subcommand to trace startup (Open / StartDiscovery):

```bash
sudo ./gonfc -debug poll
```

## API overview

```go
import "github.com/lulu-box/gonfc/nfc"

nfc.SetTagHandler(nfc.TagHandler{
    OnDiscovered: func(t nfc.Tag) { ... },
    OnRemoved:    func() { ... },
})
nfc.Open()
defer nfc.Close()

nfc.StartDiscovery(nfc.DefaultDiscovery())
defer nfc.StopDiscovery()

info, ok := tag.NDEFInfo()
raw, rt, err := tag.Read()
err = tag.Write(record)

record, _ := nfc.NewTextRecord("en", "Hello")
record, _ := nfc.NewURIRecord("https://example.com")
lang, text, _ := nfc.ParseText(raw)
uri, _ := nfc.ParseURI(raw)
```
