package main

import (
	"testing"

	"github.com/lulu-box/gonfc/nfc"
)

func TestTagTypeName(t *testing.T) {
	tests := []struct {
		tech uint
		want string
	}{
		{nfc.TargetISO14443_3A, "ISO14443-3A"},
		{nfc.TargetMifareClassic, "Mifare Classic"},
		{nfc.TargetNDEF, "NDEF"},
		{999, "type 999"},
	}
	for _, tc := range tests {
		if got := tagTypeName(tc.tech); got != tc.want {
			t.Errorf("tagTypeName(%d) = %q, want %q", tc.tech, got, tc.want)
		}
	}
}

func TestBuildRecordErrors(t *testing.T) {
	if _, err := buildRecord([]string{"text"}); err == nil {
		t.Error("expected error for missing text argument")
	}
	if _, err := buildRecord(nil); err == nil {
		t.Error("expected error for no arguments")
	}
	if _, err := buildRecord([]string{"bogus", "value"}); err == nil {
		t.Error("expected error for unknown write type")
	}
}
