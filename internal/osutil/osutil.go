// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

// Package osutil implements utilities for native OS support.
package osutil

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/syncthing/syncthing/internal/sync"
)

var ErrNoHome = errors.New("No home directory found - set $HOME (or the platform equivalent).")

// Try to keep this entire operation atomic-like. We shouldn't be doing this
// often enough that there is any contention on this lock.
var renameLock sync.Mutex = sync.NewMutex()

// TryRename renames a file, leaving source file intact in case of failure.
// Tries hard to succeed on various systems by temporarily tweaking directory
// permissions and removing the destination file when necessary.
func TryRename(from, to string) error {
	renameLock.Lock()
	defer renameLock.Unlock()

	return withPreparedTarget(to, func() error {
		return os.Rename(from, to)
	})
}

// Rename moves a temporary file to it's final place.
// Will make sure to delete the from file if the operation fails, so use only
// for situations like committing a temp file to it's final location.
// Tries hard to succeed on various systems by temporarily tweaking directory
// permissions and removing the destination file when necessary.
func Rename(from, to string) error {
	// Don't leave a dangling temp file in case of rename error
	defer os.Remove(from)
	return TryRename(from, to)
}

// Copy copies the file content from source to destination.
// Tries hard to succeed on various systems by temporarily tweaking directory
// permissions and removing the destination file when necessary.
func Copy(from, to string) (err error) {
	return withPreparedTarget(to, func() error {
		return copyFileContents(from, to)
	})
}

// InWritableDir calls fn(path), while making sure that the directory
// containing `path` is writable for the duration of the call.
func InWritableDir(fn func(string) error, path string) error {
	dir := filepath.Dir(path)
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("Not a directory: " + path)
	}
	if info.Mode()&0200 == 0 {
		// A non-writeable directory (for this user; we assume that's the
		// relevant part). Temporarily change the mode so we can delete the
		// file or directory inside it.
		err = os.Chmod(dir, 0755)
		if err == nil {
			defer func() {
				err = os.Chmod(dir, info.Mode())
				if err != nil {
					// We managed to change the permission bits like a
					// millisecond ago, so it'd be bizarre if we couldn't
					// change it back.
					panic(err)
				}
			}()
		}
	}

	return fn(path)
}

// On Windows, removes the read-only attribute from the target prior deletion.
func Remove(path string) error {
	if runtime.GOOS == "windows" {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.Mode()&0200 == 0 {
			os.Chmod(path, 0700)
		}
	}
	return os.Remove(path)
}

func ExpandTilde(path string) (string, error) {
	if path == "~" {
		return getHomeDir()
	}

	path = filepath.FromSlash(path)
	if !strings.HasPrefix(path, fmt.Sprintf("~%c", os.PathSeparator)) {
		return path, nil
	}

	home, err := getHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, path[2:]), nil
}

func getHomeDir() (string, error) {
	var home string

	switch runtime.GOOS {
	case "windows":
		home = filepath.Join(os.Getenv("HomeDrive"), os.Getenv("HomePath"))
		if home == "" {
			home = os.Getenv("UserProfile")
		}
	default:
		home = os.Getenv("HOME")
	}

	if home == "" {
		return "", ErrNoHome
	}

	return home, nil
}

// Tries hard to succeed on various systems by temporarily tweaking directory
// permissions and removing the destination file when necessary.
func withPreparedTarget(to string, f func() error) error {
	// Make sure the destination directory is writeable
	toDir := filepath.Dir(to)
	if info, err := os.Stat(toDir); err == nil && info.IsDir() && info.Mode()&0200 == 0 {
		os.Chmod(toDir, 0755)
		defer os.Chmod(toDir, info.Mode())
	}

	// On Windows, make sure the destination file is writeable (or we can't delete it)
	if runtime.GOOS == "windows" {
		os.Chmod(to, 0666)
		err := os.Remove(to)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return f()
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
