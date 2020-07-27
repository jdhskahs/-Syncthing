// Copyright (C) 2020 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRealCase(t *testing.T) {
	// Verify realCase lookups on various underlying filesystems.

	t.Run("fake-sensitive", func(t *testing.T) {
		testRealCase(t, newFakeFilesystem(t.Name()))
	})
	t.Run("fake-insensitive", func(t *testing.T) {
		testRealCase(t, newFakeFilesystem(t.Name()+"?insens=true"))
	})
	t.Run("actual", func(t *testing.T) {
		fsys, tmpDir := setup(t)
		defer os.RemoveAll(tmpDir)
		testRealCase(t, fsys)
	})
}

func testRealCase(t *testing.T, fsys Filesystem) {
	testFs := NewCaseFilesystem(fsys).(*caseFilesystem)
	comps := []string{"Foo", "bar", "BAZ", "bAs"}
	path := filepath.Join(comps...)
	testFs.MkdirAll(filepath.Join(comps[:len(comps)-1]...), 0777)
	fd, err := testFs.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	fd.Close()

	for i, tc := range []struct {
		in  string
		len int
	}{
		{path, 4},
		{strings.ToLower(path), 4},
		{strings.ToUpper(path), 4},
		{"foo", 1},
		{"FOO", 1},
		{"foO", 1},
		{filepath.Join("Foo", "bar"), 2},
		{filepath.Join("Foo", "bAr"), 2},
		{filepath.Join("FoO", "bar"), 2},
		{filepath.Join("foo", "bar", "BAZ"), 3},
		{filepath.Join("Foo", "bar", "bAz"), 3},
		{filepath.Join("foo", "bar", "BAZ"), 3}, // Repeat on purpose
	} {
		out, err := testFs.realCase(tc.in)
		if err != nil {
			t.Error(err)
		} else if exp := filepath.Join(comps[:tc.len]...); out != exp {
			t.Errorf("tc %v: Expected %v, got %v", i, exp, out)
		}
	}
}

func TestRealCaseSensitive(t *testing.T) {
	// Verify that realCase returns the best on-disk case for case sensitive
	// systems. Test is skipped if the underlying fs is insensitive.

	t.Run("fake-sensitive", func(t *testing.T) {
		testRealCaseSensitive(t, newFakeFilesystem(t.Name()))
	})
	t.Run("actual", func(t *testing.T) {
		fsys, tmpDir := setup(t)
		defer os.RemoveAll(tmpDir)
		testRealCaseSensitive(t, fsys)
	})
}

func testRealCaseSensitive(t *testing.T, fsys Filesystem) {
	testFs := NewCaseFilesystem(fsys).(*caseFilesystem)

	names := make([]string, 2)
	names[0] = "foo"
	names[1] = strings.ToUpper(names[0])
	for _, n := range names {
		if err := testFs.MkdirAll(n, 0777); err != nil {
			if IsErrCaseConflict(err) {
				t.Skip("Filesystem is case-insensitive")
			}
			t.Fatal(err)
		}
	}

	for _, n := range names {
		if rn, err := testFs.realCase(n); err != nil {
			t.Error(err)
		} else if rn != n {
			t.Errorf("Got %v, expected %v", rn, n)
		}
	}
}

func TestCaseFSStat(t *testing.T) {
	// Verify that a Stat() lookup behaves in a case sensitive manner
	// regardless of the underlying fs.

	t.Run("fake-sensitive", func(t *testing.T) {
		testCaseFSStat(t, newFakeFilesystem(t.Name()))
	})
	t.Run("fake-insensitive", func(t *testing.T) {
		testCaseFSStat(t, newFakeFilesystem(t.Name()+"?insens=true"))
	})
	t.Run("actual", func(t *testing.T) {
		fsys, tmpDir := setup(t)
		defer os.RemoveAll(tmpDir)
		testCaseFSStat(t, fsys)
	})
}

func testCaseFSStat(t *testing.T, fsys Filesystem) {
	fd, err := fsys.Create("foo")
	if err != nil {
		t.Fatal(err)
	}
	fd.Close()

	// Check if the underlying fs is sensitive or not
	sensitive := true
	if _, err = fsys.Stat("FOO"); err == nil {
		sensitive = false
	}

	testFs := NewCaseFilesystem(fsys)
	_, err = testFs.Stat("FOO")
	if sensitive {
		if IsNotExist(err) {
			t.Log("pass: case sensitive underlying fs")
		} else {
			t.Error("expected NotExist, not", err, "for sensitive fs")
		}
	} else if IsErrCaseConflict(err) {
		t.Log("pass: case insensitive underlying fs")
	} else {
		t.Error("expected ErrCaseConflict, not", err, "for insensitive fs")
	}
}

func BenchmarkWalkCaseFakeFS10k(b *testing.B) {
	nfiles := 10_000
	fsys, paths := fakefsForBenchmark(b, nfiles)
	b.Run("raw", func(b *testing.B) {
		benchmarkWalkFakeFS(b, fsys, paths)
		b.ReportAllocs()
	})
	b.Run("case", func(b *testing.B) {
		benchmarkWalkFakeFS(b, NewCaseFilesystem(fsys), paths)
		b.ReportAllocs()
	})
}

func fakefsForBenchmark(b *testing.B, nfiles int) (Filesystem, []string) {
	fsys := NewFilesystem(FilesystemTypeFake, fmt.Sprintf("fakefsForBenchmark?files=%d&insens=true", nfiles))

	var paths []string
	if err := fsys.Walk("/", func(path string, info FileInfo, err error) error {
		paths = append(paths, path)
		return err
	}); err != nil {
		b.Fatal(err)
	}
	if len(paths) < b.N {
		b.Fatal("didn't find enough stuff")
	}

	return fsys, paths
}

func benchmarkWalkFakeFS(b *testing.B, fsys Filesystem, paths []string) {
	// Simulate a scanner pass over the filesystem. First walk it to
	// discover all names, then stat each name individually to check if it's
	// been deleted or not (pretending that they all existed in the
	// database).

	var ms0 runtime.MemStats
	runtime.ReadMemStats(&ms0)
	t0 := time.Now()

	for i := 0; i < b.N; i++ {
		if err := fsys.Walk("/", func(path string, info FileInfo, err error) error {
			return err
		}); err != nil {
			b.Fatal(err)
		}

		for _, p := range paths {
			if _, err := fsys.Lstat(p); err != nil {
				b.Fatal(err)
			}
		}
	}

	t1 := time.Now()
	var ms1 runtime.MemStats
	runtime.ReadMemStats(&ms1)

	// We add metrics per path entry
	b.ReportMetric(float64(t1.Sub(t0))/float64(b.N)/float64(len(paths)), "ns/entry")
	b.ReportMetric(float64(ms1.Alloc-ms0.Alloc)/float64(b.N)/float64(len(paths)), "allocs/entry")
	b.ReportMetric(float64(ms1.TotalAlloc-ms0.TotalAlloc)/float64(b.N)/float64(len(paths)), "B/entry")
}
