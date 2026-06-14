// Command gonfc is a small Go demo for the linux_libnfc-nci stack, mirroring the
// poll / read / write flows of the C nfcDemoApp.
//
//	gonfc poll                  detect tags continuously
//	gonfc read                  read NDEF from the next tag, then exit
//	gonfc write text <text> [lang]
//	gonfc write uri  <uri>
package main

import (
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/lulu-box/gonfc/nfc"
)

var debug = flag.Bool("debug", false, "enable debug logging")

var (
	discovered = make(chan nfc.Tag, 8)
	removed    = make(chan struct{}, 8)
)

func main() {
	flag.Parse()
	nfc.SetDebug(*debug)

	args := flag.Args()
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}

	var err error
	switch args[0] {
	case "poll":
		err = run(poll)
	case "read":
		err = run(readOne)
	case "write":
		var msg []byte
		if msg, err = buildRecord(args[1:]); err == nil {
			err = run(func(ctx context.Context) error { return writeOne(ctx, msg) })
		}
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Print(`gonfc - Go demo for linux_libnfc-nci

  gonfc [-debug] poll                   detect tags continuously
  gonfc [-debug] read                   read NDEF from the next tag, then exit
  gonfc [-debug] write text <text> [lang]
  gonfc [-debug] write uri  <uri>
`)
}

func run(fn func(context.Context) error) error {
	nfc.SetTagHandler(nfc.TagHandler{
		OnDiscovered: func(t nfc.Tag) {
			select {
			case discovered <- t:
			default:
			}
		},
		OnRemoved: func() {
			select {
			case removed <- struct{}{}:
			default:
			}
		},
	})
	defer nfc.ClearTagHandler()

	if err := nfc.Open(); err != nil {
		return err
	}
	defer nfc.Close()

	nfc.StartDiscovery(nfc.DefaultDiscovery())
	defer nfc.StopDiscovery()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return fn(ctx)
}

func poll(ctx context.Context) error {
	fmt.Println("Polling for tags (Ctrl-C to stop)...")
	for {
		select {
		case <-ctx.Done():
			return nil
		case t := <-discovered:
			printTag(t)
			printNDEF(t)
		case <-removed:
			fmt.Println("  tag removed")
		}
	}
}

func readOne(ctx context.Context) error {
	t, err := waitTag(ctx)
	if err != nil {
		return err
	}
	printTag(t)
	printNDEF(t)
	return nil
}

func writeOne(ctx context.Context, msg []byte) error {
	t, err := waitTag(ctx)
	if err != nil {
		return err
	}
	printTag(t)
	if err := t.Write(msg); err != nil {
		return err
	}
	fmt.Println("  write OK")
	printNDEF(t)
	return nil
}

// waitTag returns the next discovered tag. Because discovered is buffered, the
// returned tag may have left the field by the time it is used; tag operations
// will then fail and should be handled accordingly.
func waitTag(ctx context.Context) (nfc.Tag, error) {
	fmt.Println("Waiting for a tag...")
	select {
	case <-ctx.Done():
		return nfc.Tag{}, errors.New("cancelled")
	case t := <-discovered:
		return t, nil
	}
}

func buildRecord(args []string) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("usage: write text <text> [lang] | write uri <uri>")
	}
	switch args[0] {
	case "text":
		lang := "en"
		if len(args) >= 3 {
			lang = args[2]
		}
		return nfc.NewTextRecord(lang, args[1])
	case "uri":
		return nfc.NewURIRecord(args[1])
	default:
		return nil, fmt.Errorf("unknown write type %q (text|uri)", args[0])
	}
}

func printTag(t nfc.Tag) {
	fmt.Printf("\nTag found: %s uid=%s\n", tagTypeName(t.Technology), strings.ToUpper(hex.EncodeToString(t.UID)))
}

func tagTypeName(tech uint) string {
	switch tech {
	case nfc.TargetISO14443_3A:
		return "ISO14443-3A"
	case nfc.TargetISO14443_3B:
		return "ISO14443-3B"
	case nfc.TargetISO14443_4:
		return "ISO14443-4"
	case nfc.TargetFelica:
		return "FeliCa"
	case nfc.TargetISO15693:
		return "ISO15693"
	case nfc.TargetNDEF:
		return "NDEF"
	case nfc.TargetNDEFFormat:
		return "NDEF formatable"
	case nfc.TargetMifareClassic:
		return "Mifare Classic"
	case nfc.TargetMifareUL:
		return "Mifare Ultralight"
	case nfc.TargetKovioBarcode:
		return "Kovio barcode"
	case nfc.TargetISO14443_3A3B:
		return "ISO14443-3A/3B"
	default:
		return fmt.Sprintf("type %d", tech)
	}
}

// readNDEF waits for NDEF detection on slow tag types (notably Mifare Classic)
// and retries transient read failures with reconnect.
func readNDEF(t nfc.Tag) ([]byte, nfc.RecordType, error) {
	if t.Technology != nfc.TargetMifareClassic {
		return t.Read()
	}

	deadline := time.Now().Add(800 * time.Millisecond)
	var lastReadErr error

	for time.Now().Before(deadline) {
		raw, rt, err := t.Read()
		if err == nil {
			return raw, rt, nil
		}
		if errors.Is(err, nfc.ErrNotNDEF) {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		lastReadErr = err
		_ = t.Reconnect()
		time.Sleep(50 * time.Millisecond)
	}

	raw, rt, err := t.Read()
	if err == nil {
		return raw, rt, nil
	}
	if lastReadErr != nil {
		return nil, nfc.RecordOther, lastReadErr
	}
	return nil, rt, err
}

func printNDEF(t nfc.Tag) {
	raw, rt, err := readNDEF(t)
	if errors.Is(err, nfc.ErrNotNDEF) {
		fmt.Println("  NDEF: no")
		if t.Formatable() {
			fmt.Println("  (formatable — try: gonfc write text \"Hello\")")
		}
		return
	}
	if err != nil {
		fmt.Printf("  NDEF: read failed: %v\n", err)
		return
	}
	if len(raw) == 0 {
		fmt.Println("  NDEF: empty")
		return
	}

	switch rt {
	case nfc.RecordText:
		lang, text, err := nfc.ParseText(raw)
		if err != nil {
			fmt.Printf("  NDEF: decode failed: %v\n", err)
			return
		}
		fmt.Printf("  NDEF text [%s]: %s\n", lang, text)
	case nfc.RecordURI:
		uri, err := nfc.ParseURI(raw)
		if err != nil {
			fmt.Printf("  NDEF: decode failed: %v\n", err)
			return
		}
		fmt.Printf("  NDEF uri: %s\n", uri)
	default:
		fmt.Printf("  NDEF raw: %s\n", strings.ToUpper(hex.EncodeToString(raw)))
	}
}
