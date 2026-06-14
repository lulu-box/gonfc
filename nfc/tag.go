package nfc

/*
#include <stdlib.h>
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

// ErrNotNDEF is returned by Read when the tag does not hold an NDEF message.
var ErrNotNDEF = errors.New("nfc: tag is not NDEF")

// NDEFInfo reports whether the tag holds NDEF and its capabilities.
func (t Tag) NDEFInfo() (NDEFInfo, bool) {
	var ci C.ndef_info_t
	if C.nfcTag_isNdef(C.uint(t.Handle), &ci) != 1 {
		return NDEFInfo{}, false
	}
	return NDEFInfo{
		Length:    uint(ci.current_ndef_length),
		MaxLength: uint(ci.max_ndef_length),
		Writable:  ci.is_writable != 0,
	}, true
}

// Reconnect reconnects and reselects the tag.
func (t Tag) Reconnect() error {
	if rc := C.nfcTag_doHandleReconnect(C.uint(t.Handle)); rc != 0 {
		return StatusError("reconnect", int(rc))
	}
	return nil
}

// Read returns the raw NDEF message and its record type. It returns ErrNotNDEF
// if the tag holds no NDEF message, and (nil, RecordOther, nil) if the message
// is empty.
func (t Tag) Read() ([]byte, RecordType, error) {
	info, ok := t.NDEFInfo()
	if !ok {
		return nil, RecordOther, ErrNotNDEF
	}
	if info.Length == 0 {
		return nil, RecordOther, nil
	}
	buf := make([]byte, info.Length)
	var rt C.nfc_friendly_type_t
	n := C.nfcTag_readNdef(C.uint(t.Handle), bytesPtr(buf), C.uint(len(buf)), &rt)
	if n <= 0 {
		return nil, RecordOther, StatusError("read", int(n))
	}
	if int(n) > len(buf) {
		n = C.int(len(buf))
	}
	return buf[:n], RecordType(rt), nil
}

// Write stores a raw NDEF message on the tag.
func (t Tag) Write(msg []byte) error {
	if len(msg) == 0 {
		return errors.New("nfc: empty message")
	}
	if rc := C.nfcTag_writeNdef(C.uint(t.Handle), bytesPtr(msg), C.uint(len(msg))); rc != 0 {
		return StatusError("write", int(rc))
	}
	return nil
}

// Formatable reports whether the tag can be NDEF-formatted.
// For Mifare Classic the C stack always returns true; actual formatability
// depends on sector keys — use WriteHint for user-facing guidance.
func (t Tag) Formatable() bool {
	return C.nfcTag_isFormatable(C.uint(t.Handle)) == 1
}

// SlowNDEFDetection reports whether NDEF detection needs time after tag arrival.
func (t Tag) SlowNDEFDetection() bool {
	return t.Technology == TargetMifareClassic
}

// WriteHint returns guidance for writing NDEF to this tag, or "" if none applies.
func (t Tag) WriteHint() string {
	switch t.Technology {
	case TargetNDEFFormat:
		return "formatable — try: gonfc write text \"Hello\""
	case TargetMifareClassic:
		return "try: gonfc write text \"Hello\" (requires factory/default Mifare sector keys)"
	}
	if t.Formatable() {
		return "formatable — try: gonfc write text \"Hello\""
	}
	return ""
}

// Format makes the tag NDEF-formatted.
func (t Tag) Format() error {
	if rc := C.nfcTag_formatTag(C.uint(t.Handle)); rc != 0 {
		return StatusError("format", int(rc))
	}
	return nil
}

// MakeReadOnly locks the tag using the given key.
func (t Tag) MakeReadOnly(key []byte) error {
	if len(key) == 0 {
		return errors.New("nfc: empty key")
	}
	if rc := C.nfcTag_makeReadOnly(C.uint(t.Handle), bytesPtr(key), C.uchar(len(key))); rc != 0 {
		return StatusError("make read-only", int(rc))
	}
	return nil
}

// SwitchRF switches the RF interface for ISO-DEP and Mifare Classic tags.
func (t Tag) SwitchRF(frameRF bool) error {
	if rc := C.nfcTag_switchRF(C.uint(t.Handle), boolToCInt(frameRF)); rc != 0 {
		return StatusError("switch RF", int(rc))
	}
	return nil
}

// Transceive sends a raw command and returns the response.
func (t Tag) Transceive(cmd []byte, maxRx int, timeoutMs uint) ([]byte, error) {
	if len(cmd) == 0 {
		return nil, errors.New("nfc: empty command")
	}
	if maxRx <= 0 {
		maxRx = 255
	}
	rx := make([]byte, maxRx)
	n := C.nfcTag_transceive(
		C.uint(t.Handle),
		bytesPtr(cmd),
		C.int(len(cmd)),
		bytesPtr(rx),
		C.int(maxRx),
		C.uint(timeoutMs),
	)
	if n <= 0 {
		return nil, StatusError("transceive", int(n))
	}
	if int(n) > maxRx {
		n = C.int(maxRx)
	}
	return rx[:n], nil
}

// ParseText decodes a Text NDEF record.
func ParseText(record []byte) (lang, text string, err error) {
	if len(record) == 0 {
		return "", "", errors.New("nfc: empty record")
	}
	src := bytesPtr(record)
	langBuf := make([]byte, len(record)+1)
	ln := C.ndef_readLanguageCode(src, C.uint(len(record)),
		(*C.char)(unsafe.Pointer(&langBuf[0])), C.uint(len(langBuf)))
	if ln < 0 {
		return "", "", errors.New("nfc: parse language failed")
	}
	txtBuf := make([]byte, len(record)+1)
	tn := C.ndef_readText(src, C.uint(len(record)),
		(*C.char)(unsafe.Pointer(&txtBuf[0])), C.uint(len(txtBuf)))
	if tn < 0 {
		return "", "", errors.New("nfc: parse text failed")
	}
	return string(langBuf[:ln]), string(txtBuf[:tn]), nil
}

// ParseURI decodes a URI NDEF record.
func ParseURI(record []byte) (string, error) {
	if len(record) == 0 {
		return "", errors.New("nfc: empty record")
	}
	src := bytesPtr(record)
	// The record stores a 1-byte abbreviation code that expands to a URI prefix
	// on decode; +28 covers the longest prefix in the NFC URI identifier table
	// (e.g. "urn:nfc:..."). Bump this if that table ever grows.
	buf := make([]byte, len(record)+28)
	n := C.ndef_readUrl(src, C.uint(len(record)),
		(*C.char)(unsafe.Pointer(&buf[0])), C.uint(len(buf)))
	if n < 0 {
		return "", errors.New("nfc: parse URI failed")
	}
	return string(buf[:n]), nil
}

// ParseHandoverSelect decodes a handover select NDEF record.
func ParseHandoverSelect(record []byte) (HandoverSelect, error) {
	if len(record) == 0 {
		return HandoverSelect{}, errors.New("nfc: empty record")
	}
	var ci C.nfc_handover_select_t
	if rc := C.ndef_readHandoverSelectInfo(bytesPtr(record), C.uint(len(record)), &ci); rc != 0 {
		return HandoverSelect{}, StatusError("parse handover select", int(rc))
	}
	return handoverSelectFromC(&ci), nil
}

// ParseHandoverRequest decodes a handover request NDEF record.
func ParseHandoverRequest(record []byte) (HandoverRequest, error) {
	if len(record) == 0 {
		return HandoverRequest{}, errors.New("nfc: empty record")
	}
	var ci C.nfc_handover_request_t
	if rc := C.ndef_readHandoverRequestInfo(bytesPtr(record), C.uint(len(record)), &ci); rc != 0 {
		return HandoverRequest{}, StatusError("parse handover request", int(rc))
	}
	return handoverRequestFromC(&ci), nil
}

// NewTextRecord builds a Text NDEF record.
func NewTextRecord(lang, text string) ([]byte, error) {
	clang := C.CString(lang)
	defer C.free(unsafe.Pointer(clang))
	ctext := C.CString(text)
	defer C.free(unsafe.Pointer(ctext))
	buf := make([]byte, len(text)+len(lang)+30)
	n := C.ndef_createText(clang, ctext, bytesPtr(buf), C.uint(len(buf)))
	if n <= 0 {
		return nil, errors.New("nfc: build text record failed")
	}
	return buf[:n], nil
}

// NewURIRecord builds a URI NDEF record.
func NewURIRecord(uri string) ([]byte, error) {
	curi := C.CString(uri)
	defer C.free(unsafe.Pointer(curi))
	buf := make([]byte, len(uri)+30)
	n := C.ndef_createUri(curi, bytesPtr(buf), C.uint(len(buf)))
	if n <= 0 {
		return nil, errors.New("nfc: build URI record failed")
	}
	return buf[:n], nil
}

// NewMimeRecord builds a MIME NDEF record.
func NewMimeRecord(mimeType string, data []byte) ([]byte, error) {
	cmime := C.CString(mimeType)
	defer C.free(unsafe.Pointer(cmime))
	buf := make([]byte, len(data)+len(mimeType)+30)
	n := C.ndef_createMime(cmime, bytesPtr(data), C.uint(len(data)), bytesPtr(buf), C.uint(len(buf)))
	if n <= 0 {
		return nil, errors.New("nfc: build MIME record failed")
	}
	return buf[:n], nil
}

// NewHandoverSelectRecord builds a Handover Select NDEF record.
func NewHandoverSelectRecord(power CarrierPowerState, carrierRef string, carrierData []byte) ([]byte, error) {
	cref := C.CString(carrierRef)
	defer C.free(unsafe.Pointer(cref))
	buf := make([]byte, len(carrierData)+len(carrierRef)+30)
	n := C.ndef_createHandoverSelect(
		C.nfc_handover_cps_t(power),
		cref,
		bytesPtr(carrierData),
		C.uint(len(carrierData)),
		bytesPtr(buf),
		C.uint(len(buf)),
	)
	if n <= 0 {
		return nil, errors.New("nfc: build handover select record failed")
	}
	return buf[:n], nil
}
