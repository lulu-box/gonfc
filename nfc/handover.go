package nfc

/*
#include "bridge.h"
*/
import "C"

import (
	"errors"
)

// SetHandoverHandler registers handlers for handover events.
func SetHandoverHandler(h HandoverHandler) error {
	setHandoverHandler(h)
	if rc := C.nfcgo_register_handover_cb(); rc != 0 {
		return StatusError("set handover handler", int(rc))
	}
	return nil
}

// ClearHandoverHandler removes the handover handler.
func ClearHandoverHandler() {
	C.nfcgo_deregister_handover_cb()
	clearHandoverHandler()
}

// SendHandoverSelect sends a handover select message to a peer.
func SendHandoverSelect(message []byte) error {
	if len(message) == 0 {
		return errors.New("nfc: empty message")
	}
	if rc := C.nfcHo_sendSelectRecord(bytesPtr(message), C.uint(len(message))); rc != 0 {
		return StatusError("send handover select", int(rc))
	}
	return nil
}

// SendHandoverError sends a handover select error to a peer.
func SendHandoverError(reason, data uint) error {
	if rc := C.nfcHo_sendSelectError(C.uint(reason), C.uint(data)); rc != 0 {
		return StatusError("send handover error", int(rc))
	}
	return nil
}
