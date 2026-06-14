// Package nfc provides Go bindings for the linux_libnfc-nci C API
// (linux_nfc_api.h) via cgo.
package nfc

/*
#cgo pkg-config: libnfc-nci
#cgo LDFLAGS: -Wl,-rpath,/usr/local/lib

#include <stdlib.h>
#include "bridge.h"
*/
import "C"

import (
	"unsafe"
)

func boolToCInt(v bool) C.int {
	if v {
		return 1
	}
	return 0
}

func bytesPtr(b []byte) *C.uchar {
	if len(b) == 0 {
		return nil
	}
	return (*C.uchar)(unsafe.Pointer(&b[0]))
}

func tagFromC(info *C.nfc_tag_info_t) Tag {
	n := int(info.uid_length)
	if n < 0 {
		n = 0
	}
	if n > len(info.uid) {
		n = len(info.uid)
	}
	var uid []byte
	if n > 0 {
		uid = C.GoBytes(unsafe.Pointer(&info.uid[0]), C.int(n))
	}
	return Tag{
		Technology: uint(info.technology),
		Handle:     uint(info.handle),
		UID:        uid,
		Protocol:   uint8(info.protocol),
	}
}

func handoverSelectFromC(c *C.nfc_handover_select_t) HandoverSelect {
	return HandoverSelect{
		Bluetooth: bluetoothFromC(&c.bluetooth),
		WiFi:      wifiPairingFromC(&c.wifi),
	}
}

func handoverRequestFromC(c *C.nfc_handover_request_t) HandoverRequest {
	var out HandoverRequest
	out.Bluetooth = bluetoothFromC(&c.bluetooth)
	if c.wifi.has_wifi != 0 {
		out.WiFi.HasWiFi = true
	}
	if c.wifi.ndef_length > 0 && c.wifi.ndef != nil {
		out.WiFi.NDEF = C.GoBytes(unsafe.Pointer(c.wifi.ndef), C.int(c.wifi.ndef_length))
	}
	return out
}

func bluetoothFromC(c *C.nfc_btoob_pairing_t) BluetoothPairing {
	var out BluetoothPairing
	out.PowerState = CarrierPowerState(c.power_state)
	out.Type = BluetoothCarrierType(C.nfcgo_bt_type(c))
	if c.ndef_length > 0 && c.ndef != nil {
		out.NDEF = C.GoBytes(unsafe.Pointer(c.ndef), C.int(c.ndef_length))
	}
	for i := 0; i < 6; i++ {
		out.Address[i] = byte(c.address[i])
	}
	if c.device_name_length > 0 && c.device_name != nil {
		out.DeviceName = C.GoBytes(unsafe.Pointer(c.device_name), C.int(c.device_name_length))
	}
	return out
}

func wifiPairingFromC(c *C.nfc_wifi_pairing_t) WiFiPairing {
	var out WiFiPairing
	out.PowerState = CarrierPowerState(c.power_state)
	if c.ndef_length > 0 && c.ndef != nil {
		out.NDEF = C.GoBytes(unsafe.Pointer(c.ndef), C.int(c.ndef_length))
	}
	if c.ssid_length > 0 && c.ssid != nil {
		out.SSID = C.GoBytes(unsafe.Pointer(c.ssid), C.int(c.ssid_length))
	}
	if c.key_length > 0 && c.key != nil {
		out.Key = C.GoBytes(unsafe.Pointer(c.key), C.int(c.key_length))
	}
	return out
}

// InitLogLevel loads log levels from the library config files.
func InitLogLevel() {
	C.InitializeLogLevel()
}
