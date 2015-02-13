// Copyright (C) 2014 The Syncthing Authors.
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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/syncthing/protocol"
	"github.com/syncthing/syncthing/internal/db"
	"github.com/syndtr/goleveldb/leveldb"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	folder := flag.String("folder", "default", "Folder ID")
	device := flag.String("device", "", "Device ID (blank for global)")
	flag.Parse()

	ldb, err := leveldb.OpenFile(flag.Arg(0), nil)
	if err != nil {
		log.Fatal(err)
	}

	fs := db.NewFileSet(*folder, ldb)

	if *device == "" {
		log.Printf("*** Global index for folder %q", *folder)
		fs.WithGlobalTruncated(func(f protocol.FileInfo) bool {
			fmt.Println(f)
			fmt.Println("\t", fs.Availability(f.Name))
			return true
		})
	} else {
		n, err := protocol.DeviceIDFromString(*device)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("*** Have index for folder %q device %q", *folder, n)
		fs.WithHaveTruncated(n, func(f protocol.FileInfo) bool {
			fmt.Println(f)
			return true
		})
	}
}
