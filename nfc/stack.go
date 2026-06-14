package nfc

/*
#include "bridge.h"
*/
import "C"

// Open initializes the NFC stack. Call Close when done.
func Open() error {
	C.InitializeLogLevel()
	Debugf("Open: doInitialize")
	rc := C.doInitialize()
	Debugf("Open: doInitialize done (%d)", int(rc))
	if rc != 0 {
		return StatusError("open", int(rc))
	}
	return nil
}

// Close shuts down the NFC stack.
func Close() error {
	Debugf("Close: doDeinitialize")
	rc := C.doDeinitialize()
	Debugf("Close: doDeinitialize done (%d)", int(rc))
	if rc != 0 {
		return StatusError("close", int(rc))
	}
	return nil
}

// Active reports whether the stack is running.
func Active() bool {
	return C.isNfcActive() == 1
}

// StartDiscovery begins polling/listening with the given options.
func StartDiscovery(opts DiscoveryOptions) {
	Debugf("StartDiscovery: registerTagCallback")
	C.nfcgo_register_tag_cb()
	Debugf("StartDiscovery: doEnableDiscovery")
	C.doEnableDiscovery(
		C.int(opts.Technologies),
		boolToCInt(opts.ReaderOnly),
		boolToCInt(opts.HostRouting),
		boolToCInt(opts.Restart),
	)
	Debugf("StartDiscovery done")
}

// StopDiscovery stops polling and listening.
func StopDiscovery() {
	Debugf("StopDiscovery")
	C.disableDiscovery()
	Debugf("StopDiscovery done")
}

// SelectNext activates the next tag in the field.
func SelectNext() error {
	if rc := C.selectNextTag(); rc != 0 {
		return StatusError("select next tag", int(rc))
	}
	return nil
}

// TagCount returns the number of tags currently in the field.
func TagCount() int {
	return int(C.getNumTags())
}

// NextProtocol returns the index of the next valid tag protocol, or -1.
func NextProtocol() int {
	return int(C.checkNextProtocol())
}

// Firmware returns the controller firmware version, or 0 on failure.
func Firmware() int {
	return int(C.getFwVersion())
}
