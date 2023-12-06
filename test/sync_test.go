// Copyright (C) 2023 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package integration

import (
	"fmt"
	"testing"

	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/fs"
	"github.com/syncthing/syncthing/lib/rand"
	"github.com/syncthing/syncthing/lib/rc"
)

func TestSyncTwoDevices(t *testing.T) {
	t.Parallel()

	// Create a source folder with some data in it.
	srcDir := generateTree(t, 100)
	// Create a destination folder to hold the synced data.
	dstDir := t.TempDir()

	// The folder needs an ID.
	folderID := rand.String(8)

	// Start the source device.
	src := startInstance(t)
	srcAPI := rc.NewAPI(src.apiAddress, src.apiKey)

	// Start the destination device.
	dst := startInstance(t)
	dstAPI := rc.NewAPI(dst.apiAddress, dst.apiKey)

	// Add the peer device to each device. Hard code the sync addresses to
	// speed things up.
	if err := srcAPI.Post("/rest/config/devices", &config.DeviceConfiguration{
		DeviceID:  dst.deviceID,
		Addresses: []string{fmt.Sprintf("tcp://127.0.0.1:%d", dst.tcpPort)},
	}, nil); err != nil {
		t.Fatal(err)
	}
	if err := dstAPI.Post("/rest/config/devices", &config.DeviceConfiguration{
		DeviceID:  src.deviceID,
		Addresses: []string{fmt.Sprintf("tcp://127.0.0.1:%d", src.tcpPort)},
	}, nil); err != nil {
		t.Fatal(err)
	}

	// Add the folder to both devices.
	if err := srcAPI.Post("/rest/config/folders", &config.FolderConfiguration{
		ID:             folderID,
		Path:           srcDir,
		FilesystemType: fs.FilesystemTypeBasic,
		Type:           config.FolderTypeSendReceive,
		Devices: []config.FolderDeviceConfiguration{
			{DeviceID: src.deviceID},
			{DeviceID: dst.deviceID},
		},
	}, nil); err != nil {
		t.Fatal(err)
	}
	if err := dstAPI.Post("/rest/config/folders", &config.FolderConfiguration{
		ID:             folderID,
		Path:           dstDir,
		FilesystemType: fs.FilesystemTypeBasic,
		Type:           config.FolderTypeSendReceive,
		Devices: []config.FolderDeviceConfiguration{
			{DeviceID: src.deviceID},
			{DeviceID: dst.deviceID},
		},
	}, nil); err != nil {
		t.Fatal(err)
	}

	// Listen to events on the destination side. Watch for the folder
	// starting to sync and then becoming idle. At that point we should be
	// done.
	lastEventID := 0
	didStartSyncing := false
loop:
	for {
		events, err := dstAPI.Events(lastEventID)
		if err != nil {
			t.Fatal(err)
		}
		for _, ev := range events {
			switch ev.Type {
			case "StateChanged":
				folder := ev.Data.(map[string]any)["folder"].(string)
				to := ev.Data.(map[string]any)["to"].(string)
				if folder == folderID && to == "syncing" {
					didStartSyncing = true
				}
				if folder == folderID && to == "idle" && didStartSyncing {
					break loop
				}
			}
			lastEventID = ev.ID
		}
	}

	// Check that the destination folder now contains the same files as the source folder.
	compareTrees(t, srcDir, dstDir)
}
