// Copyright (C) 2019 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

func init() {
	registerCopyRangeImplementation(CopyRangeTypeAllWithFallback, copyRangeAllWithFallback)
}

func copyRangeAllWithFallback(src, dst basicFile, srcOffset, dstOffset, size int64) error {
	var err error
	for _, method := range []CopyRangeType{CopyRangeTypeIoctl, CopyRangeTypeCopyFileRange, CopyRangeTypeSendFile, CopyRangeTypeStandard} {
		if err = CopyRange(method, src, dst, srcOffset, dstOffset, size); err == nil {
			return nil
		}
	}
	return err
}
