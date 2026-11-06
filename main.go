// Command nanokvm-edid builds a 256-byte EDID: a base block with a
// 2560x1440@60 (CVT-RB) fallback, plus a DisplayID 1.2 extension declaring
// 5120x1440@~30Hz (CVT-RB). It writes kvm-5120x1440-30hz.bin and prints the
// EDID as hex, 16 bytes per line.
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
)

const outFile = "kvm-5120x1440-30hz.bin"

func checksum(block []byte) byte {
	var sum byte
	for _, b := range block {
		sum += b
	}
	return -sum
}

func mfgID(s string) []byte {
	var v uint16
	for _, c := range s {
		v = v<<5 | uint16(c-'A'+1)
	}
	return binary.BigEndian.AppendUint16(nil, v)
}

func dtd(pclkKHz, ha, hb, hfp, hs, va, vb, vfp, vs, wmm, hmm int, flags byte) []byte {
	p := pclkKHz / 10
	return []byte{
		byte(p), byte(p >> 8),
		byte(ha), byte(hb), byte((ha>>8)<<4 | hb>>8),
		byte(va), byte(vb), byte((va>>8)<<4 | vb>>8),
		byte(hfp), byte(hs), byte((vfp&0xF)<<4 | vs&0xF),
		byte((hfp>>8)<<6 | (hs>>8)<<4 | (vfp>>4)<<2 | vs>>4),
		byte(wmm), byte(hmm), byte((wmm>>8)<<4 | hmm>>8),
		0, 0, flags,
	}
}

func textDesc(tag byte, s string) []byte {
	b := []byte(s)
	if len(b) > 13 {
		b = b[:13]
	}
	if len(b) < 13 {
		b = append(b, '\n')
		b = append(b, bytes.Repeat([]byte{' '}, 13-len(b))...)
	}
	return append([]byte{0, 0, 0, tag, 0}, b...)
}

// didType1 encodes a DisplayID Type I detailed timing (20 bytes).
func didType1(pclk10kHz uint32, flags byte, ha, hb, hfp uint16, hpos bool, hs, va, vb, vfp uint16, vpos bool, vs uint16) []byte {
	pol := func(v uint16, pos bool) uint16 {
		if pos {
			return v | 0x8000
		}
		return v
	}
	t := binary.LittleEndian.AppendUint32(nil, pclk10kHz-1)[:3]
	t = append(t, flags)
	for _, v := range []uint16{ha - 1, hb - 1, pol(hfp-1, hpos), hs - 1, va - 1, vb - 1, pol(vfp-1, vpos), vs - 1} {
		t = binary.LittleEndian.AppendUint16(t, v)
	}
	return t
}

func buildBase() []byte {
	var base []byte
	base = append(base, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00)
	base = append(base, mfgID("KVM")...)
	base = binary.LittleEndian.AppendUint16(base, 0x5120)                           // product code
	base = binary.LittleEndian.AppendUint32(base, 1)                                // serial
	base = append(base, 1, 2026-1990)                                               // week, year
	base = append(base, 1, 4)                                                       // EDID 1.4
	base = append(base, 0xA5)                                                       // digital, 8bpc, DisplayPort
	base = append(base, 119, 34)                                                    // cm
	base = append(base, 120)                                                        // gamma 2.2
	base = append(base, 0x06)                                                       // sRGB default, preferred timing is native
	base = append(base, 0xEE, 0x91, 0xA3, 0x54, 0x4C, 0x99, 0x26, 0x0F, 0x50, 0x54) // sRGB chroma
	base = append(base, 0x21, 0x08, 0x00)                                           // 640x480@60, 800x600@60, 1024x768@60
	base = append(base, 0xD1, 0xC0)                                                 // 1920x1080@60
	for i := 0; i < 7; i++ {
		base = append(base, 0x01, 0x01)
	}
	// Preferred: 2560x1440@60 CVT-RB, 241.5 MHz, +H -V
	base = append(base, dtd(241500, 2560, 160, 48, 32, 1440, 41, 3, 5, 1190, 335, 0x1A)...)
	// Range limits: 24-75 Hz V, 30-140 kHz H, 250 MHz max, range-limits-only
	base = append(base, 0, 0, 0, 0xFD, 0, 24, 75, 30, 140, 25, 0x01, 0x0A)
	base = append(base, bytes.Repeat([]byte{0x20}, 6)...)
	base = append(base, textDesc(0xFC, "KVM 5120x1440")...)
	base = append(base, 0, 0, 0, 0x10, 0) // dummy descriptor
	base = append(base, make([]byte, 13)...)
	base = append(base, 1) // 1 extension
	base = append(base, checksum(base))
	if len(base) != 128 {
		log.Fatalf("base block is %d bytes, want 128", len(base))
	}
	return base
}

func buildExtension() []byte {
	// DisplayID Type I timing: 5120x1440, CVT-RB, 231.25 MHz -> 29.98 Hz
	t := didType1(23125, 0x80|0x08, 5120, 160, 48, true, 32, 1440, 21, 3, false, 10)
	if len(t) != 20 {
		log.Fatalf("timing is %d bytes, want 20", len(t))
	}
	pid := []byte{0x00, 0x00, 0x00}
	pid = binary.LittleEndian.AppendUint16(pid, 0x5120)
	pid = binary.LittleEndian.AppendUint32(pid, 1)
	pid = append(pid, 1, 26, 0)

	blocks := append([]byte{0x00, 0x00, byte(len(pid))}, pid...)
	blocks = append(blocks, 0x03, 0x00, byte(len(t)))
	blocks = append(blocks, t...)

	section := append([]byte{0x12, byte(len(blocks)), 0x00, 0x00}, blocks...)
	section = append(section, checksum(section))

	ext := append([]byte{0x70}, section...)
	ext = append(ext, make([]byte, 127-len(ext))...)
	ext = append(ext, checksum(ext))
	if len(ext) != 128 {
		log.Fatalf("extension block is %d bytes, want 128", len(ext))
	}
	return ext
}

// Build returns the complete 256-byte EDID.
func Build() []byte {
	return append(buildBase(), buildExtension()...)
}

// Dump formats the EDID as space-separated hex, 16 bytes per line.
func Dump(edid []byte) string {
	lines := make([]string, 0, len(edid)/16)
	for i := 0; i < len(edid); i += 16 {
		h := hex.EncodeToString(edid[i:min(i+16, len(edid))])
		pairs := make([]string, 0, len(h)/2)
		for j := 0; j < len(h); j += 2 {
			pairs = append(pairs, h[j:j+2])
		}
		lines = append(lines, strings.Join(pairs, " "))
	}
	return strings.Join(lines, "\n")
}

func main() {
	edid := Build()
	if err := os.WriteFile(outFile, edid, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Println(Dump(edid))
}
