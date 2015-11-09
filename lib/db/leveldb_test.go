// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package db

import (
	"bytes"
	"testing"
)

func TestDeviceKey(t *testing.T) {
	fld := []byte("folder6789012345678901234567890123456789012345678901234567890123")
	dev := []byte("device67890123456789012345678901")
	name := []byte("name")

	db := &Instance{}

	key := db.deviceKey(fld, dev, name)

	fld2 := db.deviceKeyFolder(key)
	if bytes.Compare(fld2, fld) != 0 {
		t.Errorf("wrong folder %q != %q", fld2, fld)
	}
	dev2 := db.deviceKeyDevice(key)
	if bytes.Compare(dev2, dev) != 0 {
		t.Errorf("wrong device %q != %q", dev2, dev)
	}
	name2 := db.deviceKeyName(key)
	if bytes.Compare(name2, name) != 0 {
		t.Errorf("wrong name %q != %q", name2, name)
	}
}

func TestGlobalKey(t *testing.T) {
	fld := []byte("folder6789012345678901234567890123456789012345678901234567890123")
	name := []byte("name")

	db := &Instance{}

	key := db.globalKey(fld, name)

	fld2 := db.globalKeyFolder(key)
	if bytes.Compare(fld2, fld) != 0 {
		t.Errorf("wrong folder %q != %q", fld2, fld)
	}
	name2 := db.globalKeyName(key)
	if bytes.Compare(name2, name) != 0 {
		t.Errorf("wrong name %q != %q", name2, name)
	}
}
