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

## Build & run

```bash
cd goapp
go build -o nfcgo .

sudo ./nfcgo poll
sudo ./nfcgo read
sudo ./nfcgo write text "Hello" en
sudo ./nfcgo write uri  https://www.nxp.com
```

## API overview

```go
import "github.com/nehmeroumani/linux_libnfc-nci/goapp/nfc"

// Stack lifecycle
nfc.Open()
defer nfc.Close()

// Tag discovery
nfc.SetTagHandler(nfc.TagHandler{
    OnDiscovered: func(t nfc.Tag) { ... },
    OnRemoved:    func() { ... },
})
nfc.StartDiscovery(nfc.DefaultDiscovery())

// Tag operations
info, ok := tag.NDEFInfo()
raw, rt, err := tag.Read()
err = tag.Write(record)

// Build and parse NDEF records
record, _ := nfc.NewTextRecord("en", "Hello")
record, _ := nfc.NewURIRecord("https://example.com")
lang, text, _ := nfc.ParseText(raw)
uri, _ := nfc.ParseURI(raw)
```

If pkg-config can't find `libnfc-nci.pc`:

```bash
PKG_CONFIG_PATH=/usr/local/lib/pkgconfig go build -o nfcgo .
```
