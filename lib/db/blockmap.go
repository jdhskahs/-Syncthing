// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package db

import (
	"encoding/binary"
	"fmt"

	"github.com/syncthing/syncthing/lib/osutil"
	"github.com/syncthing/syncthing/lib/protocol"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var blockFinder *BlockFinder

const maxBatchSize = 1000

type BlockMap struct {
	db     *Lowlevel
	folder uint32
}

func NewBlockMap(db *Lowlevel, folder string) *BlockMap {
	return &BlockMap{
		db:     db,
		folder: db.folderIdx.ID([]byte(folder)),
	}
}

// Add files to the block map, ignoring any deleted or invalid files.
func (m *BlockMap) Add(files []protocol.FileInfo) error {
	batch := new(leveldb.Batch)
	buf := make([]byte, 4)
	var key []byte
	for _, file := range files {
		m.checkFlush(batch)

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
func (m *BlockMap) Update(files []protocol.FileInfo) {
	fn, finished := m.updateFn()
	for _, f := range files {
		fn(f, nil, false)
	}
	finished()
}

func (m *BlockMap) updateFn() (func(nf protocol.FileInfo, ef FileIntf, efOk bool), func()) {
	batch := new(leveldb.Batch)
	buf := make([]byte, 4)
	var key []byte
	return func(file protocol.FileInfo, ef FileIntf, efOk bool) {
		m.checkFlush(batch)

		if efOk {
			m.discard(ef.(protocol.FileInfo), key, batch)
		}

		switch {
		case file.IsDirectory():
		case file.IsDeleted() || file.IsInvalid():
			for _, block := range file.Blocks {
				key = m.blockKeyInto(key, block.Hash, file.Name)
				batch.Delete(key)
			}
		default:
			for i, block := range file.Blocks {
				binary.BigEndian.PutUint32(buf, uint32(i))
				key = m.blockKeyInto(key, block.Hash, file.Name)
				batch.Put(key, buf)
			}
		}
	}, func() { m.db.Write(batch, nil) }
}

// Discard block map state, removing the given files
func (m *BlockMap) Discard(files []protocol.FileInfo) error {
	batch := new(leveldb.Batch)
	var key []byte
	for _, file := range files {
		m.checkFlush(batch)
		m.discard(file, key, batch)
	}
	return m.db.Write(batch, nil)
}

func (m *BlockMap) discard(file protocol.FileInfo, key []byte, batch *leveldb.Batch) {
	for _, block := range file.Blocks {
		key = m.blockKeyInto(key, block.Hash, file.Name)
		batch.Delete(key)
	}
}

func (m *BlockMap) checkFlush(batch *leveldb.Batch) error {
	if batch.Len() > maxBatchSize {
		if err := m.db.Write(batch, nil); err != nil {
			return err
		}
		batch.Reset()
	}
	return nil
}

// Drop block map, removing all entries related to this block map from the db.
func (m *BlockMap) Drop() error {
	batch := new(leveldb.Batch)
	iter := m.db.NewIterator(util.BytesPrefix(m.blockKeyInto(nil, nil, "")[:keyPrefixLen+keyFolderLen]), nil)
	defer iter.Release()
	for iter.Next() {
		m.checkFlush(batch)

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
	db *Lowlevel
}

func NewBlockFinder(db *Lowlevel) *BlockFinder {
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
		folderID := f.db.folderIdx.ID([]byte(folder))
		key = blockKeyInto(key, hash, folderID, "")
		iter := f.db.NewIterator(util.BytesPrefix(key), nil)
		defer iter.Release()

		for iter.Next() && iter.Error() == nil {
			file := blockKeyName(iter.Key())
			index := int32(binary.BigEndian.Uint32(iter.Value()))
			if iterFn(folder, osutil.NativeFilename(file), index) {
				return true
			}
		}
	}
	return false
}

// m.blockKey returns a byte slice encoding the following information:
//	   keyTypeBlock (1 byte)
//	   folder (4 bytes)
//	   block hash (32 bytes)
//	   file name (variable size)
func blockKeyInto(o, hash []byte, folder uint32, file string) []byte {
	reqLen := keyPrefixLen + keyFolderLen + keyHashLen + len(file)
	if cap(o) < reqLen {
		o = make([]byte, reqLen)
	} else {
		o = o[:reqLen]
	}
	o[0] = KeyTypeBlock
	binary.BigEndian.PutUint32(o[keyPrefixLen:], folder)
	copy(o[keyPrefixLen+keyFolderLen:], hash)
	copy(o[keyPrefixLen+keyFolderLen+keyHashLen:], []byte(file))
	return o
}

// blockKeyName returns the file name from the block key
func blockKeyName(data []byte) string {
	if len(data) < keyPrefixLen+keyFolderLen+keyHashLen+1 {
		panic("Incorrect key length")
	}
	if data[0] != KeyTypeBlock {
		panic("Incorrect key type")
	}

	file := string(data[keyPrefixLen+keyFolderLen+keyHashLen:])
	return file
}
