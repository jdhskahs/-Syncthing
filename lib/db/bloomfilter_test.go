// Copyright (C) 2020 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package db

import (
	"crypto/sha256"
	"math/rand"
	"testing"
)

func TestBloomfilter(t *testing.T) {
	const n = 100000

	// For p = .01, n = 100000, the optimal number of bits is 958505.84
	// is 15127.39.
	bf := newBloomfilter(n, .01)
	if bf.nbits() < 958506 {
		t.Fatalf("bloom filter with %d bits too small", bf.nbits())
	}

	t.Logf("k = %d; m/n = %d/%d = %.3f",
		bf.nhashes, bf.nbits(), n, float64(bf.nbits())/n)

	// Assume that 100k random SHA-256 values are all distinct.
	r := rand.New(rand.NewSource(0xb1007))
	hashes := make([]byte, n*sha256.Size)
	r.Read(hashes)

	for i := 0; i < n; i++ {
		bf.Add(hashes[sha256.Size*i : sha256.Size*(i+1)])
	}

	for i := 0; i < n; i++ {
		hash := hashes[sha256.Size*i : sha256.Size*(i+1)]
		if !bf.Test(hash) {
			t.Errorf("%032x added to Bloom filter but not found", hash)
		}
	}

	// Try some more values to get a sense of the false positive rate.
	// Assume these are unique and distinct from the ones we added.
	const nTest = 10000
	fp := 0
	hash := make([]byte, sha256.Size)
	for i := 0; i < nTest; i++ {
		r.Read(hash)
		if bf.Test(hash) {
			fp++
		}
	}

	fpRate := float64(fp) / nTest
	if fpRate > .02 {
		t.Errorf("false positive rate = %.2f%%, want at most .02", 100*fpRate)
	}
}

func benchmarkBloomfilterAdd(b *testing.B, n int) {
	hash := make([]byte, n*sha256.Size)

	r := rand.New(rand.NewSource(98621))
	r.Read(hash)

	b.SetBytes(int64(len(hash)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bf := newBloomfilter(n, .01)
		for len(hash) > 0 {
			bf.Add(hash[:sha256.Size])
			hash = hash[sha256.Size:]
		}
	}
}

func BenchmarkBloomfilterAdd1e5(b *testing.B) { benchmarkBloomfilterAdd(b, 1e5) }
func BenchmarkBloomfilterAdd1e6(b *testing.B) { benchmarkBloomfilterAdd(b, 1e6) }
func BenchmarkBloomfilterAdd1e7(b *testing.B) { benchmarkBloomfilterAdd(b, 1e7) }

func benchmarkBloomfilterTest(b *testing.B, n int) {
	hash := make([]byte, n*sha256.Size)
	r := rand.New(rand.NewSource(0xa58a7))
	r.Read(hash)

	bf := newBloomfilter(n, .01)

	b.SetBytes(sha256.Size)

	h := make([]byte, sha256.Size)
	fp := 0
	for i := 0; i < b.N; i++ {
		r.Read(h)
		if bf.Test(h) {
			fp++
		}
	}

	b.Logf("false positive rate = %.3f%%", 100*float64(fp)/float64(b.N))
}

func BenchmarkBloomfilterTest1e5(b *testing.B) { benchmarkBloomfilterTest(b, 1e5) }
func BenchmarkBloomfilterTest1e6(b *testing.B) { benchmarkBloomfilterTest(b, 1e6) }
func BenchmarkBloomfilterTest1e7(b *testing.B) { benchmarkBloomfilterTest(b, 1e7) }
