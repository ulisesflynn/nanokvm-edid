# nanokvm-edid

Custom EDID for the Sipeed NanoKVM Go that advertises a 5120x1440 (32:9 ultrawide) mode.

The generated EDID is 256 bytes:

- **Base block (EDID 1.4):** 2560x1440@60 Hz (CVT reduced blanking) as the fallback preferred mode, plus range limits of 24-75 Hz vertical, 30-140 kHz horizontal and 250 MHz max pixel clock.
- **DisplayID 1.2 extension:** 5120x1440 at ~29.98 Hz (CVT-RB, 231.25 MHz pixel clock).

## Files

- `make_edid.py`: generator script (Python 3, standard library only).
- `kvm-5120x1440-30hz.bin`: generated 256-byte EDID binary.
- `kvm-5120x1440-30hz.hex`: hex dump of the same EDID.

## Usage

```sh
python3 make_edid.py   # writes kvm-5120x1440-30hz.bin and prints the hex
```

Load the resulting `.bin` onto the NanoKVM Go as its custom EDID.
