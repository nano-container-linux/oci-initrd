// Package xzopen opens the .xz stream of a rootfs layer with a memory bound
// that does not depend on what the stream's header declares.
//
// It is its own package, and not a function in src/init.go, for two reasons:
// the Taskfile builds the init binary from the single file src/init.go (a
// second file in package main would not be compiled), and init.go imports
// linux-only packages, so nothing in it can be tested off Linux.
package xzopen

import (
	"io"

	"github.com/go-compressions/xz"
)

// MaxDictionary bounds the window an .xz rootfs may make the decoder keep:
// 256 MiB, go-compressions/xz's own default, stated here so the bound is
// readable (xz -9 uses 64 MiB). The window grows with the decoded data, never
// from the dictionary size a header declares, so this caps what real data can
// ask for and is not an allocation. The rootfs is extracted into RAM anyway:
// data that needs more fails with an error that satisfies
// errors.Is(err, xz.ErrTooLarge).
const MaxDictionary = 256 << 20

// NewReader returns a streaming decoder for the .xz stream(s) in r.
func NewReader(r io.Reader) (*xz.Reader, error) {
	return xz.ReaderConfig{MaxDictionary: MaxDictionary}.NewReader(r)
}
