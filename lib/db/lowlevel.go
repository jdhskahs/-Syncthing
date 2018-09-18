// Copyright (C) 2018 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package db

import (
	"os"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

const (
	dbMaxOpenFiles = 100
	dbWriteBuffer  = 4 << 20
)

// Lowlevel is the lowest level database interface. It has a very simple
// purpose: hold the actual *leveldb.DB database, and the in-memory state
// that belong to that database. In the same way that a single on disk
// database can only be opened once, there should be only one Lowlevel for
// any given *leveldb.DB.
type Lowlevel struct {
	*leveldb.DB
	location  string
	folderIdx *smallIndex
	deviceIdx *smallIndex
}

// OpenLowlevel attempts to open the database at the given location,
// and runs recovery on it if opening fails. Worst case, if recovery is not
// possible, the database is erased and created from scratch.
func OpenLowlevel(location string) (*Lowlevel, error) {
	opts := &opt.Options{
		OpenFilesCacheCapacity: dbMaxOpenFiles,
		WriteBuffer:            dbWriteBuffer,
	}

	db, err := leveldb.OpenFile(location, opts)
	if leveldbIsCorrupted(err) {
		db, err = leveldb.RecoverFile(location, opts)
	}
	if leveldbIsCorrupted(err) {
		// The database is corrupted, and we've tried to recover it but it
		// didn't work. At this point there isn't much to do beyond dropping
		// the database and reindexing...
		l.Infoln("Database corruption detected, unable to recover. Reinitializing...")
		if err := os.RemoveAll(location); err != nil {
			return nil, errorSuggestion{err, "failed to delete corrupted database"}
		}
		db, err = leveldb.OpenFile(location, opts)
	}
	if err != nil {
		return nil, errorSuggestion{err, "is another instance of Syncthing running?"}
	}
	return NewLowlevel(db, location), nil
}

func OpenMemoryLowlevel() *Lowlevel {
	db, _ := leveldb.Open(storage.NewMemStorage(), nil)
	return NewLowlevel(db, "<memory>")
}

// lowlevelFromDB wraps the given *leveldb.DB into a *lowlevel
func NewLowlevel(db *leveldb.DB, location string) *Lowlevel {
	return &Lowlevel{
		DB:        db,
		location:  location,
		folderIdx: newSmallIndex(db, []byte{KeyTypeFolderIdx}),
		deviceIdx: newSmallIndex(db, []byte{KeyTypeDeviceIdx}),
	}
}

// A "better" version of leveldb's errors.IsCorrupted.
func leveldbIsCorrupted(err error) bool {
	switch {
	case err == nil:
		return false

	case errors.IsCorrupted(err):
		return true

	case strings.Contains(err.Error(), "corrupted"):
		return true
	}

	return false
}
