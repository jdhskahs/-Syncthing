// Copyright (C) 2015 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package discover

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/syncthing/syncthing/lib/protocol"
	"github.com/syncthing/syncthing/lib/tlsutil"
)

func TestParseOptions(t *testing.T) {
	testcases := []struct {
		in   string
		out  string
		opts serverOptions
	}{
		{"https://example.com/", "https://example.com/", serverOptions{prio: defaultGlobalPrio}},
		{"https://example.com/?insecure", "https://example.com/", serverOptions{insecure: true, prio: defaultGlobalPrio}},
		{"https://example.com/?insecure=true", "https://example.com/", serverOptions{insecure: true, prio: defaultGlobalPrio}},
		{"https://example.com/?insecure=yes", "https://example.com/", serverOptions{insecure: true, prio: defaultGlobalPrio}},
		{"https://example.com/?insecure=false&noannounce", "https://example.com/", serverOptions{noAnnounce: true, prio: defaultGlobalPrio}},
		{"https://example.com/?id=abc", "https://example.com/", serverOptions{id: "abc", insecure: true, prio: defaultGlobalPrio}},
		{"https://example.com/?id=abc&prio=47", "https://example.com/", serverOptions{id: "abc", insecure: true, prio: 47}},
	}

	for _, tc := range testcases {
		res, opts, err := parseOptions(tc.in)
		if err != nil {
			t.Errorf("Unexpected err %v for %v", err, tc.in)
			continue
		}
		if res != tc.out {
			t.Errorf("Incorrect server, %v!= %v for %v", res, tc.out, tc.in)
		}
		if opts != tc.opts {
			t.Errorf("Incorrect options, %v!= %v for %v", opts, tc.opts, tc.in)
		}
	}
}

func TestGlobalOverHTTP(t *testing.T) {
	// HTTP works for queries, but is obviously insecure and we can't do
	// announces over it (as we don't present a certificate). As such, http://
	// is only allowed in combination with the "insecure" and "noannounce"
	// parameters.

	if _, err := NewGlobal("http://192.0.2.42/", tls.Certificate{}, nil, nil); err == nil {
		t.Fatal("http is not allowed without insecure and noannounce")
	}

	if _, err := NewGlobal("http://192.0.2.42/?insecure", tls.Certificate{}, nil, nil); err == nil {
		t.Fatal("http is not allowed without noannounce")
	}

	if _, err := NewGlobal("http://192.0.2.42/?noannounce", tls.Certificate{}, nil, nil); err == nil {
		t.Fatal("http is not allowed without insecure")
	}

	// Now lets check that lookups work over HTTP, given the correct options.

	list, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer list.Close()

	s := new(fakeDiscoveryServer)
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handler)
	go http.Serve(list, mux)

	direct, relays, err := testLookup("http://" + list.Addr().String() + "?insecure&noannounce")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(direct) != 1 || direct[0] != "tcp://192.0.2.42:22000?prio=100" {
		t.Errorf("incorrect direct list: %+v", direct)
	}
	if len(relays) != 1 || relays[0] != (Relay{URL: "relay://192.0.2.43:443", Latency: 42}) {
		t.Errorf("incorrect relays list: %+v", direct)
	}
}

func TestGlobalOverHTTPS(t *testing.T) {
	dir, err := ioutil.TempDir("", "syncthing")
	if err != nil {
		t.Fatal(err)
	}

	// Generate a server certificate, using fewer bits than usual to hurry the
	// process along a bit.
	cert, err := tlsutil.NewCertificate(dir+"/cert.pem", dir+"/key.pem", "syncthing", 1024)
	if err != nil {
		t.Fatal(err)
	}

	list, err := tls.Listen("tcp4", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		t.Fatal(err)
	}
	defer list.Close()

	s := new(fakeDiscoveryServer)
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handler)
	go http.Serve(list, mux)

	// With default options the lookup code expects the server certificate to
	// check out according to the usual CA chains etc. That won't be the case
	// here so we expect the lookup to fail.

	url := "https://" + list.Addr().String()
	if _, _, err := testLookup(url); err == nil {
		t.Fatalf("unexpected nil error when we should have got a certificate error")
	}

	// With "insecure" set, whatever certificate is on the other side should
	// be accepted.

	url = "https://" + list.Addr().String() + "?insecure"
	if direct, relays, err := testLookup(url); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if len(direct) != 1 || direct[0] != "tcp://192.0.2.42:22000?prio=100" {
			t.Errorf("incorrect direct list: %+v", direct)
		}
		if len(relays) != 1 || relays[0] != (Relay{URL: "relay://192.0.2.43:443", Latency: 42}) {
			t.Errorf("incorrect relays list: %+v", direct)
		}
	}

	// With "id" set to something incorrect, the checks should fail again.

	url = "https://" + list.Addr().String() + "?id=" + protocol.LocalDeviceID.String()
	if _, _, err := testLookup(url); err == nil {
		t.Fatalf("unexpected nil error for incorrect discovery server ID")
	}

	// With the correct device ID, the check should pass and we should get a
	// lookup response.

	id := protocol.NewDeviceID(cert.Certificate[0])
	url = "https://" + list.Addr().String() + "?id=" + id.String()
	if direct, relays, err := testLookup(url); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if len(direct) != 1 || direct[0] != "tcp://192.0.2.42:22000?prio=100" {
			t.Errorf("incorrect direct list: %+v", direct)
		}
		if len(relays) != 1 || relays[0] != (Relay{URL: "relay://192.0.2.43:443", Latency: 42}) {
			t.Errorf("incorrect relays list: %+v", direct)
		}
	}
}

func TestGlobalAnnounce(t *testing.T) {
	dir, err := ioutil.TempDir("", "syncthing")
	if err != nil {
		t.Fatal(err)
	}

	// Generate a server certificate, using fewer bits than usual to hurry the
	// process along a bit.
	cert, err := tlsutil.NewCertificate(dir+"/cert.pem", dir+"/key.pem", "syncthing", 1024)
	if err != nil {
		t.Fatal(err)
	}

	list, err := tls.Listen("tcp4", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		t.Fatal(err)
	}
	defer list.Close()

	s := new(fakeDiscoveryServer)
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handler)
	go http.Serve(list, mux)

	url := "https://" + list.Addr().String() + "?insecure"
	disco, err := NewGlobal(url, cert, new(fakeAddressLister), new(fakeRelayStatus))
	if err != nil {
		t.Fatal(err)
	}

	go disco.Serve()
	defer disco.Stop()

	// The discovery thing should attempt an announcement immediately. We wait
	// for it to succeed, a while.
	t0 := time.Now()
	for err := disco.Error(); err != nil; err = disco.Error() {
		if time.Since(t0) > 10*time.Second {
			t.Fatal("announce failed:", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !strings.Contains(string(s.announce), "tcp://0.0.0.0:22000") {
		t.Errorf("announce missing direct address: %s", s.announce)
	}
	if !strings.Contains(string(s.announce), "relay://192.0.2.42:443") {
		t.Errorf("announce missing relay address: %s", s.announce)
	}
}

func testLookup(url string) ([]string, []Relay, error) {
	disco, err := NewGlobal(url, tls.Certificate{}, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	go disco.Serve()
	defer disco.Stop()

	return disco.Lookup(protocol.LocalDeviceID)
}

type fakeDiscoveryServer struct {
	announce []byte
}

func (s *fakeDiscoveryServer) handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		s.announce, _ = ioutil.ReadAll(r.Body)
		w.WriteHeader(204)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"direct":["tcp://192.0.2.42:22000"], "relays":[{"url": "relay://192.0.2.43:443", "latency": 42}]}`))
	}
}

type fakeAddressLister struct{}

func (f *fakeAddressLister) ExternalAddresses() []string {
	return []string{"tcp://0.0.0.0:22000"}
}
func (f *fakeAddressLister) AllAddresses() []string {
	return []string{"tcp://0.0.0.0:22000", "tcp://192.168.0.1:22000"}
}

type fakeRelayStatus struct{}

func (f *fakeRelayStatus) Relays() []string {
	return []string{"relay://192.0.2.42:443"}
}
func (f *fakeRelayStatus) RelayStatus(uri string) (time.Duration, bool) {
	return 42 * time.Millisecond, true
}
