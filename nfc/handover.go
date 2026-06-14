package nfc

/*
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"fmt"
)

// SetHandoverHandler registers handlers for handover events.
func SetHandoverHandler(h HandoverHandler) error {
	setHandoverHandler(h)
	if rc := C.nfcgo_register_handover_cb(); rc != 0 {
		return fmt.Errorf("nfc: set handover handler failed (%d)", int(rc))
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
		return fmt.Errorf("nfc: send handover select failed (%d)", int(rc))
	}
	return nil
}

// SendHandoverError sends a handover select error to a peer.
func SendHandoverError(reason, data uint) error {
	if rc := C.nfcHo_sendSelectError(C.uint(reason), C.uint(data)); rc != 0 {
		return fmt.Errorf("nfc: send handover error failed (%d)", int(rc))
	}
	return nil
}
