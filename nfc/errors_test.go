package nfc

import (
	"errors"
	"fmt"
	"testing"
)

func TestStatusString(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{StatusFailed, "FAILED (0x03)"},
		{StatusTimeout, "TIMEOUT (0xB2)"},
		{-1, "OPERATION_FAILED"},
		{0x99, "unknown status (0x99)"},
	}
	for _, tc := range tests {
		if got := StatusString(tc.code); got != tc.want {
			t.Errorf("StatusString(%#x) = %q, want %q", tc.code, got, tc.want)
		}
	}
}

func TestStatusError(t *testing.T) {
	if StatusError("open", 0) != nil {
		t.Fatal("StatusError with code 0 should return nil")
	}
	err := StatusError("write", StatusFailed)
	if got := err.Error(); got != "nfc: write failed: FAILED (0x03)" {
		t.Fatalf("Error() = %q", got)
	}
}

func TestStatusCodeAndIsStatus(t *testing.T) {
	err := StatusError("read", StatusBadHandle)
	wrapped := fmt.Errorf("outer: %w", err)

	code, ok := StatusCode(wrapped)
	if !ok || code != StatusBadHandle {
		t.Fatalf("StatusCode = (%d, %v), want (%d, true)", code, ok, StatusBadHandle)
	}
	if !IsStatus(wrapped, StatusBadHandle) {
		t.Fatal("IsStatus should match wrapped OpError")
	}
	if IsStatus(wrapped, StatusFailed) {
		t.Fatal("IsStatus should not match wrong code")
	}
	if IsStatus(errors.New("other"), StatusBadHandle) {
		t.Fatal("IsStatus should return false for unrelated errors")
	}
}

func TestIsTagGone(t *testing.T) {
	if !IsTagGone(StatusError("reconnect", StatusBadHandle)) {
		t.Fatal("BAD_HANDLE should mean tag gone")
	}
	if !IsTagGone(StatusError("read", StatusLinkLoss)) {
		t.Fatal("LINK_LOSS should mean tag gone")
	}
	if IsTagGone(StatusError("write", StatusFailed)) {
		t.Fatal("FAILED should not mean tag gone")
	}
	wrapped := fmt.Errorf("tag no longer in field: %w", StatusError("reconnect", StatusBadHandle))
	if !IsTagGone(wrapped) {
		t.Fatal("IsTagGone should match wrapped OpError")
	}
}
