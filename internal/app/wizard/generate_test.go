package wizard

import (
	"encoding/hex"
	"testing"
)

func TestGenerateSecret_Length(t *testing.T) {
	for _, byteLen := range []int{16, 32, 64} {
		s, err := GenerateSecret(byteLen)
		if err != nil {
			t.Fatalf("GenerateSecret(%d): %v", byteLen, err)
		}
		// hex encodes 2 chars per byte
		if len(s) != byteLen*2 {
			t.Errorf("GenerateSecret(%d): got len %d, want %d", byteLen, len(s), byteLen*2)
		}
	}
}

func TestGenerateSecret_ValidHex(t *testing.T) {
	s, err := GenerateSecret(32)
	if err != nil {
		t.Fatalf("GenerateSecret: %v", err)
	}
	if _, err := hex.DecodeString(s); err != nil {
		t.Errorf("GenerateSecret returned non-hex string: %q", s)
	}
}

func TestGenerateSecret_Unique(t *testing.T) {
	a, err := GenerateSecret(32)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	b, err := GenerateSecret(32)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if a == b {
		t.Error("two GenerateSecret calls returned identical values")
	}
}

func TestGenerateAPIKey_Length(t *testing.T) {
	key, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey: %v", err)
	}
	// 32 bytes → 64 hex chars
	if len(key) != 64 {
		t.Errorf("GenerateAPIKey: got len %d, want 64", len(key))
	}
}
