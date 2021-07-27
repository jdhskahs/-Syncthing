// Copyright (C) 2021 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/disk"
)

var osEncoderMap map[string]FilesystemEncoderType
var fstypeEncoderMap map[string]FilesystemEncoderType

func init() {
	osEncoderMap = make(map[string]FilesystemEncoderType)

	osEncoderMap["android"] = FilesystemEncoderTypeAndroid
	osEncoderMap["ios"] = FilesystemEncoderTypeIos
	osEncoderMap["plan9"] = FilesystemEncoderTypePlan9
	osEncoderMap["windows"] = FilesystemEncoderTypeWindows

	fstypeEncoderMap = make(map[string]FilesystemEncoderType)

	// See https://en.wikipedia.org/wiki/Comparison_of_file_systems#Limits
	// See https://en.wikipedia.org/wiki/Filename#Reserved_characters_and_words
	fstypeEncoderMap["hfs"] = FilesystemEncoderTypeIos // No unicode?
	// fstypeEncoderMap["HFSPLUS"] = FilesystemEncoderTypeIos // ?

	fstypeEncoderMap["cifs"] = FilesystemEncoderTypeWindows
	fstypeEncoderMap["exfat"] = FilesystemEncoderTypeWindows
	fstypeEncoderMap["fat"] = FilesystemEncoderTypeWindows // No unicode?
	fstypeEncoderMap["fat32"] = FilesystemEncoderTypeWindows
	fstypeEncoderMap["hpfs"] = FilesystemEncoderTypeWindows  // No unicode?
	fstypeEncoderMap["msdos"] = FilesystemEncoderTypeWindows // No unicode?
	fstypeEncoderMap["ntfs"] = FilesystemEncoderTypeWindows
	fstypeEncoderMap["refs"] = FilesystemEncoderTypeWindows
	fstypeEncoderMap["smb"] = FilesystemEncoderTypeWindows
	fstypeEncoderMap["vfat"] = FilesystemEncoderTypeWindows
}

const cannotDetermineFilesystem = "The filesystem for %q cannot be determined, will use the %q filesystem encoder by default"

func GetFilesystemEncoderType(name string) (FilesystemEncoderType, error) {
	encoderType, ok := osEncoderMap[runtime.GOOS]
	if !ok {
		encoderType = FilesystemEncoderTypeDefault
	}

	u, err := disk.Usage(name)
	if err != nil {
		l.Debugf(cannotDetermineFilesystem, name, encoderType)
		return encoderType, err
	}

	fsType := strings.ToLower(u.Fstype)

	if fsType == "" {
		partitions, err := disk.Partitions(false)
		if err != nil {
			l.Debugf(cannotDetermineFilesystem, name, encoderType)
			return encoderType, err
		}

		if runtime.GOOS == "windows" {
			volumeName := strings.ToLower(filepath.VolumeName(name))
			if volumeName != "" {
				for _, partition := range partitions {
					device := strings.ToLower(partition.Device)
					if strings.HasPrefix(device, volumeName) {
						fsType = partition.Fstype
						break
					}
				}
			}
		}
	}

	encType, ok := fstypeEncoderMap[fsType]
	if !ok {
		l.Debugf(cannotDetermineFilesystem, name, encoderType)
		return encoderType, nil
	}
	l.Debugf("%q is formatted as %q, and will use the %q filesystem encoder", name, fsType, encoderType)
	return encType, nil
}
