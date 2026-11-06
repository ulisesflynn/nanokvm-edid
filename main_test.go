package main

import (
	"bytes"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

func TestBuildMatchesReferenceBin(t *testing.T) {
	want, err := os.ReadFile("kvm-5120x1440-30hz.bin")
	if err != nil {
		t.Fatal(err)
	}
	if got := Build(); !bytes.Equal(got, want) {
		t.Fatalf("EDID mismatch\ngot:\n%s\nwant:\n%s", Dump(got), Dump(want))
	}
}

func TestBuildMatchesReferenceHex(t *testing.T) {
	raw, err := os.ReadFile("kvm-5120x1440-30hz.hex")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := hex.EncodeToString(Build()), strings.TrimSpace(string(raw)); got != want {
		t.Fatalf("hex mismatch\ngot:  %s\nwant: %s", got, want)
	}
}

func TestChecksums(t *testing.T) {
	edid := Build()
	if len(edid) != 256 {
		t.Fatalf("len = %d, want 256", len(edid))
	}
	for i := 0; i < 256; i += 128 {
		var sum byte
		for _, b := range edid[i : i+128] {
			sum += b
		}
		if sum != 0 {
			t.Errorf("block %d checksum sum = %#x, want 0", i/128, sum)
		}
	}
}
