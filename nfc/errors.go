package nfc

import (
	"errors"
	"fmt"
)

// NFC/NCI status codes (nfc_api.h / nci_defs.h).
const (
	StatusOK                 = 0x00
	StatusRejected           = 0x01
	StatusMessageCorrupted   = 0x02
	StatusFailed             = 0x03
	StatusNotInitialized     = 0x04
	StatusSyntaxError        = 0x05
	StatusSemanticError      = 0x06
	StatusUnknownGID         = 0x07
	StatusUnknownOID         = 0x08
	StatusInvalidParam       = 0x09
	StatusMsgSizeTooBig      = 0x0A
	StatusNotSupported       = 0x0B
	StatusAlreadyStarted     = 0xA0
	StatusActivationFailed   = 0xA1
	StatusTearDown           = 0xA2
	StatusRFTransmissionErr  = 0xB0
	StatusRFProtocolErr      = 0xB1
	StatusTimeout            = 0xB2
	StatusEEIntfActiveFail   = 0xC0
	StatusEETransmissionErr  = 0xC1
	StatusEEProtocolErr      = 0xC2
	StatusEETimeout          = 0xC3
	StatusBufferFull         = 0xE0
	StatusCmdStarted         = 0xE3
	StatusHWTimeout          = 0xE4
	StatusContinue           = 0xE5
	StatusRefused            = 0xE6
	StatusBadResp            = 0xE7
	StatusCmdNotCompleted    = 0xE8
	StatusNoBuffers          = 0xE9
	StatusWrongProtocol      = 0xEA
	StatusBusy               = 0xEB
	StatusLinkLoss           = 0xFC
	StatusBadLength          = 0xFD
	StatusBadHandle          = 0xFE
	StatusCongested          = 0xFF
)

var statusText = map[int]string{
	StatusOK:                "OK",
	StatusRejected:          "REJECTED",
	StatusMessageCorrupted:  "MESSAGE_CORRUPTED",
	StatusFailed:            "FAILED",
	StatusNotInitialized:    "NOT_INITIALIZED",
	StatusSyntaxError:       "SYNTAX_ERROR",
	StatusSemanticError:     "SEMANTIC_ERROR",
	StatusUnknownGID:        "UNKNOWN_GID",
	StatusUnknownOID:        "UNKNOWN_OID",
	StatusInvalidParam:      "INVALID_PARAM",
	StatusMsgSizeTooBig:     "MSG_SIZE_TOO_BIG",
	StatusNotSupported:      "NOT_SUPPORTED",
	StatusAlreadyStarted:    "ALREADY_STARTED",
	StatusActivationFailed:  "ACTIVATION_FAILED",
	StatusTearDown:          "TEAR_DOWN",
	StatusRFTransmissionErr: "RF_TRANSMISSION_ERR",
	StatusRFProtocolErr:     "RF_PROTOCOL_ERR",
	StatusTimeout:           "TIMEOUT",
	StatusEEIntfActiveFail:  "EE_INTF_ACTIVE_FAIL",
	StatusEETransmissionErr: "EE_TRANSMISSION_ERR",
	StatusEEProtocolErr:     "EE_PROTOCOL_ERR",
	StatusEETimeout:         "EE_TIMEOUT",
	StatusBufferFull:        "BUFFER_FULL",
	StatusCmdStarted:        "CMD_STARTED",
	StatusHWTimeout:         "HW_TIMEOUT",
	StatusContinue:          "CONTINUE",
	StatusRefused:           "REFUSED",
	StatusBadResp:           "BAD_RESP",
	StatusCmdNotCompleted:   "CMD_NOT_COMPLETED",
	StatusNoBuffers:         "NO_BUFFERS",
	StatusWrongProtocol:     "WRONG_PROTOCOL",
	StatusBusy:              "BUSY",
	StatusLinkLoss:          "LINK_LOSS",
	StatusBadLength:         "BAD_LENGTH",
	StatusBadHandle:         "BAD_HANDLE",
	StatusCongested:         "CONGESTED",
	-1:                      "OPERATION_FAILED",
}

// OpError is an NFC operation failure with a status code from the C stack.
type OpError struct {
	Op   string
	Code int
}

func (e *OpError) Error() string {
	return fmt.Sprintf("nfc: %s failed: %s", e.Op, StatusString(e.Code))
}

// StatusString returns a human-readable name for an NFC/NCI status code.
func StatusString(code int) string {
	if name, ok := statusText[code]; ok {
		if code < 0 {
			return name
		}
		return fmt.Sprintf("%s (0x%02X)", name, code)
	}
	if code < 0 {
		return fmt.Sprintf("error %d", code)
	}
	return fmt.Sprintf("unknown status (0x%02X)", code)
}

// StatusError builds an OpError for a non-zero status code.
func StatusError(op string, code int) error {
	if code == 0 {
		return nil
	}
	return &OpError{Op: op, Code: code}
}

// StatusCode returns the NFC status code from an OpError, if present.
func StatusCode(err error) (int, bool) {
	var e *OpError
	if errors.As(err, &e) {
		return e.Code, true
	}
	return 0, false
}

// IsStatus reports whether err is an OpError with the given status code.
func IsStatus(err error, code int) bool {
	c, ok := StatusCode(err)
	return ok && c == code
}

// IsTagGone reports whether err indicates the tag left the field.
func IsTagGone(err error) bool {
	return IsStatus(err, StatusBadHandle) || IsStatus(err, StatusLinkLoss)
}
