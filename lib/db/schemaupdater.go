// Copyright (C) 2018 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package db

import (
	"fmt"
	"strings"

	"github.com/syncthing/syncthing/lib/db/backend"
	"github.com/syncthing/syncthing/lib/protocol"
)

// List of all dbVersion to dbMinSyncthingVersion pairs for convenience
//   0: v0.14.0
//   1: v0.14.46
//   2: v0.14.48
//   3-5: v0.14.49
//   6: v0.14.50
//   7: v0.14.53
//   8-9: v1.4.0
//   10-11: v1.6.0
//   12: v1.7.0
const (
	dbVersion             = 12
	dbMinSyncthingVersion = "v1.7.0"
)

type databaseDowngradeError struct {
	minSyncthingVersion string
}

func (e databaseDowngradeError) Error() string {
	if e.minSyncthingVersion == "" {
		return "newer Syncthing required"
	}
	return fmt.Sprintf("Syncthing %s required", e.minSyncthingVersion)
}

func UpdateSchema(db *Lowlevel) error {
	updater := &schemaUpdater{db}
	return updater.updateSchema()
}

type schemaUpdater struct {
	*Lowlevel
}

func (db *schemaUpdater) updateSchema() error {
	// Updating the schema can touch any and all parts of the database. Make
	// sure we do not run GC concurrently with schema migrations.
	db.gcMut.Lock()
	defer db.gcMut.Unlock()

	miscDB := NewMiscDataNamespace(db.Lowlevel)
	prevVersion, _, err := miscDB.Int64("dbVersion")
	if err != nil {
		return err
	}

	if prevVersion > dbVersion {
		err := databaseDowngradeError{}
		if minSyncthingVersion, ok, dbErr := miscDB.String("dbMinSyncthingVersion"); dbErr != nil {
			return dbErr
		} else if ok {
			err.minSyncthingVersion = minSyncthingVersion
		}
		return err
	}

	if prevVersion == dbVersion {
		return nil
	}

	type migration struct {
		schemaVersion int64
		migration     func(prevVersion int) error
	}
	var migrations = []migration{
		{1, db.updateSchema0to1},
		{2, db.updateSchema1to2},
		{3, db.updateSchema2to3},
		{5, db.updateSchemaTo5},
		{6, db.updateSchema5to6},
		{7, db.updateSchema6to7},
		{9, db.updateSchemaTo9},
		{10, db.updateSchemaTo10},
		{11, db.updateSchemaTo11},
		{12, db.updateSchemaTo12},
	}

	for _, m := range migrations {
		if prevVersion < m.schemaVersion {
			l.Infof("Migrating database to schema version %d...", m.schemaVersion)
			if err := m.migration(int(prevVersion)); err != nil {
				return err
			}
		}
	}

	if err := miscDB.PutInt64("dbVersion", dbVersion); err != nil {
		return err
	}
	if err := miscDB.PutString("dbMinSyncthingVersion", dbMinSyncthingVersion); err != nil {
		return err
	}

	l.Infoln("Compacting database after migration...")
	return db.Compact()
}

func (db *schemaUpdater) updateSchema0to1(_ int) error {
	t, err := db.newReadWriteTransaction()
	if err != nil {
		return err
	}
	defer t.close()

	dbi, err := t.NewPrefixIterator([]byte{KeyTypeDevice})
	if err != nil {
		return err
	}
	defer dbi.Release()

	symlinkConv := 0
	changedFolders := make(map[string]struct{})
	ignAdded := 0
	meta := newMetadataTracker() // dummy metadata tracker
	var gk, buf []byte

	for dbi.Next() {
		folder, ok := db.keyer.FolderFromDeviceFileKey(dbi.Key())
		if !ok {
			// not having the folder in the index is bad; delete and continue
			if err := t.Delete(dbi.Key()); err != nil {
				return err
			}
			continue
		}
		device, ok := db.keyer.DeviceFromDeviceFileKey(dbi.Key())
		if !ok {
			// not having the device in the index is bad; delete and continue
			if err := t.Delete(dbi.Key()); err != nil {
				return err
			}
			continue
		}
		name := db.keyer.NameFromDeviceFileKey(dbi.Key())

		// Remove files with absolute path (see #4799)
		if strings.HasPrefix(string(name), "/") {
			if _, ok := changedFolders[string(folder)]; !ok {
				changedFolders[string(folder)] = struct{}{}
			}
			gk, err = db.keyer.GenerateGlobalVersionKey(gk, folder, name)
			if err != nil {
				return err
			}
			// Purposely pass nil file name to remove from global list,
			// but don't touch meta and needs
			buf, err = t.removeFromGlobal(gk, buf, folder, device, nil, nil)
			if err != nil && err != errEntryFromGlobalMissing {
				return err
			}
			if err := t.Delete(dbi.Key()); err != nil {
				return err
			}
			continue
		}

		// Change SYMLINK_FILE and SYMLINK_DIRECTORY types to the current SYMLINK
		// type (previously SYMLINK_UNKNOWN). It does this for all devices, both
		// local and remote, and does not reset delta indexes. It shouldn't really
		// matter what the symlink type is, but this cleans it up for a possible
		// future when SYMLINK_FILE and SYMLINK_DIRECTORY are no longer understood.
		var f protocol.FileInfo
		if err := f.Unmarshal(dbi.Value()); err != nil {
			// probably can't happen
			continue
		}
		if f.Type == protocol.FileInfoTypeDeprecatedSymlinkDirectory || f.Type == protocol.FileInfoTypeDeprecatedSymlinkFile {
			f.Type = protocol.FileInfoTypeSymlink
			bs, err := f.Marshal()
			if err != nil {
				panic("can't happen: " + err.Error())
			}
			if err := t.Put(dbi.Key(), bs); err != nil {
				return err
			}
			symlinkConv++
		}

		// Add invalid files to global list
		if f.IsInvalid() {
			gk, err = db.keyer.GenerateGlobalVersionKey(gk, folder, name)
			if err != nil {
				return err
			}
			if buf, ok, err = t.updateGlobal(gk, buf, folder, device, f, meta); err != nil {
				return err
			} else if ok {
				if _, ok = changedFolders[string(folder)]; !ok {
					changedFolders[string(folder)] = struct{}{}
				}
				ignAdded++
			}
		}
		if err := t.Checkpoint(); err != nil {
			return err
		}
	}

	for folder := range changedFolders {
		if err := db.dropFolderMeta([]byte(folder)); err != nil {
			return err
		}
	}
	return t.Commit()
}

// updateSchema1to2 introduces a sequenceKey->deviceKey bucket for local items
// to allow iteration in sequence order (simplifies sending indexes).
func (db *schemaUpdater) updateSchema1to2(_ int) error {
	t, err := db.newReadWriteTransaction()
	if err != nil {
		return err
	}
	defer t.close()

	var sk []byte
	var dk []byte
	for _, folderStr := range db.ListFolders() {
		folder := []byte(folderStr)
		var putErr error
		err := t.withHave(folder, protocol.LocalDeviceID[:], nil, true, func(f FileIntf) bool {
			sk, putErr = db.keyer.GenerateSequenceKey(sk, folder, f.SequenceNo())
			if putErr != nil {
				return false
			}
			dk, putErr = db.keyer.GenerateDeviceFileKey(dk, folder, protocol.LocalDeviceID[:], []byte(f.FileName()))
			if putErr != nil {
				return false
			}
			putErr = t.Put(sk, dk)
			return putErr == nil
		})
		if putErr != nil {
			return putErr
		}
		if err != nil {
			return err
		}
	}
	return t.Commit()
}

// updateSchema2to3 introduces a needKey->nil bucket for locally needed files.
func (db *schemaUpdater) updateSchema2to3(_ int) error {
	t, err := db.newReadWriteTransaction()
	if err != nil {
		return err
	}
	defer t.close()

	var nk []byte
	var dk []byte
	for _, folderStr := range db.ListFolders() {
		folder := []byte(folderStr)
		var putErr error
		err := t.withGlobal(folder, nil, true, func(f FileIntf) bool {
			name := []byte(f.FileName())
			dk, putErr = db.keyer.GenerateDeviceFileKey(dk, folder, protocol.LocalDeviceID[:], name)
			if putErr != nil {
				return false
			}
			var v protocol.Vector
			haveFile, ok, err := t.getFileTrunc(dk, true)
			if err != nil {
				putErr = err
				return false
			}
			if ok {
				v = haveFile.FileVersion()
			}
			fv := FileVersion{
				Version: f.FileVersion(),
				Invalid: f.IsInvalid(),
				Deleted: f.IsDeleted(),
			}
			if !need(fv, ok, v) {
				return true
			}
			nk, putErr = t.keyer.GenerateNeedFileKey(nk, folder, []byte(f.FileName()))
			if putErr != nil {
				return false
			}
			putErr = t.Put(nk, nil)
			return putErr == nil
		})
		if putErr != nil {
			return putErr
		}
		if err != nil {
			return err
		}
	}
	return t.Commit()
}

// updateSchemaTo5 resets the need bucket due to bugs existing in the v0.14.49
// release candidates (dbVersion 3 and 4)
// https://github.com/syncthing/syncthing/issues/5007
// https://github.com/syncthing/syncthing/issues/5053
func (db *schemaUpdater) updateSchemaTo5(prevVersion int) error {
	if prevVersion != 3 && prevVersion != 4 {
		return nil
	}

	t, err := db.newReadWriteTransaction()
	if err != nil {
		return err
	}
	var nk []byte
	for _, folderStr := range db.ListFolders() {
		nk, err = db.keyer.GenerateNeedFileKey(nk, []byte(folderStr), nil)
		if err != nil {
			return err
		}
		if err := t.deleteKeyPrefix(nk[:keyPrefixLen+keyFolderLen]); err != nil {
			return err
		}
	}
	if err := t.Commit(); err != nil {
		return err
	}

	return db.updateSchema2to3(2)
}

func (db *schemaUpdater) updateSchema5to6(_ int) error {
	// For every local file with the Invalid bit set, clear the Invalid bit and
	// set LocalFlags = FlagLocalIgnored.

	t, err := db.newReadWriteTransaction()
	if err != nil {
		return err
	}
	defer t.close()

	var dk []byte

	for _, folderStr := range db.ListFolders() {
		folder := []byte(folderStr)
		var iterErr error
		err := t.withHave(folder, protocol.LocalDeviceID[:], nil, false, func(f FileIntf) bool {
			if !f.IsInvalid() {
				return true
			}

			fi := f.(protocol.FileInfo)
			fi.RawInvalid = false
			fi.LocalFlags = protocol.FlagLocalIgnored
			bs, _ := fi.Marshal()

			dk, iterErr = db.keyer.GenerateDeviceFileKey(dk, folder, protocol.LocalDeviceID[:], []byte(fi.Name))
			if iterErr != nil {
				return false
			}
			if iterErr = t.Put(dk, bs); iterErr != nil {
				return false
			}
			iterErr = t.Checkpoint()
			return iterErr == nil
		})
		if iterErr != nil {
			return iterErr
		}
		if err != nil {
			return err
		}
	}
	return t.Commit()
}

// updateSchema6to7 checks whether all currently locally needed files are really
// needed and removes them if not.
func (db *schemaUpdater) updateSchema6to7(_ int) error {
	t, err := db.newReadWriteTransaction()
	if err != nil {
		return err
	}
	defer t.close()

	var gk []byte
	var nk []byte

	for _, folderStr := range db.ListFolders() {
		folder := []byte(folderStr)
		var delErr error
		err := t.withNeedLocal(folder, false, func(f FileIntf) bool {
			name := []byte(f.FileName())
			gk, delErr = db.keyer.GenerateGlobalVersionKey(gk, folder, name)
			if delErr != nil {
				return false
			}
			svl, err := t.Get(gk)
			if err != nil {
				// If there is no global list, we hardly need it.
				key, err := t.keyer.GenerateNeedFileKey(nk, folder, name)
				if err != nil {
					delErr = err
					return false
				}
				delErr = t.Delete(key)
				return delErr == nil
			}
			var fl VersionList
			err = fl.Unmarshal(svl)
			if err != nil {
				// This can't happen, but it's ignored everywhere else too,
				// so lets not act on it.
				return true
			}
			globalFV := FileVersion{
				Version: f.FileVersion(),
				Invalid: f.IsInvalid(),
				Deleted: f.IsDeleted(),
			}

			if localFV, haveLocalFV := fl.Get(protocol.LocalDeviceID[:]); !need(globalFV, haveLocalFV, localFV.Version) {
				key, err := t.keyer.GenerateNeedFileKey(nk, folder, name)
				if err != nil {
					delErr = err
					return false
				}
				delErr = t.Delete(key)
			}
			return delErr == nil
		})
		if delErr != nil {
			return delErr
		}
		if err != nil {
			return err
		}
		if err := t.Checkpoint(); err != nil {
			return err
		}
	}
	return t.Commit()
}

func (db *schemaUpdater) updateSchemaTo9(prev int) error {
	// Loads and rewrites all files with blocks, to deduplicate block lists.

	t, err := db.newReadWriteTransaction()
	if err != nil {
		return err
	}
	defer t.close()

	if err := db.rewriteFiles(t); err != nil {
		return err
	}

	db.recordTime(indirectGCTimeKey)

	return t.Commit()
}

func (db *schemaUpdater) rewriteFiles(t readWriteTransaction) error {
	it, err := t.NewPrefixIterator([]byte{KeyTypeDevice})
	if err != nil {
		return err
	}
	for it.Next() {
		intf, err := t.unmarshalTrunc(it.Value(), false)
		if backend.IsNotFound(err) {
			// Unmarshal error due to missing parts (block list), probably
			// due to a bad migration in a previous RC. Drop this key, as
			// getFile would anyway return this as a "not found" in the
			// normal flow of things.
			if err := t.Delete(it.Key()); err != nil {
				return err
			}
			continue
		} else if err != nil {
			return err
		}
		fi := intf.(protocol.FileInfo)
		if fi.Blocks == nil {
			continue
		}
		if err := t.putFile(it.Key(), fi, false); err != nil {
			return err
		}
		if err := t.Checkpoint(); err != nil {
			return err
		}
	}
	it.Release()
	return it.Error()
}

func (db *schemaUpdater) updateSchemaTo10(_ int) error {
	t, err := db.newReadWriteTransaction()
	if err != nil {
		return err
	}
	defer t.close()

	var buf []byte

	for _, folderStr := range db.ListFolders() {
		folder := []byte(folderStr)

		buf, err = t.keyer.GenerateGlobalVersionKey(buf, folder, nil)
		if err != nil {
			return err
		}
		buf = globalVersionKey(buf).WithoutName()
		dbi, err := t.NewPrefixIterator(buf)
		if err != nil {
			return err
		}
		defer dbi.Release()

		for dbi.Next() {
			var vl VersionList
			if err := vl.Unmarshal(dbi.Value()); err != nil {
				return err
			}

			changed := false
			name := t.keyer.NameFromGlobalVersionKey(dbi.Key())

			for i, fv := range vl.Versions {
				buf, err = t.keyer.GenerateDeviceFileKey(buf, folder, fv.Device, name)
				if err != nil {
					return err
				}
				f, ok, err := t.getFileTrunc(buf, true)
				if !ok {
					return errEntryFromGlobalMissing
				}
				if err != nil {
					return err
				}
				if f.IsDeleted() {
					vl.Versions[i].Deleted = true
					changed = true
				}
			}

			if changed {
				if err := t.Put(dbi.Key(), mustMarshal(&vl)); err != nil {
					return err
				}
				if err := t.Checkpoint(); err != nil {
					return err
				}
			}
		}
		dbi.Release()
	}

	// Trigger metadata recalc
	if err := t.deleteKeyPrefix([]byte{KeyTypeFolderMeta}); err != nil {
		return err
	}

	return t.Commit()
}

func (db *schemaUpdater) updateSchemaTo11(_ int) error {
	// Populates block list map for every folder.

	t, err := db.newReadWriteTransaction()
	if err != nil {
		return err
	}
	defer t.close()

	var dk []byte
	for _, folderStr := range db.ListFolders() {
		folder := []byte(folderStr)
		var putErr error
		err := t.withHave(folder, protocol.LocalDeviceID[:], nil, true, func(fi FileIntf) bool {
			f := fi.(FileInfoTruncated)
			if f.IsDirectory() || f.IsDeleted() || f.IsSymlink() || f.IsInvalid() || f.BlocksHash == nil {
				return true
			}

			name := []byte(f.FileName())
			dk, putErr = db.keyer.GenerateBlockListMapKey(dk, folder, f.BlocksHash, name)
			if putErr != nil {
				return false
			}

			if putErr = t.Put(dk, nil); putErr != nil {
				return false
			}
			putErr = t.Checkpoint()
			return putErr == nil
		})
		if putErr != nil {
			return putErr
		}
		if err != nil {
			return err
		}
	}
	return t.Commit()
}

func (db *schemaUpdater) updateSchemaTo12(_ int) error {
	// Loads and rewrites all files, to deduplicate version vectors.

	t, err := db.newReadWriteTransaction()
	if err != nil {
		return err
	}
	defer t.close()

	if err := db.rewriteFiles(t); err != nil {
		return err
	}

	return t.Commit()
}
