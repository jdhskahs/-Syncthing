// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package scanner

import (
	"os"
	"path/filepath"

	"github.com/syncthing/syncthing/lib/protocol"
	"github.com/syncthing/syncthing/lib/sync"
)

// The parallell hasher reads FileInfo structures from the inbox, hashes the
// file to populate the Blocks element and sends it to the outbox. A number of
// workers are used in parallel. The outbox will become closed when the inbox
// is closed and all items handled.

func newParallelHasher(algo HashAlgorithm, dir string, blockSize, workers int, outbox, inbox chan protocol.FileInfo, counter *int64, done chan struct{}) {
	wg := sync.NewWaitGroup()
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			hashFiles(algo, dir, blockSize, outbox, inbox, counter)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		if done != nil {
			close(done)
		}
		close(outbox)
	}()
}

func HashFile(algo HashAlgorithm, path string, blockSize int, sizeHint int64, counter *int64) ([]protocol.BlockInfo, error) {
	fd, err := os.Open(path)
	if err != nil {
		l.Debugln("open:", err)
		return []protocol.BlockInfo{}, err
	}
	defer fd.Close()

	if sizeHint == 0 {
		fi, err := fd.Stat()
		if err != nil {
			l.Debugln("stat:", err)
			return []protocol.BlockInfo{}, err
		}
		sizeHint = fi.Size()
	}

	return Blocks(algo, fd, blockSize, sizeHint, counter)
}

func hashFiles(algo HashAlgorithm, dir string, blockSize int, outbox, inbox chan protocol.FileInfo, counter *int64) {
	for f := range inbox {
		if f.IsDirectory() || f.IsDeleted() {
			panic("Bug. Asked to hash a directory or a deleted file.")
		}

		blocks, err := HashFile(algo, filepath.Join(dir, f.Name), blockSize, f.CachedSize, counter)
		if err != nil {
			l.Debugln("hash error:", f.Name, err)
			continue
		}

		f.Blocks = blocks
		outbox <- f
	}
}
