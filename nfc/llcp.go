package nfc

/*
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"fmt"
)

// SetLLCPClientHandler registers handlers for connectionless LLCP client events.
func SetLLCPClientHandler(h PeerHandler) error {
	setLLCPClientHandler(h)
	if rc := C.nfcgo_register_llcp_client_cb(); rc != 0 {
		return fmt.Errorf("nfc: set LLCP client handler failed (%d)", int(rc))
	}
	return nil
}

// ClearLLCPClientHandler removes the LLCP client handler.
func ClearLLCPClientHandler() {
	C.nfcgo_deregister_llcp_client_cb()
	clearLLCPClientHandler()
}

// StartLLCPServer starts a connectionless LLCP server.
func StartLLCPServer(h LLCPHandler) error {
	setLLCPHandler(h)
	if rc := C.nfcgo_register_llcp_server_cb(); rc != 0 {
		return fmt.Errorf("nfc: start LLCP server failed (%d)", int(rc))
	}
	return nil
}

// StopLLCPServer stops the connectionless LLCP server.
func StopLLCPServer() {
	C.nfcgo_stop_llcp_server()
	clearLLCPHandler()
}

// LLCPSend sends a connectionless LLCP message.
func LLCPSend(msg []byte) error {
	if len(msg) == 0 {
		return errors.New("nfc: empty message")
	}
	if rc := C.nfcLlcp_ConnLessSendMessage(bytesPtr(msg), C.uint(len(msg))); rc != 0 {
		return fmt.Errorf("nfc: LLCP send failed (%d)", int(rc))
	}
	return nil
}

// LLCPReceive reads a connectionless LLCP message into buf.
func LLCPReceive(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, errors.New("nfc: empty buffer")
	}
	var length C.uint = C.uint(len(buf))
	if rc := C.nfcLlcp_ConnLessReceiveMessage(bytesPtr(buf), &length); rc != 0 {
		return 0, fmt.Errorf("nfc: LLCP receive failed (%d)", int(rc))
	}
	if int(length) > len(buf) {
		length = C.uint(len(buf))
	}
	return int(length), nil
}
