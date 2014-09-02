// Copyright (C) 2014 Jakob Borg and Contributors (see the CONTRIBUTORS file).
// All rights reserved. Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build darwin

package protocol

// Darwin uses NFD normalization

import "code.google.com/p/go.text/unicode/norm"

type nativeModel struct {
	next Model
}

func (m nativeModel) Index(nodeID NodeID, repo string, files []FileInfo) {
	for i := range files {
		files[i].Name = norm.NFD.String(files[i].Name)
	}
	m.next.Index(nodeID, repo, files)
}

func (m nativeModel) IndexUpdate(nodeID NodeID, repo string, files []FileInfo) {
	for i := range files {
		files[i].Name = norm.NFD.String(files[i].Name)
	}
	m.next.IndexUpdate(nodeID, repo, files)
}

func (m nativeModel) Request(nodeID NodeID, repo string, name string, offset int64, size int) ([]byte, error) {
	name = norm.NFD.String(name)
	return m.next.Request(nodeID, repo, name, offset, size)
}

func (m nativeModel) ClusterConfig(nodeID NodeID, config ClusterConfigMessage) {
	m.next.ClusterConfig(nodeID, config)
}

func (m nativeModel) Close(nodeID NodeID, err error) {
	defer l.CaptureAndRepanic()
	m.next.Close(nodeID, err)
}
