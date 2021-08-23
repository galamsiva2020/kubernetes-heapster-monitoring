// Copyright 2011 The Snappy-Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package snappy implements the snappy block-based compression format.
// It aims for very high speeds and reasonable compression.
//
// The C++ snappy implementation is at https://github.com/google/snappy
package snappy

import (
	"hash/crc32"
)

/*
Each encoded block begins with the varint-encoded length of the decoded data,
followed by a sequence of chunks. Chunks begin and end on byte boundaries. The
first byte of each chunk is broken into its 2 least and 6 most significant bits
called l and m: l ranges in [0, 4) and m ranges in [0, 64). l is the chunk tag.
Zero means a literal tag. All other values mean a copy tag.

For literal tags:
  - If m < 60, the next 1 + m bytes are literal bytes.
  - Otherwise, let n be the little-endian unsigned integer denoted by the next
    m - 59 bytes. The next 1 + n bytes after that are literal bytes.

For copy tags, length bytes are copied from offset bytes ago, in the style of
Lempel-Ziv compression algorithms. In particular:
  - For l == 1, the offset ranges in [0, 1<<11) and the length in [4, 12).
    The length is 4 + the low 3 bits of m. The high 3 bits of m form bits 8-10
    of the offset. The next byte is bits 0-7 of the offset.
  - For l == 2, the offset ranges in [0, 1<<16) and the length in [1, 65).
    The length is 1 + m. The offset is the little-endian unsigned integer
    denoted by the next 2 bytes.
  - For l == 3, this tag is a legacy format that is no longer supported.
*/
const (
	tagLiteral = 0x00
	tagCopy1   = 0x01
	tagCopy2   = 0x02
	tagCopy4   = 0x03
)

const (
	checksumSize    = 4
	chunkHeaderSize = 4
	magicChunk      = "\xff\x06\x00\x00" + magicBody
	magicBody       = "sNaPpY"
	// https://github.com/google/snappy/blob/master/framing_format.txt says
	// that "the uncompressed data in a chunk must be no longer than 65536 bytes".
	maxUncompressedChunkLen = 65536
)

const (
	chunkTypeCompressedData   = 0x00
	chunkTypeUncompressedData = 0x01
	chunkTypePadding          = 0xfe
	chunkTypeStreamIdentifier = 0xff
)

var crcTable = crc32.MakeTable(crc32.Castagnoli)

// crc implements the checksum specified in section 3 of
// https://github.com/google/snappy/blob/master/framing_format.txt
func crc(b []byte) uint32 {
	c := crc32.Update(0, crcTable, b)
	return uint32(c>>15|c<<17) + 0xa282ead8
}
