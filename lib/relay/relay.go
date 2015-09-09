// Copyright (C) 2015 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package relay

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/syncthing/relaysrv/client"
	"github.com/syncthing/relaysrv/protocol"
	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/discover"
	"github.com/syncthing/syncthing/lib/osutil"
	"github.com/syncthing/syncthing/lib/sync"

	"github.com/thejerf/suture"
)

func NewSvc(cfg *config.Wrapper, tlsCfg *tls.Config, conns chan<- Connection) *Svc {
	svc := &Svc{
		Supervisor: suture.New("Svc", suture.Spec{
			Log: func(log string) {
				if debug {
					l.Debugln(log)
				}
			},
			FailureBackoff:   5 * time.Minute,
			FailureDecay:     float64((10 * time.Minute) / time.Second),
			FailureThreshold: 5,
		}),
		cfg:    cfg,
		tlsCfg: tlsCfg,

		tokens:  make(map[string]suture.ServiceToken),
		clients: make(map[string]*client.ProtocolClient),
		mut:     sync.NewRWMutex(),

		invitations: make(chan protocol.SessionInvitation),
	}

	rcfg := cfg.Raw()
	svc.CommitConfiguration(rcfg, rcfg)
	cfg.Subscribe(svc)

	receiver := &invitationReceiver{
		tlsCfg:      tlsCfg,
		conns:       conns,
		invitations: svc.invitations,
		stop:        make(chan struct{}),
	}

	svc.Add(receiver)

	return svc
}

type Svc struct {
	*suture.Supervisor
	cfg    *config.Wrapper
	tlsCfg *tls.Config

	tokens      map[string]suture.ServiceToken
	clients     map[string]*client.ProtocolClient
	mut         sync.RWMutex
	invitations chan protocol.SessionInvitation
}

func (s *Svc) VerifyConfiguration(from, to config.Configuration) error {
	for _, addr := range to.Options.RelayServers {
		_, err := url.Parse(addr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Svc) CommitConfiguration(from, to config.Configuration) bool {
	existing := make(map[string]*url.URL, len(to.Options.RelayServers))

	for _, addr := range to.Options.RelayServers {
		uri, err := url.Parse(addr)
		if err != nil {
			if debug {
				l.Debugln("Failed to parse relay address", addr, err)
			}
			continue
		}
		existing[uri.String()] = uri
	}

	// Query dynamic addresses, and pick the closest relay from the ones they provide.
	for key, uri := range existing {
		if uri.Scheme != "dynamic+http" && uri.Scheme != "dynamic+https" {
			continue
		}
		delete(existing, key)

		// Trim off the `dynamic+` prefix
		uri.Scheme = uri.Scheme[8:]

		if debug {
			l.Debugln("Looking up dynamic relays from", uri)
		}

		data, err := http.Get(uri.String())
		if err != nil {
			if debug {
				l.Debugln("Failed to lookup dynamic relays", err)
			}
			continue
		}

		var ann dynamicAnnouncement
		err = json.NewDecoder(data.Body).Decode(&ann)
		data.Body.Close()
		if err != nil {
			if debug {
				l.Debugln("Failed to lookup dynamic relays", err)
			}
			continue
		}

		dynRelays := make([]discover.Relay, 0, len(ann.Relays))
		for _, relayAnn := range ann.Relays {
			ruri, err := url.Parse(relayAnn.URL)
			if err != nil {
				if debug {
					l.Debugln("Failed to parse dynamic relay address", relayAnn.URL, err)
				}
				continue
			}
			if debug {
				l.Debugln("Found", ruri, "via", uri)
			}
			dynRelays = append(dynRelays, discover.Relay{
				Address: ruri.String(),
			})
		}

		dynRelayAddrs := discover.RelayAddressesSortedByLatency(dynRelays)
		if len(dynRelayAddrs) > 0 {
			closestRelay := dynRelayAddrs[0]
			if debug {
				l.Debugln("Picking", closestRelay, "as closest dynamic relay from", uri)
			}
			ruri, _ := url.Parse(closestRelay)
			existing[closestRelay] = ruri
		} else if debug {
			l.Debugln("No dynamic relay found on", uri)
		}
	}

	s.mut.Lock()

	for key, uri := range existing {
		_, ok := s.tokens[key]
		if !ok {
			if debug {
				l.Debugln("Connecting to relay", uri)
			}
			c := client.NewProtocolClient(uri, s.tlsCfg.Certificates, s.invitations)
			s.tokens[key] = s.Add(c)
			s.clients[key] = c
		}
	}

	for key, token := range s.tokens {
		_, ok := existing[key]
		if !ok {
			err := s.Remove(token)
			delete(s.tokens, key)
			delete(s.clients, key)
			if debug {
				l.Debugln("Disconnecting from relay", key, err)
			}
		}
	}

	s.mut.Unlock()

	return true
}

func (s *Svc) ClientStatus() map[string]bool {
	if s == nil {
		// A nil client does not have a status, really. Yet we may be called
		// this way, for raisins...
		return nil
	}

	s.mut.RLock()
	status := make(map[string]bool, len(s.clients))
	for uri, client := range s.clients {
		status[uri] = client.StatusOK()
	}
	s.mut.RUnlock()
	return status
}

type invitationReceiver struct {
	invitations chan protocol.SessionInvitation
	tlsCfg      *tls.Config
	conns       chan<- Connection
	stop        chan struct{}
}

func (r *invitationReceiver) Serve() {
	for {
		select {
		case inv := <-r.invitations:
			if debug {
				l.Debugln("Received relay invitation", inv)
			}
			conn, err := client.JoinSession(inv)
			if err != nil {
				if debug {
					l.Debugf("Failed to join relay session %s: %v", inv, err)
				}
				continue
			}

			err = osutil.SetTCPOptions(conn.(*net.TCPConn))
			if err != nil {
				l.Infoln(err)
			}

			var tc *tls.Conn

			if inv.ServerSocket {
				tc = tls.Server(conn, r.tlsCfg)
			} else {
				tc = tls.Client(conn, r.tlsCfg)
			}
			err = tc.Handshake()
			if err != nil {
				l.Infof("TLS handshake (BEP/relay %s): %v", inv, err)
				tc.Close()
				continue
			}
			r.conns <- Connection{tc, ConnectionTypeRelayAccept}

		case <-r.stop:
			return
		}
	}
}

func (r *invitationReceiver) Stop() {
	close(r.stop)
}

// This is the announcement recieved from the relay server;
// {"relays": [{"url": "relay://10.20.30.40:5060"}, ...]}
type dynamicAnnouncement struct {
	Relays []struct {
		URL string
	}
}
