// Copyright (C) 2016 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

// The existence of this file means we get 0% test coverage rather than no
// test coverage at all. Remove when implementing an actual test.

package model_test

import (
	"bytes"
	"testing"

	"github.com/greatroar/rolling"
	"github.com/greatroar/rolling/adler32"
)

var payload = []byte("abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz")

func TestWeakhashFinder(t *testing.T) {
	f := bytes.NewReader(payload)

	finder := rolling.NewScanner(f, adler32.NewRolling(4))
	hashes := []uint32{65143183, 65798547}
	for _, h := range hashes {
		finder.Add(h)
	}

	offsets := map[uint32][]int64{
		65143183: {1, 27, 53, 79},
		65798547: {2, 28, 54, 80},
	}

	b := make([]byte, 4)
	for finder.Scan() {
		h, offset := finder.Match(b)
		if offset != offsets[h][0] {
			t.Fatalf("expected %08x at %d, found it at %d",
				h, offsets[h][0], offset)
		}
		if !bytes.Equal(b, payload[offset:offset+4]) {
			t.Errorf("Not equal at %d: %s != %s", offset, string(b),
				string(payload[offset:offset+4]))
		}

		offsets[h] = offsets[h][1:]
	}

	for h, off := range offsets {
		if len(off) > 0 {
			t.Errorf("didn't find all matches for %08x: %v left", h, off)
		}
	}
}
