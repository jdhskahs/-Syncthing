// Copyright (C) 2014 Jakob Borg and Contributors (see the CONTRIBUTORS file).
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program. If not, see <http://www.gnu.org/licenses/>.

package model

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/syncthing/syncthing/internal/config"
	"github.com/syncthing/syncthing/internal/events"
	"github.com/syncthing/syncthing/internal/osutil"
	"github.com/syncthing/syncthing/internal/protocol"
	"github.com/syncthing/syncthing/internal/scanner"
	"github.com/syncthing/syncthing/internal/versioner"
)

// TODO: Stop on errors

const (
	copiersPerFolder   = 1
	pullersPerFolder   = 16
	finishersPerFolder = 2
	pauseIntv          = 60 * time.Second
	nextPullIntv       = 10 * time.Second
	checkPullIntv      = 1 * time.Second
)

// A pullBlockState is passed to the puller routine for each block that needs
// to be fetched.
type pullBlockState struct {
	*sharedPullerState
	block protocol.BlockInfo
}

// A copyBlocksState is passed to copy routine if the file has blocks to be
// copied from the original.
type copyBlocksState struct {
	*sharedPullerState
	blocks []protocol.BlockInfo
}

var (
	activity    = newDeviceActivity()
	errNoDevice = errors.New("no available source device")
)

type Puller struct {
	folder    string
	dir       string
	scanIntv  time.Duration
	model     *Model
	stop      chan struct{}
	versioner versioner.Versioner
}

// Serve will run scans and pulls. It will return when Stop()ed or on a
// critical error.
func (p *Puller) Serve() {
	if debug {
		l.Debugln(p, "starting")
		defer l.Debugln(p, "exiting")
	}

	p.stop = make(chan struct{})

	pullTimer := time.NewTimer(checkPullIntv)
	scanTimer := time.NewTimer(time.Millisecond) // The first scan should be done immediately.

	defer func() {
		pullTimer.Stop()
		scanTimer.Stop()
		// TODO: Should there be an actual FolderStopped state?
		p.model.setState(p.folder, FolderIdle)
	}()

	var prevVer uint64

	// Clean out old temporaries before we start pulling
	p.clean()

	// We don't start pulling files until a scan has been completed.
	initialScanCompleted := false

loop:
	for {
		select {
		case <-p.stop:
			return

		// TODO: We could easily add a channel here for notifications from
		// Index(), so that we immediately start a pull when new index
		// information is available. Before that though, I'd like to build a
		// repeatable benchmark of how long it takes to sync a change from
		// device A to device B, so we have something to work against.
		case <-pullTimer.C:
			if !initialScanCompleted {
				// How did we even get here?
				if debug {
					l.Debugln(p, "skip (initial)")
				}
				pullTimer.Reset(nextPullIntv)
				continue
			}

			// RemoteLocalVersion() is a fast call, doesn't touch the database.
			curVer := p.model.RemoteLocalVersion(p.folder)
			if curVer == prevVer {
				if debug {
					l.Debugln(p, "skip (curVer == prevVer)", prevVer)
				}
				pullTimer.Reset(checkPullIntv)
				continue
			}

			if debug {
				l.Debugln(p, "pulling", prevVer, curVer)
			}
			p.model.setState(p.folder, FolderSyncing)
			tries := 0
			for {
				tries++
				changed := p.pullerIteration(copiersPerFolder, pullersPerFolder, finishersPerFolder)
				if debug {
					l.Debugln(p, "changed", changed)
				}

				if changed == 0 {
					// No files were changed by the puller, so we are in
					// sync. Remember the local version number and
					// schedule a resync a little bit into the future.

					if lv := p.model.RemoteLocalVersion(p.folder); lv < curVer {
						// There's a corner case where the device we needed
						// files from disconnected during the puller
						// iteration. The files will have been removed from
						// the index, so we've concluded that we don't need
						// them, but at the same time we have the local
						// version that includes those files in curVer. So we
						// catch the case that localVersion might have
						// decresed here.
						l.Debugln(p, "adjusting curVer", lv)
						curVer = lv
					}
					prevVer = curVer
					pullTimer.Reset(nextPullIntv)
					break
				}

				if tries > 10 {
					// We've tried a bunch of times to get in sync, but
					// we're not making it. Probably there are write
					// errors preventing us. Flag this with a warning and
					// wait a bit longer before retrying.
					l.Warnf("Folder %q isn't making progress - check logs for possible root cause. Pausing puller for %v.", p.folder, pauseIntv)
					pullTimer.Reset(pauseIntv)
					break
				}
			}
			p.model.setState(p.folder, FolderIdle)

		// The reason for running the scanner from within the puller is that
		// this is the easiest way to make sure we are not doing both at the
		// same time.
		case <-scanTimer.C:
			if debug {
				l.Debugln(p, "rescan")
			}
			p.model.setState(p.folder, FolderScanning)
			if err := p.model.ScanFolder(p.folder); err != nil {
				invalidateFolder(p.model.cfg, p.folder, err)
				break loop
			}
			p.model.setState(p.folder, FolderIdle)
			scanTimer.Reset(p.scanIntv)
			if !initialScanCompleted {
				l.Infoln("Completed initial scan (rw) of folder", p.folder)
				initialScanCompleted = true
			}
		}
	}
}

func (p *Puller) Stop() {
	close(p.stop)
}

func (p *Puller) String() string {
	return fmt.Sprintf("puller/%s@%p", p.folder, p)
}

// pullerIteration runs a single puller iteration for the given folder and
// returns the number items that should have been synced (even those that
// might have failed). One puller iteration handles all files currently
// flagged as needed in the folder. The specified number of copier, puller and
// finisher routines are used. It's seldom efficient to use more than one
// copier routine, while multiple pullers are essential and multiple finishers
// may be useful (they are primarily CPU bound due to hashing).
func (p *Puller) pullerIteration(ncopiers, npullers, nfinishers int) int {
	pullChan := make(chan pullBlockState)
	copyChan := make(chan copyBlocksState)
	finisherChan := make(chan *sharedPullerState)

	var wg sync.WaitGroup
	var doneWg sync.WaitGroup

	for i := 0; i < ncopiers; i++ {
		wg.Add(1)
		go func() {
			// copierRoutine finishes when copyChan is closed
			p.copierRoutine(copyChan, finisherChan)
			wg.Done()
		}()
	}

	for i := 0; i < npullers; i++ {
		wg.Add(1)
		go func() {
			// pullerRoutine finishes when pullChan is closed
			p.pullerRoutine(pullChan, finisherChan)
			wg.Done()
		}()
	}

	for i := 0; i < nfinishers; i++ {
		doneWg.Add(1)
		// finisherRoutine finishes when finisherChan is closed
		go func() {
			p.finisherRoutine(finisherChan)
			doneWg.Done()
		}()
	}

	p.model.fmut.RLock()
	files := p.model.folderFiles[p.folder]
	p.model.fmut.RUnlock()

	// !!!
	// WithNeed takes a database snapshot (by necessity). By the time we've
	// handled a bunch of files it might have become out of date and we might
	// be attempting to sync with an old version of a file...
	// !!!

	var dirsChanged, dirsDeleted, filesDeleted []protocol.FileInfo
	filesChanged := map[string]protocol.FileInfo{}

	files.WithNeed(protocol.LocalDeviceID, func(intf protocol.FileIntf) bool {
		file := intf.(protocol.FileInfo)

		events.Default.Log(events.ItemStarted, map[string]string{
			"folder": p.folder,
			"item":   file.Name,
		})

		if debug {
			l.Debugln(p, "handling", file.Name)
		}

		switch {
		case protocol.IsDirectory(file.Flags) && protocol.IsDeleted(file.Flags):
			// A deleted directory
			dirsDeleted = append(dirsDeleted, file)
		case protocol.IsDirectory(file.Flags):
			// A new or changed directory
			dirsChanged = append(dirsChanged, file)
		case protocol.IsDeleted(file.Flags):
			// A deleted file
			filesDeleted = append(filesDeleted, file)
		default:
			// A new or changed file. This is the only case where we will do
			// stuff in the background; the other three + copying will be done
			// synchronously.
			filesChanged[file.Name] = file
		}

		return true
	})

	// Create all directories first, as we need to make sure they exist prior
	// trying to copy something into them.
	for _, dir := range dirsChanged {
		p.handleDir(dir)
	}

	buckets := make(map[string][]protocol.FileInfo, len(filesChanged))
	// Setup a minimal search space
	for _, file := range filesChanged {
		if len(file.Blocks) > 0 {
			buckets[string(file.Blocks[0].Hash)] = []protocol.FileInfo{}
		}
	}

	// Figure out which of the existing files could be potential candidates
	files.WithHave(protocol.LocalDeviceID, func(intf protocol.FileIntf) bool {
		file := intf.(protocol.FileInfo)
		if len(file.Blocks) > 0 {
			key := string(file.Blocks[0].Hash)
			_, ok := buckets[key]
			if ok {
				if debug {
					l.Debugln(file.Name, "marked as potential copy candidate")
				}
				buckets[key] = append(buckets[key], file)
			}
		}
		return true
	})

	filesCopied := [][]protocol.FileInfo{}

	// For every file that we are changing, check if any of the existing files
	// is the desired result. If any, mark the pair as copies.
	// In an ideal world this should be more efficient, as we should be able to
	// build files from multiple different existing files.
nextChange:
	for _, file := range filesChanged {
		if len(file.Blocks) == 0 {
			continue
		}
		for _, candidate := range buckets[string(file.Blocks[0].Hash)] {
			if scanner.BlocksEqual(file.Blocks, candidate.Blocks) {
				filesCopied = append(filesCopied, []protocol.FileInfo{
					file, candidate,
				})
				continue nextChange
			}
		}
	}

	// Handle copies
	for _, pair := range filesCopied {
		err := p.copyFile(pair[1], pair[0])
		// Discard the file from filesChanged list, as we already know that the
		// blocks are are identical, and we've done the work handleFile would
		// do.
		if err == nil {
			delete(filesChanged, pair[0].Name)
		}
	}

	// Do the rest of the work
	for _, file := range filesChanged {
		p.handleFile(file, copyChan, pullChan)
	}

	for _, file := range filesDeleted {
		p.deleteFile(file)
	}

	for _, dir := range dirsDeleted {
		p.deleteDir(dir)
	}

	// Signal copy and puller routines that we are done with the in data for
	// this iteration
	close(copyChan)
	close(pullChan)

	// Wait for them to finish, then signal the finisher chan that there will
	// be no more input.
	wg.Wait()
	close(finisherChan)

	// Wait for the finisherChan to finish.
	doneWg.Wait()

	changed := len(dirsChanged) + len(dirsDeleted) + len(filesChanged)
	changed += len(filesCopied) + len(filesDeleted)
	return changed
}

// handleDir creates or updates the given directory
func (p *Puller) handleDir(file protocol.FileInfo) {
	realName := filepath.Join(p.dir, file.Name)
	mode := os.FileMode(file.Flags & 0777)

	if debug {
		curFile := p.model.CurrentFolderFile(p.folder, file.Name)
		l.Debugf("need dir\n\t%v\n\t%v", file, curFile)
	}

	if info, err := os.Stat(realName); err != nil {
		if os.IsNotExist(err) {
			// The directory doesn't exist, so we create it with the right
			// mode bits from the start.

			mkdir := func(path string) error {
				// We declare a function that acts on only the path name, so
				// we can pass it to InWritableDir. We use a regular Mkdir and
				// not MkdirAll because the parent should already exist.
				return os.Mkdir(path, mode)
			}

			if err = osutil.InWritableDir(mkdir, realName); err == nil {
				p.model.updateLocal(p.folder, file)
			} else {
				l.Infof("Puller (folder %q, file %q): %v", p.folder, file.Name, err)
			}
			return
		}

		// Weird error when stat()'ing the dir. Probably won't work to do
		// anything else with it if we can't even stat() it.
		l.Infof("Puller (folder %q, file %q): %v", p.folder, file.Name, err)
		return
	} else if !info.IsDir() {
		l.Infof("Puller (folder %q, file %q): should be dir, but is not", p.folder, file.Name)
		return
	}

	// The directory already exists, so we just correct the mode bits. (We
	// don't handle modification times on directories, because that sucks...)
	// It's OK to change mode bits on stuff within non-writable directories.

	if err := os.Chmod(realName, mode); err == nil {
		p.model.updateLocal(p.folder, file)
	} else {
		l.Infof("Puller (folder %q, file %q): %v", p.folder, file.Name, err)
	}
}

// deleteDir attempts to delete the given directory
func (p *Puller) deleteDir(file protocol.FileInfo) {
	realName := filepath.Join(p.dir, file.Name)
	err := osutil.InWritableDir(os.Remove, realName)
	if err == nil || os.IsNotExist(err) {
		p.model.updateLocal(p.folder, file)
	}
}

// deleteFile attempts to delete the given file
func (p *Puller) deleteFile(file protocol.FileInfo) {
	realName := filepath.Join(p.dir, file.Name)

	var err error
	if p.versioner != nil {
		err = osutil.InWritableDir(p.versioner.Archive, realName)
	} else {
		err = osutil.InWritableDir(os.Remove, realName)
	}

	if err != nil {
		l.Infof("Puller (folder %q, file %q): delete: %v", p.folder, file.Name, err)
	} else {
		p.model.updateLocal(p.folder, file)
	}
}

// copyFile attempts to copy from source to destination, maintaining the
// flags specified by destination, returning an error on failure.
func (p *Puller) copyFile(src, dst protocol.FileInfo) error {
	srcName := filepath.Join(p.dir, src.Name)

	if debug {
		l.Debugln(p, "copying", src.Name, "to", dst.Name)
	}

	// Hackish curry function
	curry := func(dstName string) error {
		return osutil.Copy(srcName, dstName)
	}

	err := osutil.InWritableDir(curry, filepath.Join(p.dir, dst.Name))
	if err != nil {
		l.Infof("Puller (folder %q, src %q, dst %q): copy: %v", p.folder, src.Name, dst.Name, err)
		return err
	}

	p.shortcutFile(dst)
	return nil
}

// handleFile queues the copies and pulls as necessary for a single new or
// changed file.
func (p *Puller) handleFile(file protocol.FileInfo, copyChan chan<- copyBlocksState, pullChan chan<- pullBlockState) {
	curFile := p.model.CurrentFolderFile(p.folder, file.Name)
	copyBlocks, pullBlocks := scanner.BlockDiff(curFile.Blocks, file.Blocks)

	if len(copyBlocks) == len(curFile.Blocks) && len(pullBlocks) == 0 {
		// We are supposed to copy the entire file, and then fetch nothing. We
		// are only updating metadata, so we don't actually *need* to make the
		// copy.
		if debug {
			l.Debugln(p, "taking shortcut on", file.Name)
		}
		p.shortcutFile(file)
		return
	}

	// Figure out the absolute filenames we need once and for all
	tempName := filepath.Join(p.dir, defTempNamer.TempName(file.Name))
	realName := filepath.Join(p.dir, file.Name)

	var reuse bool

	// Check for an old temporary file which might have some blocks we could
	// reuse.
	tempBlocks, err := scanner.HashFile(tempName, protocol.BlockSize)
	if err == nil {
		// Check for any reusable blocks in the temp file
		tempCopyBlocks, _ := scanner.BlockDiff(tempBlocks, file.Blocks)

		// block.String() returns a string unique to the block
		existingBlocks := make(map[string]bool, len(tempCopyBlocks))
		for _, block := range tempCopyBlocks {
			existingBlocks[block.String()] = true
		}

		// Since the blocks are already there, we don't need to copy them
		// nor we need to pull them, hence discard blocks which are already
		// there, if they are exactly the same...
		var newCopyBlocks []protocol.BlockInfo
		for _, block := range copyBlocks {
			_, ok := existingBlocks[block.String()]
			if !ok {
				newCopyBlocks = append(newCopyBlocks, block)
			}
		}

		var newPullBlocks []protocol.BlockInfo
		for _, block := range pullBlocks {
			_, ok := existingBlocks[block.String()]
			if !ok {
				newPullBlocks = append(newPullBlocks, block)
			}
		}

		// If any blocks could be reused, let the sharedpullerstate know
		// which flags it is expected to set on the file.
		// Also update the list of work for the routines.
		if len(copyBlocks) != len(newCopyBlocks) || len(pullBlocks) != len(newPullBlocks) {
			reuse = true
			copyBlocks = newCopyBlocks
			pullBlocks = newPullBlocks
		} else {
			// Otherwise, discard the file ourselves in order for the
			// sharedpuller not to panic when it fails to exlusively create a
			// file which already exists
			os.Remove(tempName)
		}
	}

	s := sharedPullerState{
		file:       file,
		folder:     p.folder,
		tempName:   tempName,
		realName:   realName,
		pullNeeded: len(pullBlocks),
		reuse:      reuse,
	}
	if len(copyBlocks) > 0 {
		s.copyNeeded = 1
	}

	if debug {
		l.Debugf("%v need file %s; copy %d, pull %d", p, file.Name, len(copyBlocks), len(pullBlocks))
	}

	if len(copyBlocks) > 0 {
		cs := copyBlocksState{
			sharedPullerState: &s,
			blocks:            copyBlocks,
		}
		copyChan <- cs
	}

	if len(pullBlocks) > 0 {
		for _, block := range pullBlocks {
			ps := pullBlockState{
				sharedPullerState: &s,
				block:             block,
			}
			pullChan <- ps
		}
	}
}

// shortcutFile sets file mode and modification time, when that's the only
// thing that has changed.
func (p *Puller) shortcutFile(file protocol.FileInfo) {
	realName := filepath.Join(p.dir, file.Name)
	err := os.Chmod(realName, os.FileMode(file.Flags&0777))
	if err != nil {
		l.Infof("Puller (folder %q, file %q): shortcut: %v", p.folder, file.Name, err)
		return
	}

	t := time.Unix(file.Modified, 0)
	err = os.Chtimes(realName, t, t)
	if err != nil {
		l.Infof("Puller (folder %q, file %q): shortcut: %v", p.folder, file.Name, err)
		return
	}

	p.model.updateLocal(p.folder, file)
}

// copierRoutine reads pullerStates until the in channel closes and performs
// the relevant copy.
func (p *Puller) copierRoutine(in <-chan copyBlocksState, out chan<- *sharedPullerState) {
	buf := make([]byte, protocol.BlockSize)

nextFile:
	for state := range in {
		dstFd, err := state.tempFile()
		if err != nil {
			// Nothing more to do for this failed file (the error was logged
			// when it happened)
			continue nextFile
		}

		srcFd, err := state.sourceFile()
		if err != nil {
			// As above
			continue nextFile
		}

		for _, block := range state.blocks {
			buf = buf[:int(block.Size)]

			_, err = srcFd.ReadAt(buf, block.Offset)
			if err != nil {
				state.earlyClose("src read", err)
				srcFd.Close()
				continue nextFile
			}

			_, err = dstFd.WriteAt(buf, block.Offset)
			if err != nil {
				state.earlyClose("dst write", err)
				srcFd.Close()
				continue nextFile
			}
		}

		srcFd.Close()
		state.copyDone()
		out <- state.sharedPullerState
	}
}

func (p *Puller) pullerRoutine(in <-chan pullBlockState, out chan<- *sharedPullerState) {
nextBlock:
	for state := range in {
		if state.failed() != nil {
			continue nextBlock
		}

		// Select the least busy device to pull the block frop.model. If we found no
		// feasible device at all, fail the block (and in the long run, the
		// file).
		potentialDevices := p.model.availability(p.folder, state.file.Name)
		selected := activity.leastBusy(potentialDevices)
		if selected == (protocol.DeviceID{}) {
			state.earlyClose("pull", errNoDevice)
			continue nextBlock
		}

		// Get an fd to the temporary file. Tehcnically we don't need it until
		// after fetching the block, but if we run into an error here there is
		// no point in issuing the request to the network.
		fd, err := state.tempFile()
		if err != nil {
			continue nextBlock
		}

		// Fetch the block, while marking the selected device as in use so that
		// leastBusy can select another device when someone else asks.
		activity.using(selected)
		buf, err := p.model.requestGlobal(selected, p.folder, state.file.Name, state.block.Offset, int(state.block.Size), state.block.Hash)
		activity.done(selected)
		if err != nil {
			state.earlyClose("pull", err)
			continue nextBlock
		}

		// Save the block data we got from the cluster
		_, err = fd.WriteAt(buf, state.block.Offset)
		if err != nil {
			state.earlyClose("save", err)
			continue nextBlock
		}

		state.pullDone()
		out <- state.sharedPullerState
	}
}

func (p *Puller) finisherRoutine(in <-chan *sharedPullerState) {
	for state := range in {
		if closed, err := state.finalClose(); closed {
			if debug {
				l.Debugln(p, "closing", state.file.Name)
			}
			if err != nil {
				l.Warnln("puller: final:", err)
				continue
			}

			// Verify the file against expected hashes
			fd, err := os.Open(state.tempName)
			if err != nil {
				os.Remove(state.tempName)
				l.Warnln("puller: final:", err)
				continue
			}
			err = scanner.Verify(fd, protocol.BlockSize, state.file.Blocks)
			fd.Close()
			if err != nil {
				os.Remove(state.tempName)
				l.Warnln("puller: final:", state.file.Name, err)
				continue
			}

			// Set the correct permission bits on the new file
			err = os.Chmod(state.tempName, os.FileMode(state.file.Flags&0777))
			if err != nil {
				os.Remove(state.tempName)
				l.Warnln("puller: final:", err)
				continue
			}

			// Set the correct timestamp on the new file
			t := time.Unix(state.file.Modified, 0)
			err = os.Chtimes(state.tempName, t, t)
			if err != nil {
				os.Remove(state.tempName)
				l.Warnln("puller: final:", err)
				continue
			}

			// If we should use versioning, let the versioner archive the old
			// file before we replace it. Archiving a non-existent file is not
			// an error.
			if p.versioner != nil {
				err = p.versioner.Archive(state.realName)
				if err != nil {
					os.Remove(state.tempName)
					l.Warnln("puller: final:", err)
					continue
				}
			}

			// Replace the original file with the new one
			err = osutil.Rename(state.tempName, state.realName)
			if err != nil {
				os.Remove(state.tempName)
				l.Warnln("puller: final:", err)
				continue
			}

			// Record the updated file in the index
			p.model.updateLocal(p.folder, state.file)
		}
	}
}

// clean deletes orphaned temporary files
func (p *Puller) clean() {
	keep := time.Duration(p.model.cfg.Options.KeepTemporariesH) * time.Hour
	now := time.Now()
	filepath.Walk(p.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode().IsRegular() && defTempNamer.IsTemporary(path) && info.ModTime().Add(keep).Before(now) {
			os.Remove(path)
		}

		return nil
	})
}

func invalidateFolder(cfg *config.Configuration, folderID string, err error) {
	for i := range cfg.Folders {
		folder := &cfg.Folders[i]
		if folder.ID == folderID {
			folder.Invalid = err.Error()
			return
		}
	}
}
