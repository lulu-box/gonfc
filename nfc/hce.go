package nfc

/*
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"fmt"
)

// SetHCEHandler registers handlers for host card emulation events.
func SetHCEHandler(h HCEHandler) {
	setHCEHandler(h)
	C.nfcgo_register_hce_cb()
}

// ClearHCEHandler removes the HCE handler.
func ClearHCEHandler() {
	C.nfcgo_deregister_hce_cb()
	clearHCEHandler()
}

// SendAPDU sends an APDU response to the remote reader.
func SendAPDU(apdu []byte) error {
	if len(apdu) == 0 {
		return errors.New("nfc: empty APDU")
	}
	if rc := C.nfcHce_sendCommand(bytesPtr(apdu), C.uint(len(apdu))); rc != 0 {
		return fmt.Errorf("nfc: send APDU failed (%d)", int(rc))
	}
	return nil
}

// RegisterT3TID registers a Type 3 Tag identifier for HCE.
func RegisterT3TID(id []byte) error {
	if len(id) == 0 {
		return errors.New("nfc: empty T3T identifier")
	}
	if rc := C.nfcHce_registerT3tIdentifier(bytesPtr(id), C.uchar(len(id))); rc != 0 {
		return fmt.Errorf("nfc: register T3T ID failed (%d)", int(rc))
	}
	return nil
}
