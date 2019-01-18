// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package db

import (
	"github.com/syncthing/syncthing/lib/protocol"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Flush batches to disk when they contain this many records.
const batchFlushSize = 64

// A readOnlyTransaction represents a database snapshot.
type readOnlyTransaction struct {
	*leveldb.Snapshot
	db *instance
}

func (db *instance) newReadOnlyTransaction() readOnlyTransaction {
	snap, err := db.GetSnapshot()
	if err != nil {
		panic(err)
	}
	return readOnlyTransaction{
		Snapshot: snap,
		db:       db,
	}
}

func (t readOnlyTransaction) close() {
	t.Release()
}

func (t readOnlyTransaction) getFile(folder, device, file []byte) (protocol.FileInfo, bool) {
	return t.getFileByKey(t.db.keyer.GenerateDeviceFileKey(nil, folder, device, file))
}

func (t readOnlyTransaction) getFileByKey(key []byte) (protocol.FileInfo, bool) {
	if f, ok := t.getFileTrunc(key, false); ok {
		return f.(protocol.FileInfo), true
	}
	return protocol.FileInfo{}, false
}

func (t readOnlyTransaction) getFileTrunc(key []byte, trunc bool) (FileIntf, bool) {
	bs, err := t.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return nil, false
	}
	if err != nil {
		l.Debugln("surprise error:", err)
		return nil, false
	}

	f, err := unmarshalTrunc(bs, trunc)
	if err != nil {
		l.Debugln("unmarshal error:", err)
		return nil, false
	}
	return f, true
}

func (t readOnlyTransaction) getGlobalInto(gk, dk, folder, file []byte, truncate bool) ([]byte, []byte, FileIntf, bool) {
	gk = t.db.keyer.GenerateGlobalVersionKey(gk, folder, file)

	bs, err := t.Get(gk, nil)
	if err != nil {
		return gk, dk, nil, false
	}

	vl, ok := unmarshalVersionList(bs)
	if !ok {
		return gk, dk, nil, false
	}

	dk = t.db.keyer.GenerateDeviceFileKey(dk, folder, vl.Versions[0].Device, file)
	if fi, ok := t.getFileTrunc(dk, truncate); ok {
		return gk, dk, fi, true
	}

	return gk, dk, nil, false
}

// A readWriteTransaction is a readOnlyTransaction plus a batch for writes.
// The batch will be committed on close() or by checkFlush() if it exceeds the
// batch size.
type readWriteTransaction struct {
	readOnlyTransaction
	*leveldb.Batch
}

func (db *instance) newReadWriteTransaction() readWriteTransaction {
	t := db.newReadOnlyTransaction()
	return readWriteTransaction{
		readOnlyTransaction: t,
		Batch:               new(leveldb.Batch),
	}
}

func (t readWriteTransaction) close() {
	t.flush()
	t.readOnlyTransaction.close()
}

func (t readWriteTransaction) checkFlush() {
	if t.Batch.Len() > batchFlushSize {
		t.flush()
		t.Batch.Reset()
	}
}

func (t readWriteTransaction) flush() {
	if err := t.db.Write(t.Batch, nil); err != nil {
		panic(err)
	}
}

func (t readWriteTransaction) insertFile(fk, folder []byte, devID protocol.DeviceID, file protocol.FileInfo) {
	l.Debugf("insert; folder=%q device=%v %v", folder, devID, file)

	t.Put(fk, mustMarshal(&file))
}

// updateGlobal adds this device+version to the version list for the given
// file. If the device is already present in the list, the version is updated.
// If the file does not have an entry in the global list, it is created.
func (t readWriteTransaction) updateGlobal(gk, folder []byte, devID protocol.DeviceID, file protocol.FileInfo, meta *metadataTracker) bool {
	l.Debugf("update global; folder=%q device=%v file=%q version=%v invalid=%v", folder, devID, file.Name, file.Version, file.IsInvalid())

	var fl VersionList
	if svl, err := t.Get(gk, nil); err == nil {
		fl.Unmarshal(svl) // Ignore error, continue with empty fl
	}
	fl, removedFV, removedAt, insertedAt := fl.update(folder, devID[:], file, t.readOnlyTransaction)
	if insertedAt == -1 {
		l.Debugln("update global; same version, global unchanged")
		return false
	}

	name := []byte(file.Name)

	var global protocol.FileInfo
	if insertedAt == 0 {
		// Inserted a new newest version
		global = file
	} else if new, ok := t.getFile(folder, fl.Versions[0].Device, name); ok {
		global = new
	} else {
		panic("This file must exist in the db")
	}

	// Fixup the list of files we need.
	t.updateLocalNeed(folder, name, fl, global)

	if removedAt != 0 && insertedAt != 0 {
		l.Debugf(`new global for "%v" after update: %v`, file.Name, fl)
		t.Put(gk, mustMarshal(&fl))
		return true
	}

	// Remove the old global from the global size counter
	var oldGlobalFV FileVersion
	if removedAt == 0 {
		oldGlobalFV = removedFV
	} else if len(fl.Versions) > 1 {
		// The previous newest version is now at index 1
		oldGlobalFV = fl.Versions[1]
	}
	if oldFile, ok := t.getFile(folder, oldGlobalFV.Device, name); ok {
		// A failure to get the file here is surprising and our
		// global size data will be incorrect until a restart...
		meta.removeFile(protocol.GlobalDeviceID, oldFile)
	}

	// Add the new global to the global size counter
	meta.addFile(protocol.GlobalDeviceID, global)

	l.Debugf(`new global for "%v" after update: %v`, file.Name, fl)
	t.Put(gk, mustMarshal(&fl))

	return true
}

// updateLocalNeeds checks whether the given file is still needed on the local
// device according to the version list and global FileInfo given and updates
// the db accordingly.
func (t readWriteTransaction) updateLocalNeed(folder, name []byte, fl VersionList, global protocol.FileInfo) {
	nk := t.db.keyer.GenerateNeedFileKey(nil, folder, name)
	hasNeeded, _ := t.db.Has(nk, nil)
	if localFV, haveLocalFV := fl.Get(protocol.LocalDeviceID[:]); need(global, haveLocalFV, localFV.Version) {
		if !hasNeeded {
			l.Debugf("local need insert; folder=%q, name=%q", folder, name)
			t.Put(nk, nil)
		}
	} else if hasNeeded {
		l.Debugf("local need delete; folder=%q, name=%q", folder, name)
		t.Delete(nk)
	}
}

func need(global FileIntf, haveLocal bool, localVersion protocol.Vector) bool {
	// We never need an invalid file.
	if global.IsInvalid() {
		return false
	}
	// We don't need a deleted file if we don't have it.
	if global.IsDeleted() && !haveLocal {
		return false
	}
	// We don't need the global file if we already have the same version.
	if haveLocal && localVersion.GreaterEqual(global.FileVersion()) {
		return false
	}
	return true
}

// removeFromGlobal removes the device from the global version list for the
// given file. If the version list is empty after this, the file entry is
// removed entirely.
func (t readWriteTransaction) removeFromGlobal(gk, folder []byte, devID protocol.DeviceID, file []byte, meta *metadataTracker) {
	l.Debugf("remove from global; folder=%q device=%v file=%q", folder, devID, file)

	svl, err := t.Get(gk, nil)
	if err != nil {
		// We might be called to "remove" a global version that doesn't exist
		// if the first update for the file is already marked invalid.
		return
	}

	var fl VersionList
	err = fl.Unmarshal(svl)
	if err != nil {
		l.Debugln("unmarshal error:", err)
		return
	}

	fl, _, removedAt := fl.pop(devID[:])
	if removedAt == -1 {
		// There is no version for the given device
		return
	}

	if removedAt == 0 {
		// A failure to get the file here is surprising and our
		// global size data will be incorrect until a restart...
		if f, ok := t.getFile(folder, devID[:], file); ok {
			meta.removeFile(protocol.GlobalDeviceID, f)
		}
	}

	if len(fl.Versions) == 0 {
		t.Delete(t.db.keyer.GenerateNeedFileKey(nil, folder, file))
		t.Delete(gk)
		return
	}

	if removedAt == 0 {
		global, ok := t.getFile(folder, fl.Versions[0].Device, file)
		if !ok {
			panic("This file must exist in the db")
		}
		t.updateLocalNeed(folder, file, fl, global)
		meta.addFile(protocol.GlobalDeviceID, global)
	}

	l.Debugf("new global after remove: %v", fl)
	t.Put(gk, mustMarshal(&fl))
}

func (t readWriteTransaction) deleteKeyPrefix(prefix []byte) {
	dbi := t.NewIterator(util.BytesPrefix(prefix), nil)
	for dbi.Next() {
		t.Delete(dbi.Key())
		t.checkFlush()
	}
	dbi.Release()
}

type marshaller interface {
	Marshal() ([]byte, error)
}

func mustMarshal(f marshaller) []byte {
	bs, err := f.Marshal()
	if err != nil {
		panic(err)
	}
	return bs
}
