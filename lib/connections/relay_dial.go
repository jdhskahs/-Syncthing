// Copyright (C) 2016 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package connections

import (
	"crypto/tls"
	"errors"
	"net"
	"net/url"
	"time"

	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/dialer"
	"github.com/syncthing/syncthing/lib/protocol"
	"github.com/syncthing/syncthing/lib/relay/client"
)

const relayPriority = 200

func init() {
	dialers["relay"] = relayDialerFactory{}
}

type relayDialer struct {
	cfg     *config.Wrapper
	tlsCfg  *tls.Config
	enabled bool
}

var ErrDialerDisabled = errors.New("disabled by configuration")

func (d *relayDialer) Dial(id protocol.DeviceID, uri *url.URL) (IntermediateConnection, error) {
	if !d.enabled {
		return IntermediateConnection{}, ErrDialerDisabled
	}

	inv, err := client.GetInvitationFromRelay(uri, id, d.tlsCfg.Certificates, 10*time.Second)
	if err != nil {
		return IntermediateConnection{}, err
	}

	conn, err := client.JoinSession(inv)
	if err != nil {
		return IntermediateConnection{}, err
	}

	err = dialer.SetTCPOptions(conn.(*net.TCPConn))
	if err != nil {
		conn.Close()
		return IntermediateConnection{}, err
	}

	var tc *tls.Conn
	if inv.ServerSocket {
		tc = tls.Server(conn, d.tlsCfg)
	} else {
		tc = tls.Client(conn, d.tlsCfg)
	}

	err = tc.Handshake()
	if err != nil {
		tc.Close()
		return IntermediateConnection{}, err
	}

	return IntermediateConnection{tc, "Relay (Client)", relayPriority}, nil
}

func (relayDialer) Priority() int {
	return relayPriority
}

func (d *relayDialer) RedialFrequency() time.Duration {
	return time.Duration(d.cfg.Options().RelayReconnectIntervalM) * time.Minute
}

func (d *relayDialer) String() string {
	return "Relay Dialer"
}

type relayDialerFactory struct{}

func (relayDialerFactory) New(cfg *config.Wrapper, tlsCfg *tls.Config) genericDialer {
	// Dialers are very short lived, so we can just grab and remember
	// cfg.Options().RelaysEnabled below and not worry about it changing.
	return &relayDialer{
		cfg:     cfg,
		tlsCfg:  tlsCfg,
		enabled: cfg.Options().RelaysEnabled,
	}
}

func (relayDialerFactory) Priority() int {
	return relayPriority
}
