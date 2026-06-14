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
	"syscall"
	"time"

	"github.com/lulu-box/gonfc/nfc"
)

const (
	ndefDetectWait   = 800 * time.Millisecond
	ndefPollInterval = 50 * time.Millisecond
	tagEventQueue    = 16
)

var debug = flag.Bool("debug", false, "enable debug logging")

// tagEvent carries a discovery or removal notification from the C callbacks.
type tagEvent struct {
	tag     nfc.Tag
	removed bool
}

var tagEvents = make(chan tagEvent, tagEventQueue)

// pushTagEvent queues a tag event from a C callback thread without blocking the
// NFC stack. If the queue is full it discards the oldest event so the newest
// (e.g. a fresh arrival waitTag is blocked on) is never lost.
func pushTagEvent(ev tagEvent) {
	select {
	case tagEvents <- ev:
		return
	default:
	}
	select {
	case <-tagEvents:
		nfc.Debugf("tag event queue full, dropped oldest event")
	default:
	}
	select {
	case tagEvents <- ev:
	default:
		fmt.Fprintln(os.Stderr, "warning: dropped tag event (queue full)")
	}
}

func main() {
	flag.Parse()
	nfc.SetDebug(*debug)

	// Global flags must precede the subcommand (standard flag.Parse behavior),
	// e.g. "gonfc -debug poll".
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
		OnDiscovered: func(t nfc.Tag) { pushTagEvent(tagEvent{tag: t}) },
		OnRemoved:    func() { pushTagEvent(tagEvent{removed: true}) },
	})

	if err := nfc.Open(); err != nil {
		return err
	}
	defer nfc.Close()

	nfc.StartDiscovery(nfc.DefaultDiscovery())
	defer nfc.StopDiscovery()
	// Deregister the C tag callback before StopDiscovery/Close tear down the
	// stack (defers run LIFO).
	defer nfc.ClearTagHandler()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return fn(ctx)
}

func poll(ctx context.Context) error {
	fmt.Println("Polling for tags (Ctrl-C to stop)...")

	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-tagEvents:
				if !ok {
					return
				}
				if ev.removed {
					fmt.Println("  tag removed")
					continue
				}
				printTag(ev.tag)
				if err := printNDEF(ctx, ev.tag); err != nil {
					// printNDEF only returns context errors; anything else means
					// the context was cancelled, so stop the worker.
					if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
						return
					}
					fmt.Fprintf(os.Stderr, "  poll: %v\n", err)
				}
			}
		}
	}()

	<-ctx.Done()
	<-workerDone
	return nil
}

func readOne(ctx context.Context) error {
	t, err := waitTag(ctx)
	if err != nil {
		return err
	}
	printTag(t)
	if err := printNDEF(ctx, t); err != nil {
		return err
	}
	return nil
}

func writeOne(ctx context.Context, msg []byte) error {
	t, err := waitTag(ctx)
	if err != nil {
		return err
	}
	printTag(t)
	if err := writeNDEF(ctx, t, msg); err != nil {
		return err
	}
	fmt.Println("  write OK")
	_ = printNDEF(ctx, t)
	return nil
}

// waitTag returns the next discovered tag, skipping removal notifications.
func waitTag(ctx context.Context) (nfc.Tag, error) {
	fmt.Println("Waiting for a tag...")
	for {
		select {
		case <-ctx.Done():
			return nfc.Tag{}, ctx.Err()
		case ev := <-tagEvents:
			if ev.removed {
				continue
			}
			return ev.tag, nil
		}
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

func sleepOrDone(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// tagReconnect reselects slow tags before NDEF operations.
func tagReconnect(t nfc.Tag) error {
	if !t.SlowNDEFDetection() {
		return nil
	}
	if err := t.Reconnect(); err != nil {
		if nfc.IsTagGone(err) {
			return fmt.Errorf("tag no longer in field: %w", err)
		}
		return err
	}
	return nil
}

// waitNDEFInfo waits for async NDEF detection on slow tag types.
func waitNDEFInfo(ctx context.Context, t nfc.Tag) (nfc.NDEFInfo, bool, error) {
	if !t.SlowNDEFDetection() {
		info, ok := t.NDEFInfo()
		return info, ok, nil
	}
	deadline := time.Now().Add(ndefDetectWait)
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return nfc.NDEFInfo{}, false, err
		}
		if info, ok := t.NDEFInfo(); ok {
			return info, true, nil
		}
		if err := sleepOrDone(ctx, ndefPollInterval); err != nil {
			return nfc.NDEFInfo{}, false, err
		}
	}
	info, ok := t.NDEFInfo()
	return info, ok, nil
}

// writeNDEF mirrors nfcDemoApp: wait for detection, format if needed, then write.
func writeNDEF(ctx context.Context, t nfc.Tag, msg []byte) error {
	if err := tagReconnect(t); err != nil {
		return err
	}

	info, hasNDEF, err := waitNDEFInfo(ctx, t)
	if err != nil {
		return err
	}
	if hasNDEF && !info.Writable {
		return errors.New("tag NDEF is read-only")
	}
	if hasNDEF && info.MaxLength > 0 && uint(len(msg)) > info.MaxLength {
		return fmt.Errorf("message too large: %d bytes, tag holds %d", len(msg), info.MaxLength)
	}
	if !hasNDEF {
		if t.Technology != nfc.TargetMifareClassic && !t.Formatable() {
			return errors.New("tag is not NDEF-formatable")
		}
		fmt.Println("  formatting tag for NDEF...")
		if err := t.Format(); err != nil {
			return fmt.Errorf("format failed: %w", err)
		}
		if err := tagReconnect(t); err != nil {
			return err
		}
	}

	if err := t.Write(msg); err != nil {
		if canRetryWrite(t, err) {
			fmt.Println("  write failed, trying format then retry...")
			if fmtErr := t.Format(); fmtErr == nil {
				if recErr := tagReconnect(t); recErr != nil {
					return recErr
				}
				if err2 := t.Write(msg); err2 != nil {
					return writeFailedErr(t, err2)
				}
				return nil
			}
		}
		return writeFailedErr(t, err)
	}
	return nil
}

func canRetryWrite(t nfc.Tag, err error) bool {
	if !nfc.IsStatus(err, nfc.StatusFailed) {
		return false
	}
	if t.Technology == nfc.TargetMifareClassic {
		return true
	}
	return t.Formatable()
}

func writeFailedErr(t nfc.Tag, err error) error {
	if t.Technology == nfc.TargetMifareClassic && nfc.IsStatus(err, nfc.StatusFailed) {
		return fmt.Errorf("%w — check MIFARE_READER_ENABLE=0x01 in libnfc-nxp.conf; tag may use non-default keys", err)
	}
	return err
}

// readNDEF waits for NDEF detection on slow tag types and retries transient failures.
func readNDEF(ctx context.Context, t nfc.Tag) ([]byte, nfc.RecordType, error) {
	if !t.SlowNDEFDetection() {
		return t.Read()
	}

	deadline := time.Now().Add(ndefDetectWait)
	var lastReadErr error

	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return nil, nfc.RecordOther, err
		}
		raw, rt, err := t.Read()
		if err == nil {
			return raw, rt, nil
		}
		if errors.Is(err, nfc.ErrNotNDEF) {
			if err := sleepOrDone(ctx, ndefPollInterval); err != nil {
				return nil, nfc.RecordOther, err
			}
			continue
		}
		if nfc.IsTagGone(err) {
			return nil, nfc.RecordOther, err
		}
		lastReadErr = err
		if recErr := tagReconnect(t); recErr != nil {
			if nfc.IsTagGone(recErr) {
				return nil, nfc.RecordOther, recErr
			}
			lastReadErr = recErr
		}
		if err := sleepOrDone(ctx, ndefPollInterval); err != nil {
			return nil, nfc.RecordOther, err
		}
	}

	raw, rt, err := t.Read()
	if err == nil {
		return raw, rt, nil
	}
	if errors.Is(err, nfc.ErrNotNDEF) {
		return nil, rt, err
	}
	if nfc.IsTagGone(err) {
		return nil, nfc.RecordOther, err
	}
	if lastReadErr != nil {
		return nil, nfc.RecordOther, lastReadErr
	}
	return nil, rt, err
}

func printNDEF(ctx context.Context, t nfc.Tag) error {
	raw, rt, err := readNDEF(ctx, t)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		if errors.Is(err, nfc.ErrNotNDEF) {
			fmt.Println("  NDEF: no")
			if hint := t.WriteHint(); hint != "" {
				fmt.Printf("  (%s)\n", hint)
			}
			return nil
		}
		fmt.Printf("  NDEF: read failed: %v\n", err)
		return nil
	}
	if len(raw) == 0 {
		fmt.Println("  NDEF: empty")
		return nil
	}

	switch rt {
	case nfc.RecordText:
		lang, text, err := nfc.ParseText(raw)
		if err != nil {
			fmt.Printf("  NDEF: decode failed: %v\n", err)
			return nil
		}
		fmt.Printf("  NDEF text [%s]: %s\n", lang, text)
	case nfc.RecordURI:
		uri, err := nfc.ParseURI(raw)
		if err != nil {
			fmt.Printf("  NDEF: decode failed: %v\n", err)
			return nil
		}
		fmt.Printf("  NDEF uri: %s\n", uri)
	default:
		fmt.Printf("  NDEF raw: %s\n", strings.ToUpper(hex.EncodeToString(raw)))
	}
	return nil
}
