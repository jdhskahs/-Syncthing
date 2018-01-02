// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package scanner

import (
	"bytes"
	"context"
	"fmt"
	"hash"
	"io"

	"github.com/chmduquesne/rollinghash/adler32"
	"github.com/syncthing/syncthing/lib/protocol"
	"github.com/syncthing/syncthing/lib/sha256"
)

var SHA256OfNothing = []uint8{0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14, 0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24, 0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c, 0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55}

type Counter interface {
	Update(bytes int64)
}

func createWriterAndWeakHash(h hash.Hash, useWeakHashes bool) (io.Writer, hash.Hash32) {
	var writer io.Writer
	var h32 hash.Hash32

	if useWeakHashes {
		h32 = adler32.New()
		writer = io.MultiWriter(h, h32)
	} else {
		h32 = noopHash{}
		writer = h
	}
	return writer, h32
}

func prepareReader(sizehint int64, reader io.Reader, blocksize int, hashLength int) ([]protocol.BlockInfo, []byte, *io.LimitedReader) {
	var blocks []protocol.BlockInfo
	var hashes []byte

	if sizehint >= 0 {
		// Allocate contiguous blocks for the BlockInfo structures and their
		// hashes once and for all, and stick to the specified size.
		reader = io.LimitReader(reader, sizehint)
		numBlocks := int(sizehint / int64(blocksize))
		blocks = make([]protocol.BlockInfo, 0, numBlocks)
		hashes = make([]byte, 0, hashLength*numBlocks)
	}

	lreader := io.LimitReader(reader, int64(blocksize)).(*io.LimitedReader)
	return blocks, hashes, lreader
}

// Blocks returns the blockwise hash of the reader.
func Blocks(ctx context.Context, reader io.Reader, blocksize int, sizehint int64, counter Counter, useWeakHashes bool) ([]protocol.BlockInfo, error) {
	strongHash := sha256.New()
	hashSize := strongHash.Size()

	writer, weakHash := createWriterAndWeakHash(strongHash, useWeakHashes)
	blocks, hashes, lreader := prepareReader(sizehint, reader, blocksize, hashSize)

	// A 32k buffer is used for copying into the hash function.
	buf := make([]byte, 32<<10)

	var offset int64
	var thisHash []byte

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		lreader.N = int64(blocksize)
		n, err := io.CopyBuffer(writer, lreader, buf)
		if err != nil {
			return nil, err
		}

		if n == 0 {
			break
		}

		counter.Update(n)

		// Carve out a hash-sized chunk of "hashes" to store the hash for this
		// block.
		hashes = strongHash.Sum(hashes)
		thisHash, hashes = hashes[:hashSize], hashes[hashSize:]

		b := protocol.BlockInfo{
			Size:     int32(n),
			Offset:   offset,
			Hash:     thisHash,
			WeakHash: weakHash.Sum32(),
		}

		blocks = append(blocks, b)
		offset += n

		strongHash.Reset()
		weakHash.Reset()
	}

	if len(blocks) == 0 {
		// Empty file
		blocks = append(blocks, protocol.BlockInfo{
			Offset: 0,
			Size:   0,
			Hash:   SHA256OfNothing,
		})
	}

	return blocks, nil
}

// PopulateOffsets sets the Offset field on each block
func PopulateOffsets(blocks []protocol.BlockInfo) {
	var offset int64
	for i := range blocks {
		blocks[i].Offset = offset
		offset += int64(blocks[i].Size)
	}
}

// BlockDiff returns lists of common and missing (to transform src into tgt)
// blocks. Both block lists must have been created with the same block size.
func BlockDiff(src, tgt []protocol.BlockInfo) (have, need []protocol.BlockInfo) {
	if len(tgt) == 0 && len(src) != 0 {
		return nil, nil
	}

	if len(tgt) != 0 && len(src) == 0 {
		// Copy the entire file
		return nil, tgt
	}

	for i := range tgt {
		if i >= len(src) || !bytes.Equal(tgt[i].Hash, src[i].Hash) {
			// Copy differing block
			need = append(need, tgt[i])
		} else {
			have = append(have, tgt[i])
		}
	}

	return have, need
}

// Verify returns nil or an error describing the mismatch between the block
// list and actual reader contents
func Verify(r io.Reader, blocksize int, blocks []protocol.BlockInfo) error {
	hf := sha256.New()
	// A 32k buffer is used for copying into the hash function.
	buf := make([]byte, 32<<10)

	for i, block := range blocks {
		lr := &io.LimitedReader{R: r, N: int64(blocksize)}
		_, err := io.CopyBuffer(hf, lr, buf)
		if err != nil {
			return err
		}

		hash := hf.Sum(nil)
		hf.Reset()

		if !bytes.Equal(hash, block.Hash) {
			return fmt.Errorf("hash mismatch %x != %x for block %d", hash, block.Hash, i)
		}
	}

	// We should have reached the end  now
	bs := make([]byte, 1)
	n, err := r.Read(bs)
	if n != 0 || err != io.EOF {
		return fmt.Errorf("file continues past end of blocks")
	}

	return nil
}

func VerifyBuffer(buf []byte, block protocol.BlockInfo) ([]byte, error) {
	if len(buf) != int(block.Size) {
		return nil, fmt.Errorf("length mismatch %d != %d", len(buf), block.Size)
	}
	hf := sha256.New()
	_, err := hf.Write(buf)
	if err != nil {
		return nil, err
	}
	hash := hf.Sum(nil)

	if !bytes.Equal(hash, block.Hash) {
		return hash, fmt.Errorf("hash mismatch %x != %x", hash, block.Hash)
	}

	return hash, nil
}

// BlocksEqual returns whether two slices of blocks are exactly the same hash
// and index pair wise.
func BlocksEqual(src, tgt []protocol.BlockInfo) bool {
	if len(tgt) != len(src) {
		return false
	}

	for i, sblk := range src {
		if !bytes.Equal(sblk.Hash, tgt[i].Hash) {
			return false
		}
	}
	return true
}

type noopHash struct{}

func (noopHash) Sum32() uint32             { return 0 }
func (noopHash) BlockSize() int            { return 0 }
func (noopHash) Size() int                 { return 0 }
func (noopHash) Reset()                    {}
func (noopHash) Sum([]byte) []byte         { return nil }
func (noopHash) Write([]byte) (int, error) { return 0, nil }

type NoopCounter struct{}

func (c *NoopCounter) Update(bytes int64) {}
