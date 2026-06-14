package nfc

/*
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"fmt"
)

// SetNCIConfig sets an NCI configuration parameter.
func SetNCIConfig(id uint8, data []byte) error {
	if len(data) == 0 {
		return errors.New("nfc: empty config data")
	}
	if rc := C.nfcManager_setConfig(C.uchar(id), C.uchar(len(data)), bytesPtr(data)); rc != 0 {
		return fmt.Errorf("nfc: set NCI config failed (%d)", int(rc))
	}
	return nil
}

// WriteT4T writes an NDEF message to the T4T NFCC in wired mode.
func WriteT4T(command, ndef []byte) error {
	if len(command) == 0 || len(ndef) == 0 {
		return errors.New("nfc: empty command or NDEF buffer")
	}
	if rc := C.doWriteT4tData(bytesPtr(command), bytesPtr(ndef), C.int(len(ndef))); rc != 0 {
		return fmt.Errorf("nfc: write T4T failed (%d)", int(rc))
	}
	return nil
}

// ReadT4T reads an NDEF message from the T4T NFCC in wired mode.
func ReadT4T(command []byte, buf []byte) (int, error) {
	if len(command) == 0 || len(buf) == 0 {
		return 0, errors.New("nfc: empty command or buffer")
	}
	var length C.int = C.int(len(buf))
	if rc := C.doReadT4tData(bytesPtr(command), bytesPtr(buf), &length); rc != 0 {
		return 0, fmt.Errorf("nfc: read T4T failed (%d)", int(rc))
	}
	if int(length) > len(buf) {
		length = C.int(len(buf))
	}
	return int(length), nil
}
