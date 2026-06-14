package nfc

/*
#include "bridge.h"
*/
import "C"

import (
	"errors"
)

// SetTagHandler registers handlers for tag discovery and removal.
// C callbacks are registered in StartDiscovery, after Open, matching nfcDemoApp.
func SetTagHandler(h TagHandler) {
	setTagHandler(h)
}

// ClearTagHandler removes the tag handler.
func ClearTagHandler() {
	C.nfcgo_deregister_tag_cb()
	clearTagHandler()
}

// SetSnepClientHandler registers handlers for SNEP client peer events.
func SetSnepClientHandler(h PeerHandler) error {
	setSnepClientHandler(h)
	if rc := C.nfcgo_register_snep_client_cb(); rc != 0 {
		return StatusError("set SNEP client handler", int(rc))
	}
	return nil
}

// ClearSnepClientHandler removes the SNEP client handler.
func ClearSnepClientHandler() {
	C.nfcgo_deregister_snep_client_cb()
	clearSnepClientHandler()
}

// StartSnepServer starts a SNEP server with the given handler.
func StartSnepServer(h SnepServerHandler) error {
	setSnepServerHandler(h)
	if rc := C.nfcgo_register_snep_server_cb(); rc != 0 {
		return StatusError("start SNEP server", int(rc))
	}
	return nil
}

// StopSnepServer stops the SNEP server.
func StopSnepServer() {
	C.nfcgo_stop_snep_server()
	clearSnepServerHandler()
}

// SnepPut sends a message to a remote SNEP server.
func SnepPut(msg []byte) error {
	if len(msg) == 0 {
		return errors.New("nfc: empty message")
	}
	if rc := C.nfcSnep_putMessage(bytesPtr(msg), C.uint(len(msg))); rc != 0 {
		return StatusError("SNEP put", int(rc))
	}
	return nil
}
