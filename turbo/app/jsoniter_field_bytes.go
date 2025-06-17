//go:build !windows
// +build !windows

package app

import (
	_ "unsafe" // required for go:linkname

	jsoniter "github.com/json-iterator/go"
)

// readObjectFieldAsBytes is a thin wrapper around jsoniter's unexported
// (*Iterator).readObjectFieldAsBytes. We bind to it via go:linkname so we can
// benefit from jsoniter's zero-allocation field name parsing without copying
// its implementation or touching unexported symbols directly.
//
// WARNING: the linkname directive is compiler-level; if json-iterator changes
// the internal symbol name this will break. Tested against v1.1.12.
//
//go:linkname jsoniter_readObjectFieldAsBytes github.com/json-iterator/go.(*Iterator).readObjectFieldAsBytes
func jsoniter_readObjectFieldAsBytes(iter *jsoniter.Iterator) []byte

// readObjectFieldAsBytes forwards to the real implementation linked above.
func readObjectFieldAsBytes(iter *jsoniter.Iterator) []byte {
	return jsoniter_readObjectFieldAsBytes(iter)
}

// Pre-allocated field key byte slices to avoid repeatedly allocating the same
// literals during comparisons.
var (
	fieldBalanceBytes = []byte("balance")
	fieldNonceBytes   = []byte("nonce")
	fieldCodeBytes    = []byte("code")
	fieldStorageBytes = []byte("storage")
)
