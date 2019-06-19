// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package versioner

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/syncthing/syncthing/lib/fs"
	"github.com/syncthing/syncthing/lib/osutil"
	"github.com/syncthing/syncthing/lib/util"
)

var errDirectory = fmt.Errorf("cannot restore on top of a directory")
var errNotFound = fmt.Errorf("version not found")
var errFileAlreadyExists = fmt.Errorf("file already exists")

// Inserts ~tag just before the extension of the filename.
func TagFilename(name, tag string) string {
	dir, file := filepath.Dir(name), filepath.Base(name)
	ext := filepath.Ext(file)
	withoutExt := file[:len(file)-len(ext)]
	return filepath.Join(dir, withoutExt+"~"+tag+ext)
}

var tagExp = regexp.MustCompile(`.*~([^~.]+)(?:\.[^.]+)?$`)

// Returns the tag from a filename, whether at the end or middle.
func ExtractTag(path string) string {
	match := tagExp.FindStringSubmatch(path)
	// match is []string{"whole match", "submatch"} when successful

	if len(match) != 2 {
		return ""
	}
	return match[1]
}

func UntagFilename(path string) (string, string) {
	ext := filepath.Ext(path)
	versionTag := ExtractTag(path)

	// Files tagged with old style tags cannot be untagged.
	if versionTag == "" {
		return "", ""
	}

	// Old style tag
	if strings.HasSuffix(ext, versionTag) {
		return strings.TrimSuffix(path, "~"+versionTag), versionTag
	}

	withoutExt := path[:len(path)-len(ext)-len(versionTag)-1]
	name := withoutExt + ext
	return name, versionTag
}

func retrieveVersions(fileSystem fs.Filesystem) (map[string][]FileVersion, error) {
	files := make(map[string][]FileVersion)

	err := fileSystem.Walk(".", func(path string, f fs.FileInfo, err error) error {
		// Skip root (which is ok to be a symlink)
		if path == "." {
			return nil
		}

		// Skip walking if we cannot walk...
		if err != nil {
			return err
		}

		// Ignore symlinks
		if f.IsSymlink() {
			return fs.SkipDir
		}

		// No records for directories
		if f.IsDir() {
			return nil
		}

		modTime := f.ModTime().Truncate(time.Second)

		path = osutil.NormalizedFilename(path)

		name, tag := UntagFilename(path)
		// Something invalid, assume it's an untagged file (trashcan versioner stuff)
		if name == "" || tag == "" {
			files[path] = append(files[path], FileVersion{
				VersionTime: modTime,
				ModTime:     modTime,
				Size:        f.Size(),
			})
			return nil
		}

		versionTime, err := time.ParseInLocation(TimeFormat, tag, time.Local)
		if err != nil {
			// Can't parse it, welp, continue
			return nil
		}

		if err == nil {
			files[name] = append(files[name], FileVersion{
				// This looks backwards, but mtime of the file is when we archived it, making that the version time
				// The mod time of the file before archiving is embedded in the file name.
				VersionTime: versionTime,
				ModTime:     modTime,
				Size:        f.Size(),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

type fileTagger func(string, string) string

func archiveFile(srcFs, dstFs fs.Filesystem, filePath string, tagger fileTagger) error {
	filePath = osutil.NativeFilename(filePath)
	info, err := srcFs.Lstat(filePath)
	if fs.IsNotExist(err) {
		l.Debugln("not archiving nonexistent file", filePath)
		return nil
	} else if err != nil {
		return err
	}
	if info.IsSymlink() {
		panic("bug: attempting to version a symlink")
	}

	_, err = dstFs.Stat(".")
	if err != nil {
		if fs.IsNotExist(err) {
			l.Debugln("creating versions dir")
			err := dstFs.Mkdir(".", 0755)
			if err != nil {
				return err
			}
			_ = dstFs.Hide(".")
		} else {
			return err
		}
	}

	file := filepath.Base(filePath)
	inFolderPath := filepath.Dir(filePath)

	err = dstFs.MkdirAll(inFolderPath, 0755)
	if err != nil && !fs.IsExist(err) {
		l.Debugln("archiving", filePath, err)
		return err
	}

	now := time.Now()

	ver := tagger(file, now.Format(TimeFormat))
	dst := filepath.Join(inFolderPath, ver)
	l.Debugln("archiving", filePath, "moving to", dst)
	err = osutil.RenameOrCopy(srcFs, dstFs, filePath, dst)

	mtime := info.ModTime()
	// If it's a trashcan versioner type thing, then it does not have version time in the name
	// so use mtime for that.
	if ver == file {
		mtime = now
	}

	_ = dstFs.Chtimes(dst, mtime, mtime)

	return err
}

func restoreFile(src, dst fs.Filesystem, filePath string, versionTime time.Time, tagger fileTagger) error {
	tag := versionTime.In(time.Local).Truncate(time.Second).Format(TimeFormat)
	taggedFilePath := tagger(filePath, tag)

	// If the something already exists where we are restoring to, archive existing file for versioning
	// remove if it's a symlink, or fail if it's a directory
	if info, err := dst.Lstat(filePath); err == nil {
		switch {
		case info.IsDir():
			return errDirectory
		case info.IsSymlink():
			// Remove existing symlinks (as we don't want to archive them)
			if err := dst.Remove(filePath); err != nil {
				return errors.Wrap(err, "removing existing symlink")
			}
		case info.IsRegular():
			if err := archiveFile(dst, src, filePath, tagger); err != nil {
				return errors.Wrap(err, "archiving existing file")
			}
		default:
			panic("bug: unknown item type")
		}
	} else if !fs.IsNotExist(err) {
		return err
	}

	filePath = osutil.NativeFilename(filePath)

	// Try and find a file that has the correct mtime
	sourceFile := ""
	sourceMtime := time.Time{}
	if info, err := src.Lstat(taggedFilePath); err == nil && info.IsRegular() {
		sourceFile = taggedFilePath
		sourceMtime = info.ModTime()
	} else if err == nil {
		l.Debugln("restore:", taggedFilePath, "not regular")
	} else {
		l.Debugln("restore:", taggedFilePath, err.Error())
	}

	// Check for untagged file
	if sourceFile == "" {
		info, err := src.Lstat(filePath)
		if err == nil && info.IsRegular() && info.ModTime().Truncate(time.Second).Equal(versionTime) {
			sourceFile = filePath
			sourceMtime = info.ModTime()
		}

	}

	if sourceFile == "" {
		return errNotFound
	}

	// Check that the target location of where we are supposed to restore does not exist.
	// This should have been taken care of by the first few lines of this function.
	if _, err := dst.Lstat(filePath); err == nil {
		return errFileAlreadyExists
	} else if !fs.IsNotExist(err) {
		return err
	}

	_ = dst.MkdirAll(filepath.Dir(filePath), 0755)
	err := osutil.RenameOrCopy(src, dst, sourceFile, filePath)
	_ = dst.Chtimes(filePath, sourceMtime, sourceMtime)
	return err
}

func fsFromParams(folderFs fs.Filesystem, params map[string]string) (versionsFs fs.Filesystem) {
	if params["fsType"] == "" && params["fsPath"] == "" {
		versionsFs = fs.NewFilesystem(folderFs.Type(), filepath.Join(folderFs.URI(), ".stversions"))

	} else if params["fsType"] == "" {
		uri := params["fsPath"]
		// We only know how to deal with relative folders for basic filesystems, as that's the only one we know
		// how to check if it's absolute or relative.
		if folderFs.Type() == fs.FilesystemTypeBasic && !filepath.IsAbs(params["fsPath"]) {
			uri = filepath.Join(folderFs.URI(), params["fsPath"])
		}
		versionsFs = fs.NewFilesystem(folderFs.Type(), uri)
	} else {
		var fsType fs.FilesystemType
		_ = fsType.UnmarshalText([]byte(params["fsType"]))
		versionsFs = fs.NewFilesystem(fsType, params["fsPath"])
	}
	l.Debugf("%s (%s) folder using %s (%s) versioner dir", folderFs.URI(), folderFs.Type(), versionsFs.URI(), versionsFs.Type())
	return
}

func findAllVersions(fs fs.Filesystem, filePath string) []string {
	inFolderPath := filepath.Dir(filePath)
	file := filepath.Base(filePath)

	// Glob according to the new file~timestamp.ext pattern.
	pattern := filepath.Join(inFolderPath, TagFilename(file, TimeGlob))
	versions, err := fs.Glob(pattern)
	if err != nil {
		l.Warnln("globbing:", err, "for", pattern)
		return nil
	}
	versions = util.UniqueTrimmedStrings(versions)
	sort.Strings(versions)

	return versions
}
