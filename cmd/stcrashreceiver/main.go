// Copyright (C) 2019 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

// Command stcrashreceiver is a trivial HTTP server that allows two things:
//
// - uploading files (crash reports) named like a SHA256 hash using a PUT request
// - checking whether such file exists using a HEAD request
//
// Typically this should be deployed behind something that manages HTTPS.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/syncthing/syncthing/lib/failhandler"
	"github.com/syncthing/syncthing/lib/sha256"

	raven "github.com/getsentry/raven-go"
)

func main() {
	dir := flag.String("dir", ".", "Directory to store reports in")
	dsn := flag.String("dsn", "", "Sentry DSN")
	listen := flag.String("listen", ":22039", "HTTP listen address")
	flag.Parse()

	mux := http.NewServeMux()

	cr := &crashReceiver{
		dir: *dir,
		dsn: *dsn,
	}
	mux.Handle("/", cr)

	if *dsn != "" {
		mux.HandleFunc("/failure", handleFailureFn(*dsn))
	}

	log.SetOutput(os.Stdout)
	if err := http.ListenAndServe(*listen, mux); err != nil {
		log.Fatalln("HTTP serve:", err)
	}
}

func handleFailureFn(dsn string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		bs, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var reports []failhandler.Report
		err = json.Unmarshal(bs, &reports)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if len(reports) == 0 {
			// Shouldn't happen
			return
		}

		version, err := parseVersion(reports[0].Version)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		for _, r := range reports {
			pkt := packet(version)
			pkt.Message = r.Descr
			pkt.Extra = raven.Extra{
				"count": r.Count,
			}

			if err := sendReport(dsn, pkt, userIDFor(req)); err != nil {
				log.Println("Failed to send  crash report:", err)
			}
		}
	}
}

// userIDFor returns a string we can use as the user ID for the purpose of
// counting affected users. It's the truncated hash of a salt, the user
// remote IP, and the current month.
func userIDFor(req *http.Request) string {
	addr := req.RemoteAddr
	if fwd := req.Header.Get("x-forwarded-for"); fwd != "" {
		addr = fwd
	}
	if host, _, err := net.SplitHostPort(addr); err == nil {
		addr = host
	}
	now := time.Now().Format("200601")
	salt := "stcrashreporter"
	hash := sha256.Sum256([]byte(salt + addr + now))
	return fmt.Sprintf("%x", hash[:8])
}
