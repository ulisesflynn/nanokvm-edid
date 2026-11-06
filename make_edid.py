#!/usr/bin/env python3
"""Build a 256-byte EDID: base block with 2560x1440@60 (CVT-RB) fallback,
plus a DisplayID 1.2 extension declaring 5120x1440@~30Hz (CVT-RB)."""
import struct

def checksum(block):
    return (-sum(block)) & 0xFF

def mfg_id(s):
    v = 0
    for c in s:
        v = (v << 5) | (ord(c) - ord('A') + 1)
    return struct.pack('>H', v)

def dtd(pclk_khz, ha, hb, hfp, hs, va, vb, vfp, vs, wmm, hmm, flags):
    p = pclk_khz // 10
    return bytes([
        p & 0xFF, p >> 8,
        ha & 0xFF, hb & 0xFF, ((ha >> 8) << 4) | (hb >> 8),
        va & 0xFF, vb & 0xFF, ((va >> 8) << 4) | (vb >> 8),
        hfp & 0xFF, hs & 0xFF, ((vfp & 0xF) << 4) | (vs & 0xF),
        ((hfp >> 8) << 6) | ((hs >> 8) << 4) | ((vfp >> 4) << 2) | (vs >> 4),
        wmm & 0xFF, hmm & 0xFF, ((wmm >> 8) << 4) | (hmm >> 8),
        0, 0, flags,
    ])

def text_desc(tag, s):
    b = s.encode('ascii')[:13]
    if len(b) < 13:
        b = b + b'\n' + b' ' * (12 - len(b))
    return bytes([0, 0, 0, tag, 0]) + b

base = bytearray()
base += bytes([0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00])
base += mfg_id('KVM')
base += struct.pack('<H', 0x5120)          # product code
base += struct.pack('<I', 1)               # serial
base += bytes([1, 2026 - 1990])            # week, year
base += bytes([1, 4])                      # EDID 1.4
base += bytes([0xA5])                      # digital, 8bpc, DisplayPort
base += bytes([119, 34])                   # cm
base += bytes([120])                       # gamma 2.2
base += bytes([0x06])                      # sRGB default, preferred timing is native
base += bytes([0xEE, 0x91, 0xA3, 0x54, 0x4C, 0x99, 0x26, 0x0F, 0x50, 0x54])  # sRGB chroma
base += bytes([0x21, 0x08, 0x00])          # 640x480@60, 800x600@60, 1024x768@60
base += bytes([0xD1, 0xC0]) + bytes([0x01, 0x01]) * 7   # 1920x1080@60
# Preferred: 2560x1440@60 CVT-RB, 241.5 MHz, +H -V
base += dtd(241500, 2560, 160, 48, 32, 1440, 41, 3, 5, 1190, 335, 0x1A)
# Range limits: 24-75 Hz V, 30-140 kHz H, 250 MHz max, range-limits-only
base += bytes([0, 0, 0, 0xFD, 0, 24, 75, 30, 140, 25, 0x01, 0x0A] + [0x20] * 6)
base += text_desc(0xFC, 'KVM 5120x1440')
base += bytes([0, 0, 0, 0x10, 0]) + bytes(13)   # dummy descriptor
base += bytes([1])                          # 1 extension
base += bytes([checksum(base)])
assert len(base) == 128

# DisplayID Type I timing: 5120x1440, CVT-RB, 231.25 MHz -> 29.98 Hz
def did_type1(pclk_10khz, flags, ha, hb, hfp, hpos, hs, va, vb, vfp, vpos, vs):
    return (struct.pack('<I', pclk_10khz - 1)[:3] + bytes([flags]) +
            struct.pack('<HHHH', ha - 1, hb - 1, (hfp - 1) | (0x8000 if hpos else 0), hs - 1) +
            struct.pack('<HHHH', va - 1, vb - 1, (vfp - 1) | (0x8000 if vpos else 0), vs - 1))

t = did_type1(23125, 0x80 | 0x08, 5120, 160, 48, True, 32, 1440, 21, 3, False, 10)
assert len(t) == 20
pid = bytes([0x00, 0x00, 0x00]) + struct.pack('<HI', 0x5120, 1) + bytes([1, 26, 0])
blocks = bytes([0x00, 0x00, len(pid)]) + pid
blocks += bytes([0x03, 0x00, len(t)]) + t
section = bytearray([0x12, len(blocks), 0x00, 0x00]) + blocks
section += bytes([checksum(section)])

ext = bytearray([0x70]) + section
ext += bytes(127 - len(ext))
ext += bytes([checksum(ext)])
assert len(ext) == 128

edid = bytes(base + ext)
open('kvm-5120x1440-30hz.bin', 'wb').write(edid)
print('\n'.join(edid[i:i+16].hex(' ') for i in range(0, 256, 16)))
