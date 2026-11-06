# nanokvm-edid

Custom EDID for the Sipeed NanoKVM Go that advertises a 5120x1440 (32:9 super ultrawide) mode.

The 5120x1440 mode targets the Samsung Odyssey G9 (49" dual-QHD) and other monitors with the same 5120x1440 resolution. With this EDID loaded, the host can drive its native 32:9 desktop through the KVM instead of falling back to a 16:9 mode.

The generated EDID is 256 bytes:

- **Base block (EDID 1.4):** 2560x1440@60 Hz (CVT reduced blanking) as the fallback preferred mode, plus range limits of 24-75 Hz vertical, 30-140 kHz horizontal and 250 MHz max pixel clock.
- **DisplayID 1.2 extension:** 5120x1440 at ~29.98 Hz (CVT-RB, 231.25 MHz pixel clock).

## Files

- `main.go`: generator (Go, standard library only).
- `main_test.go`: checks the generated EDID against the reference `.bin`/`.hex` and verifies both block checksums.
- `kvm-5120x1440-30hz.bin`: generated 256-byte EDID binary.
- `kvm-5120x1440-30hz.hex`: hex dump of the same EDID.

## Usage

Requires Go 1.21 or newer.

Generate the EDID from the repository root:

```sh
go run .
```

This writes `kvm-5120x1440-30hz.bin` to the current directory and prints the EDID as hex, 16 bytes per line:

```
00 ff ff ff ff ff ff 00 2e cd 20 51 01 00 00 00
01 24 01 04 a5 77 22 78 06 ee 91 a3 54 4c 99 26
...
```

Or build a standalone binary and run it from any directory (the `.bin` is written to the directory you run it in):

```sh
go build -o nanokvm-edid .
./nanokvm-edid
```

Run the tests to confirm the output still matches the reference EDID:

```sh
go test ./...
```

To change the timings, edit the `buildBase` and `buildExtension` functions in `main.go`. Then regenerate, and update `kvm-5120x1440-30hz.hex` (`xxd -p -c 256 kvm-5120x1440-30hz.bin > kvm-5120x1440-30hz.hex`) so the tests match the new output.

Load the resulting `.bin` onto the NanoKVM Go as its custom EDID.
