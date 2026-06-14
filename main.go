// Command nfcgo is a small Go demo for the linux_libnfc-nci stack, mirroring the
// poll / read / write flows of the C nfcDemoApp.
//
//	nfcgo poll                  detect tags continuously
//	nfcgo read                  read NDEF from the next tag, then exit
//	nfcgo write text <text> [lang]
//	nfcgo write uri  <uri>
package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/lulu-box/gonfc/nfc"
)

var (
	discovered = make(chan nfc.Tag, 8)
	removed    = make(chan struct{}, 8)
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "poll":
		err = run(poll)
	case "read":
		err = run(readOne)
	case "write":
		var msg []byte
		if msg, err = buildRecord(os.Args[2:]); err == nil {
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
	fmt.Print(`nfcgo - Go demo for linux_libnfc-nci

  nfcgo poll                   detect tags continuously
  nfcgo read                   read NDEF from the next tag, then exit
  nfcgo write text <text> [lang]
  nfcgo write uri  <uri>
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
	fmt.Printf("\nTag found: tech=%d uid=%s\n", t.Technology, strings.ToUpper(hex.EncodeToString(t.UID)))
}

func printNDEF(t nfc.Tag) {
	raw, rt, err := t.Read()
	if errors.Is(err, nfc.ErrNotNDEF) {
		fmt.Println("  NDEF: no")
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
