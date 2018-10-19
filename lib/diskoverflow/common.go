// Copyright (C) 2016 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

// Package diskoverflow provides several data container types which are limited
// in their memory usage. Once the total memory limit is reached, all new data
// is written to disk.
// Do not use any instances of these types concurrently!
package diskoverflow

import "github.com/syncthing/syncthing/lib/protocol"

// Value must be implemented by every type that is to be stored in a disk spilling container.
type Value interface {
	Size() int64
	Marshal() []byte
	Unmarshal([]byte) Value // The returned Value must not be a reference to the receiver.
}

// ValueFileInfo implements Value for protocol.FileInfo
type ValueFileInfo struct{ protocol.FileInfo }

func (s *ValueFileInfo) Size() int64 {
	return int64(s.ProtoSize())
}

func (s *ValueFileInfo) Marshal() []byte {
	data, err := s.FileInfo.Marshal()
	if err != nil {
		panic("bug: marshalling FileInfo should never fail: " + err.Error())
	}
	return data
}

func (s *ValueFileInfo) Unmarshal(v []byte) Value {
	out := &ValueFileInfo{}
	if err := out.FileInfo.Unmarshal(v); err != nil {
		panic("unmarshal failed: " + err.Error())
	}
	return out
}

// Magical limit below which the underlying containers of slices/maps are never
// reset to save space.
// Variable for test, shouldn't ever be changed in code.
var minCompactionSize int64 = 10 << protocol.MiB

type common interface {
	close()
	length() int
}

type Iterator interface {
	Release()
	Next() bool
}

type ValueIterator interface {
	Iterator
	Value() Value
}

type SortValueIterator interface {
	Iterator
	Value() SortValue
}

type iteratorParent interface {
	released()
	value() interface{}
}

const (
	concurrencyMsg = "iteration in progress - don't modify or start a new iteration concurrently"
)

type memIterator struct {
	pos     int
	len     int
	reverse bool
	parent  iteratorParent
}

func newMemIterator(p iteratorParent, reverse bool, len int) *memIterator {
	it := &memIterator{
		len:     len,
		reverse: reverse,
		parent:  p,
	}
	if reverse {
		it.pos = len
	} else {
		it.pos = -1
	}
	return it
}

func (si *memIterator) Next() bool {
	if si.reverse {
		if si.pos == 0 {
			return false
		}
		si.pos--
		return true
	}
	if si.pos == si.len-1 {
		return false
	}
	si.pos++
	return true
}

func (si *memIterator) Release() {
	si.parent.released()
}
