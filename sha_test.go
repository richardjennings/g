package g

import (
	"encoding/hex"
	"testing"
)

func TestNewSha(t *testing.T) {
	// A known 20-byte binary value and its corresponding 40-byte hex encoding.
	binInput := [20]byte{
		0xde, 0xad, 0xbe, 0xef, 0x01,
		0x02, 0x03, 0x04, 0x05, 0x06,
		0x07, 0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}
	hexInput := []byte(hex.EncodeToString(binInput[:]))

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
		wantHex string
	}{
		{
			name:    "valid 40-byte hex input",
			input:   hexInput,
			wantErr: false,
			wantHex: string(hexInput),
		},
		{
			name:    "valid 20-byte binary input",
			input:   binInput[:],
			wantErr: false,
			wantHex: string(hexInput),
		},
		{
			name:    "invalid length 10 bytes",
			input:   []byte("0123456789"),
			wantErr: true,
		},
		{
			name:    "invalid length 0 bytes",
			input:   []byte{},
			wantErr: true,
		},
		{
			name:    "invalid hex in 40-byte input",
			input:   []byte("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sha, err := NewSha(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := sha.AsHexString(); got != tt.wantHex {
				t.Errorf("AsHexString() = %s, want %s", got, tt.wantHex)
			}
			if !sha.IsSet() {
				t.Error("expected IsSet() to be true")
			}
		})
	}
}

func TestShaFromHexString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantHex string
	}{
		{
			name:    "valid hex string",
			input:   "deadbeef01020304050607080900010203040506",
			wantErr: false,
			wantHex: "deadbeef01020304050607080900010203040506",
		},
		{
			name:    "invalid hex characters",
			input:   "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			wantErr: true,
		},
		{
			name:    "odd length hex string",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "wrong decoded length (too short)",
			input:   "deadbeef",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sha, err := ShaFromHexString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := sha.AsHexString(); got != tt.wantHex {
				t.Errorf("AsHexString() = %s, want %s", got, tt.wantHex)
			}
		})
	}
}

func TestSha_Matches(t *testing.T) {
	a, err := ShaFromHexString("deadbeef01020304050607080900010203040506")
	if err != nil {
		t.Fatal(err)
	}
	b, err := ShaFromHexString("deadbeef01020304050607080900010203040506")
	if err != nil {
		t.Fatal(err)
	}
	c, err := ShaFromHexString("1111111111111111111111111111111111111111")
	if err != nil {
		t.Fatal(err)
	}

	if !a.Matches(b) {
		t.Error("expected a to match b")
	}
	if a.Matches(c) {
		t.Error("expected a not to match c")
	}
}

func TestSha_String(t *testing.T) {
	sha, err := ShaFromHexString("deadbeef01020304050607080900010203040506")
	if err != nil {
		t.Fatal(err)
	}
	want := "deadbeef01020304050607080900010203040506"
	if got := sha.String(); got != want {
		t.Errorf("String() = %s, want %s", got, want)
	}
}

func TestSha_IsSet(t *testing.T) {
	// A zero-value Sha should not be set.
	var zero Sha
	if zero.IsSet() {
		t.Error("expected zero Sha to have IsSet() == false")
	}

	// A Sha created via NewSha should be set.
	sha, err := ShaFromHexString("deadbeef01020304050607080900010203040506")
	if err != nil {
		t.Fatal(err)
	}
	if !sha.IsSet() {
		t.Error("expected created Sha to have IsSet() == true")
	}
}

func TestSha_AsHexBytes(t *testing.T) {
	sha, err := ShaFromHexString("deadbeef01020304050607080900010203040506")
	if err != nil {
		t.Fatal(err)
	}
	got := sha.AsHexBytes()
	want := "deadbeef01020304050607080900010203040506"
	if string(got) != want {
		t.Errorf("AsHexBytes() = %s, want %s", got, want)
	}
	if len(got) != 40 {
		t.Errorf("AsHexBytes() length = %d, want 40", len(got))
	}
}

func TestSha_AsArray(t *testing.T) {
	bin := [20]byte{
		0xde, 0xad, 0xbe, 0xef, 0x01,
		0x02, 0x03, 0x04, 0x05, 0x06,
		0x07, 0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}
	sha, err := NewSha(bin[:])
	if err != nil {
		t.Fatal(err)
	}
	got := sha.AsArray()
	if got != bin {
		t.Errorf("AsArray() = %x, want %x", got, bin)
	}
}

func TestSha_AsByteSlice(t *testing.T) {
	bin := [20]byte{
		0xde, 0xad, 0xbe, 0xef, 0x01,
		0x02, 0x03, 0x04, 0x05, 0x06,
		0x07, 0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}
	sha, err := NewSha(bin[:])
	if err != nil {
		t.Fatal(err)
	}
	got := sha.AsByteSlice()
	if len(got) != 20 {
		t.Errorf("AsByteSlice() length = %d, want 20", len(got))
	}
	for i, b := range got {
		if b != bin[i] {
			t.Errorf("AsByteSlice()[%d] = %x, want %x", i, b, bin[i])
		}
	}
}
