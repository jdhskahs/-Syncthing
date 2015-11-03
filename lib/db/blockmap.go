// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

// Package db provides a set type to track local/remote files with newness
// checks. We must do a certain amount of normalization in here. We will get
// fed paths with either native or wire-format separators and encodings
// depending on who calls us. We transform paths to wire-format (NFC and
// slashes) on the way to the database, and transform to native format
// (varying separator and encoding) on the way back out.
package db

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/syncthing/syncthing/lib/osutil"
	"github.com/syncthing/syncthing/lib/protocol"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var blockFinder *BlockFinder

const maxBatchSize = 256 << 10

type BlockMap struct {
	db     *Instance
	folder string
}

func NewBlockMap(db *Instance, folder string) *BlockMap {
	return &BlockMap{
		db:     db,
		folder: folder,
	}
}

// Add files to the block map, ignoring any deleted or invalid files.
func (m *BlockMap) Add(files []protocol.FileInfo) error {
	batch := new(leveldb.Batch)
	buf := make([]byte, 4)
	var key []byte
	for _, file := range files {
		if batch.Len() > maxBatchSize {
			if err := m.db.Write(batch, nil); err != nil {
				return err
			}
			batch.Reset()
		}

		if file.IsDirectory() || file.IsDeleted() || file.IsInvalid() {
			continue
		}

		for i, block := range file.Blocks {
			binary.BigEndian.PutUint32(buf, uint32(i))
			key = m.blockKeyInto(key, block.Hash, file.Name)
			batch.Put(key, buf)
		}
	}
	return m.db.Write(batch, nil)
}

// Update block map state, removing any deleted or invalid files.
func (m *BlockMap) Update(files []protocol.FileInfo) error {
	batch := new(leveldb.Batch)
	buf := make([]byte, 4)
	var key []byte
	for _, file := range files {
		if batch.Len() > maxBatchSize {
			if err := m.db.Write(batch, nil); err != nil {
				return err
			}
			batch.Reset()
		}

		if file.IsDirectory() {
			continue
		}

		if file.IsDeleted() || file.IsInvalid() {
			for _, block := range file.Blocks {
				key = m.blockKeyInto(key, block.Hash, file.Name)
				batch.Delete(key)
			}
			continue
		}

		for i, block := range file.Blocks {
			binary.BigEndian.PutUint32(buf, uint32(i))
			key = m.blockKeyInto(key, block.Hash, file.Name)
			batch.Put(key, buf)
		}
	}
	return m.db.Write(batch, nil)
}

// Discard block map state, removing the given files
func (m *BlockMap) Discard(files []protocol.FileInfo) error {
	batch := new(leveldb.Batch)
	var key []byte
	for _, file := range files {
		if batch.Len() > maxBatchSize {
			if err := m.db.Write(batch, nil); err != nil {
				return err
			}
			batch.Reset()
		}

		for _, block := range file.Blocks {
			key = m.blockKeyInto(key, block.Hash, file.Name)
			batch.Delete(key)
		}
	}
	return m.db.Write(batch, nil)
}

// Drop block map, removing all entries related to this block map from the db.
func (m *BlockMap) Drop() error {
	batch := new(leveldb.Batch)
	iter := m.db.NewIterator(util.BytesPrefix(m.blockKeyInto(nil, nil, "")[:1+64]), nil)
	defer iter.Release()
	for iter.Next() {
		if batch.Len() > maxBatchSize {
			if err := m.db.Write(batch, nil); err != nil {
				return err
			}
			batch.Reset()
		}

		batch.Delete(iter.Key())
	}
	if iter.Error() != nil {
		return iter.Error()
	}
	return m.db.Write(batch, nil)
}

func (m *BlockMap) blockKeyInto(o, hash []byte, file string) []byte {
	return blockKeyInto(o, hash, m.folder, file)
}

type BlockFinder struct {
	db *Instance
}

func NewBlockFinder(db *Instance) *BlockFinder {
	if blockFinder != nil {
		return blockFinder
	}

	f := &BlockFinder{
		db: db,
	}

	return f
}

func (f *BlockFinder) String() string {
	return fmt.Sprintf("BlockFinder@%p", f)
}

// Iterate takes an iterator function which iterates over all matching blocks
// for the given hash. The iterator function has to return either true (if
// they are happy with the block) or false to continue iterating for whatever
// reason. The iterator finally returns the result, whether or not a
// satisfying block was eventually found.
func (f *BlockFinder) Iterate(folders []string, hash []byte, iterFn func(string, string, int32) bool) bool {
	var key []byte
	for _, folder := range folders {
		key = blockKeyInto(key, hash, folder, "")
		iter := f.db.NewIterator(util.BytesPrefix(key), nil)
		defer iter.Release()

		for iter.Next() && iter.Error() == nil {
			folder, file := fromBlockKey(iter.Key())
			index := int32(binary.BigEndian.Uint32(iter.Value()))
			if iterFn(folder, osutil.NativeFilename(file), index) {
				return true
			}
		}
	}
	return false
}

// Fix repairs incorrect blockmap entries, removing the old entry and
// replacing it with a new entry for the given block
func (f *BlockFinder) Fix(folder, file string, index int32, oldHash, newHash []byte) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(index))

	batch := new(leveldb.Batch)
	batch.Delete(blockKeyInto(nil, oldHash, folder, file))
	batch.Put(blockKeyInto(nil, newHash, folder, file), buf)
	return f.db.Write(batch, nil)
}

// m.blockKey returns a byte slice encoding the following information:
//	   keyTypeBlock (1 byte)
//	   folder (64 bytes)
//	   block hash (32 bytes)
//	   file name (variable size)
func blockKeyInto(o, hash []byte, folder, file string) []byte {
	reqLen := 1 + 64 + 32 + len(file)
	if cap(o) < reqLen {
		o = make([]byte, reqLen)
	} else {
		o = o[:reqLen]
	}
	o[0] = KeyTypeBlock
	copy(o[1:], []byte(folder))
	copy(o[1+64:], []byte(hash))
	copy(o[1+64+32:], []byte(file))
	return o
}

func fromBlockKey(data []byte) (string, string) {
	if len(data) < 1+64+32+1 {
		panic("Incorrect key length")
	}
	if data[0] != KeyTypeBlock {
		panic("Incorrect key type")
	}

	file := string(data[1+64+32:])

	slice := data[1 : 1+64]
	izero := bytes.IndexByte(slice, 0)
	if izero > -1 {
		return string(slice[:izero]), file
	}
	return string(slice), file
}
