// Copyright (C) 2020 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

// +build noquic go1.16

package connections

func init() {
	for _, scheme := range []string{"quic", "quic4", "quic6"} {
		listeners[scheme] = invalidListener{err: errUnsupported}
		dialers[scheme] = invalidDialer{err: errUnsupported}
	}
}
