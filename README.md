# oci-initrd

Pure-Go initramfs `/init` (PID 1) that boots a machine by pulling its
**kernel modules** and **root filesystem** from an **OCI registry** over
HTTP, extracting them into the ramdisk, and handing off to the real init.

No shell, no busybox: a single static Go binary drives the whole early
boot.

## How it works

The binary is installed as `/init` inside the initramfs. On start it:

1. Mounts `/proc`, `/sys` and `/dev`.
2. Configures the network from the kernel command line (via `netlink`).
3. To avoid a name clash with the root filesystem's own `/init`, renames
   itself to `/oci-init` and re-executes.
4. Downloads the **modloop** image (kernel modules) to
   `/modloop.squashfs` and mounts it.
5. Downloads the **rootfs** image and extracts its tar stream
   (`gzip`- or `xz`-compressed) into the ramdisk.
6. Boots into the extracted root filesystem.

Blobs are fetched from the registry with a progress bar
(`schollz/progressbar`).

## Kernel command line

The boot target is selected entirely through `/proc/cmdline`:

| Parameter  | Meaning                                                        |
| ---------- | -------------------------------------------------------------- |
| `oci_http` | Base HTTP endpoint of the OCI registry to pull blobs from.     |
| `modloop`  | OCI reference of the kernel-modules (`.squashfs`) image.       |
| `rootfs`   | OCI reference of the root filesystem (tar, `gzip` or `xz`).    |

References accept `http://`, `https://` and `oci://` schemes.

## What bounds an `xz` rootfs

The rootfs arrives from the network, and `/init` is PID 1 on a machine whose
memory is the ramdisk, so the decoder must not size its memory from the
stream. The `xz` layer is decoded with
[`go-compressions/xz`](https://github.com/go-compressions/xz), not
`ulikunitz/xz`. The latter allocates the LZMA2 dictionary whole from the block
header (`DictCap` is a minimum, not a ceiling): a 44-byte `.xz` declaring a
1.5 GiB dictionary allocated 1,610,663,976 bytes. Here the window grows with
the decoded data: a stream declaring 1.5 GiB over a few bytes allocates about
22 KB (`internal/xzopen/xzopen_test.go`, run in CI as a compiled binary under
`ulimit -v` and `GOMEMLIMIT`). The dictionary a stream may make the decoder
keep is capped at 256 MiB (`xzopen.MaxDictionary`, above the 64 MiB of
`xz -9`); data that needs more fails with `xz.ErrTooLarge` and the boot stops
with "Failed to open xz rootfs" or "Error reading tar rootfs". The `/init`
binary is slightly smaller than before (11,299,388 to 11,235,046 bytes,
linux/amd64).

The reader lives in `internal/xzopen` rather than in `src/init.go` because the
Taskfile builds the init binary from that one file and because `init.go`
imports Linux-only packages, so nothing in it can be tested elsewhere.

## Build

Uses [Task](https://taskfile.dev). The default architecture is `arm64`.

```sh
task build              # -> out/initrd-oci-arm64.gz
task build ARCH=amd64   # -> out/initrd-oci-amd64.gz
task clean
task help               # list tasks
```

The Go module is named `oci-init`; the source lives in `src/init.go`.

## Relationship to the rest of nano-container-linux

`oci-initrd` produces the initramfs that fetches its payload from an OCI
registry; its sibling [`oci-pxe`](https://github.com/nano-container-linux/oci-pxe)
covers the PXE side of the same "boot from an OCI registry" idea.

## License

BSD-3-Clause — see [LICENSE](LICENSE).
