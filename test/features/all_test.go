// Copyright (C) 2018 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package features

import (
	"os"
	"testing"

	"github.com/syncthing/syncthing/test"
)

func TestMain(m *testing.M) {
	tempDir := test.NewTemporaryDirectoryForTests()
	defer tempDir.Cleanup()

	os.Exit(m.Run())
}
