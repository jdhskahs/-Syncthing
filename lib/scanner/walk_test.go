// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package scanner

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	rdebug "runtime/debug"
	"sort"
	"sync"
	"testing"

	"github.com/d4l3k/messagediff"
	"github.com/syncthing/syncthing/lib/fs"
	"github.com/syncthing/syncthing/lib/ignore"
	"github.com/syncthing/syncthing/lib/osutil"
	"github.com/syncthing/syncthing/lib/protocol"
	"golang.org/x/text/unicode/norm"
)

type testfile struct {
	name   string
	length int64
	hash   string
}

type testfileList []testfile

var testdata = testfileList{
	{"afile", 4, "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c"},
	{"dir1", 128, ""},
	{filepath.Join("dir1", "dfile"), 5, "49ae93732fcf8d63fe1cce759664982dbd5b23161f007dba8561862adc96d063"},
	{"dir2", 128, ""},
	{filepath.Join("dir2", "cfile"), 4, "bf07a7fbb825fc0aae7bf4a1177b2b31fcf8a3feeaf7092761e18c859ee52a9c"},
	{"excludes", 37, "df90b52f0c55dba7a7a940affe482571563b1ac57bd5be4d8a0291e7de928e06"},
	{"further-excludes", 5, "7eb0a548094fa6295f7fd9200d69973e5f5ec5c04f2a86d998080ac43ecf89f1"},
}

func init() {
	// This test runs the risk of entering infinite recursion if it fails.
	// Limit the stack size to 10 megs to crash early in that case instead of
	// potentially taking down the box...
	rdebug.SetMaxStack(10 * 1 << 20)
}

func TestWalkSub(t *testing.T) {
	ignores := ignore.New(false)
	err := ignores.Load("testdata/.stignore")
	if err != nil {
		t.Fatal(err)
	}

	fchan, err := Walk(context.TODO(), Config{
		Dir:       "testdata",
		Subs:      []string{"dir2"},
		BlockSize: 128 * 1024,
		Matcher:   ignores,
		Hashers:   2,
	})
	var files []protocol.FileInfo
	for f := range fchan {
		files = append(files, f)
	}
	if err != nil {
		t.Fatal(err)
	}

	// The directory contains two files, where one is ignored from a higher
	// level. We should see only the directory and one of the files.

	if len(files) != 2 {
		t.Fatalf("Incorrect length %d != 2", len(files))
	}
	if files[0].Name != "dir2" {
		t.Errorf("Incorrect file %v != dir2", files[0])
	}
	if files[1].Name != filepath.Join("dir2", "cfile") {
		t.Errorf("Incorrect file %v != dir2/cfile", files[1])
	}
}

func TestWalk(t *testing.T) {
	ignores := ignore.New(false)
	err := ignores.Load("testdata/.stignore")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ignores)

	fchan, err := Walk(context.TODO(), Config{
		Dir:       "testdata",
		BlockSize: 128 * 1024,
		Matcher:   ignores,
		Hashers:   2,
	})

	if err != nil {
		t.Fatal(err)
	}

	var tmp []protocol.FileInfo
	for f := range fchan {
		tmp = append(tmp, f)
	}
	sort.Sort(fileList(tmp))
	files := fileList(tmp).testfiles()

	if diff, equal := messagediff.PrettyDiff(testdata, files); !equal {
		t.Errorf("Walk returned unexpected data. Diff:\n%s", diff)
	}
}

func TestWalkError(t *testing.T) {
	_, err := Walk(context.TODO(), Config{
		Dir:       "testdata-missing",
		BlockSize: 128 * 1024,
		Hashers:   2,
	})

	if err == nil {
		t.Error("no error from missing directory")
	}

	_, err = Walk(context.TODO(), Config{
		Dir:       "testdata/bar",
		BlockSize: 128 * 1024,
	})

	if err == nil {
		t.Error("no error from non-directory")
	}
}

func TestVerify(t *testing.T) {
	blocksize := 16
	// data should be an even multiple of blocksize long
	data := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut e")
	buf := bytes.NewBuffer(data)
	progress := newByteCounter()
	defer progress.Close()

	blocks, err := Blocks(context.TODO(), buf, blocksize, -1, progress, false)
	if err != nil {
		t.Fatal(err)
	}
	if exp := len(data) / blocksize; len(blocks) != exp {
		t.Fatalf("Incorrect number of blocks %d != %d", len(blocks), exp)
	}

	if int64(len(data)) != progress.Total() {
		t.Fatalf("Incorrect counter value %d  != %d", len(data), progress.Total())
	}

	buf = bytes.NewBuffer(data)
	err = Verify(buf, blocksize, blocks)
	t.Log(err)
	if err != nil {
		t.Fatal("Unexpected verify failure", err)
	}

	buf = bytes.NewBuffer(append(data, '\n'))
	err = Verify(buf, blocksize, blocks)
	t.Log(err)
	if err == nil {
		t.Fatal("Unexpected verify success")
	}

	buf = bytes.NewBuffer(data[:len(data)-1])
	err = Verify(buf, blocksize, blocks)
	t.Log(err)
	if err == nil {
		t.Fatal("Unexpected verify success")
	}

	data[42] = 42
	buf = bytes.NewBuffer(data)
	err = Verify(buf, blocksize, blocks)
	t.Log(err)
	if err == nil {
		t.Fatal("Unexpected verify success")
	}
}

func TestNormalization(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("Normalization test not possible on darwin")
		return
	}

	os.RemoveAll("testdata/normalization")
	defer os.RemoveAll("testdata/normalization")

	tests := []string{
		"0-A",            // ASCII A -- accepted
		"1-\xC3\x84",     // NFC 'Ä' -- conflicts with the entry below, accepted
		"1-\x41\xCC\x88", // NFD 'Ä' -- conflicts with the entry above, ignored
		"2-\xC3\x85",     // NFC 'Å' -- accepted
		"3-\x41\xCC\x83", // NFD 'Ã' -- converted to NFC
		"4-\xE2\x98\x95", // U+2615 HOT BEVERAGE (☕) -- accepted
		"5-\xCD\xE2",     // EUC-CN "wài" (外) -- ignored (not UTF8)
	}
	numInvalid := 2

	if runtime.GOOS == "windows" {
		// On Windows, in case 5 the character gets replaced with a
		// replacement character \xEF\xBF\xBD at the point it's written to disk,
		// which means it suddenly becomes valid (sort of).
		numInvalid--
	}

	numValid := len(tests) - numInvalid

	for _, s1 := range tests {
		// Create a directory for each of the interesting strings above
		if err := osutil.MkdirAll(filepath.Join("testdata/normalization", s1), 0755); err != nil {
			t.Fatal(err)
		}

		for _, s2 := range tests {
			// Within each dir, create a file with each of the interesting
			// file names. Ensure that the file doesn't exist when it's
			// created. This detects and fails if there's file name
			// normalization stuff at the filesystem level.
			if fd, err := os.OpenFile(filepath.Join("testdata/normalization", s1, s2), os.O_CREATE|os.O_EXCL, 0644); err != nil {
				t.Fatal(err)
			} else {
				fd.WriteString("test")
				fd.Close()
			}
		}
	}

	// We can normalize a directory name, but we can't descend into it in the
	// same pass due to how filepath.Walk works. So we run the scan twice to
	// make sure it all gets done. In production, things will be correct
	// eventually...

	_, err := walkDir("testdata/normalization")
	if err != nil {
		t.Fatal(err)
	}
	tmp, err := walkDir("testdata/normalization")
	if err != nil {
		t.Fatal(err)
	}

	files := fileList(tmp).testfiles()

	// We should have one file per combination, plus the directories
	// themselves

	expectedNum := numValid*numValid + numValid
	if len(files) != expectedNum {
		t.Errorf("Expected %d files, got %d", expectedNum, len(files))
	}

	// The file names should all be in NFC form.

	for _, f := range files {
		t.Logf("%q (% x) %v", f.name, f.name, norm.NFC.IsNormalString(f.name))
		if !norm.NFC.IsNormalString(f.name) {
			t.Errorf("File name %q is not NFC normalized", f.name)
		}
	}
}

func TestIssue1507(t *testing.T) {
	w := &walker{}
	c := make(chan protocol.FileInfo, 100)
	fn := w.walkAndHashFiles(context.TODO(), c, c)

	fn("", nil, protocol.ErrClosed)
}

func TestWalkSymlinkUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping unsupported symlink test")
		return
	}

	// Create a folder with a symlink in it

	os.RemoveAll("_symlinks")
	defer os.RemoveAll("_symlinks")

	os.Mkdir("_symlinks", 0755)
	os.Symlink("destination", "_symlinks/link")

	// Scan it

	fchan, err := Walk(context.TODO(), Config{
		Dir:       "_symlinks",
		BlockSize: 128 * 1024,
	})

	if err != nil {
		t.Fatal(err)
	}

	var files []protocol.FileInfo
	for f := range fchan {
		files = append(files, f)
	}

	// Verify that we got one symlink and with the correct attributes

	if len(files) != 1 {
		t.Errorf("expected 1 symlink, not %d", len(files))
	}
	if len(files[0].Blocks) != 0 {
		t.Errorf("expected zero blocks for symlink, not %d", len(files[0].Blocks))
	}
	if files[0].SymlinkTarget != "destination" {
		t.Errorf("expected symlink to have target destination, not %q", files[0].SymlinkTarget)
	}
}

func TestWalkSymlinkWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("skipping unsupported symlink test")
	}

	// Create a folder with a symlink in it

	os.RemoveAll("_symlinks")
	defer os.RemoveAll("_symlinks")

	os.Mkdir("_symlinks", 0755)
	if err := os.Symlink("destination", "_symlinks/link"); err != nil {
		// Probably we require permissions we don't have.
		t.Skip(err)
	}

	// Scan it

	fchan, err := Walk(context.TODO(), Config{
		Dir:       "_symlinks",
		BlockSize: 128 * 1024,
	})

	if err != nil {
		t.Fatal(err)
	}

	var files []protocol.FileInfo
	for f := range fchan {
		files = append(files, f)
	}

	// Verify that we got zero symlinks

	if len(files) != 0 {
		t.Errorf("expected zero symlinks, not %d", len(files))
	}
}

func walkDir(dir string) ([]protocol.FileInfo, error) {
	fchan, err := Walk(context.TODO(), Config{
		Dir:           dir,
		BlockSize:     128 * 1024,
		AutoNormalize: true,
		Hashers:       2,
	})

	if err != nil {
		return nil, err
	}

	var tmp []protocol.FileInfo
	for f := range fchan {
		tmp = append(tmp, f)
	}
	sort.Sort(fileList(tmp))

	return tmp, nil
}

type fileList []protocol.FileInfo

func (l fileList) Len() int {
	return len(l)
}

func (l fileList) Less(a, b int) bool {
	return l[a].Name < l[b].Name
}

func (l fileList) Swap(a, b int) {
	l[a], l[b] = l[b], l[a]
}

func (l fileList) testfiles() testfileList {
	testfiles := make(testfileList, len(l))
	for i, f := range l {
		if len(f.Blocks) > 1 {
			panic("simple test case stuff only supports a single block per file")
		}
		testfiles[i] = testfile{name: f.Name, length: f.FileSize()}
		if len(f.Blocks) == 1 {
			testfiles[i].hash = fmt.Sprintf("%x", f.Blocks[0].Hash)
		}
	}
	return testfiles
}

func (l testfileList) String() string {
	var b bytes.Buffer
	b.WriteString("{\n")
	for _, f := range l {
		fmt.Fprintf(&b, "  %s (%d bytes): %s\n", f.name, f.length, f.hash)
	}
	b.WriteString("}")
	return b.String()
}

var initOnce sync.Once

const (
	testdataSize = 17 << 20
	testdataName = "_random.data"
)

func BenchmarkHashFile(b *testing.B) {
	initOnce.Do(initTestFile)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := HashFile(context.TODO(), fs.DefaultFilesystem, testdataName, protocol.BlockSize, nil, true); err != nil {
			b.Fatal(err)
		}
	}

	b.SetBytes(testdataSize)
	b.ReportAllocs()
}

func initTestFile() {
	fd, err := os.Create(testdataName)
	if err != nil {
		panic(err)
	}

	lr := io.LimitReader(rand.Reader, testdataSize)
	if _, err := io.Copy(fd, lr); err != nil {
		panic(err)
	}

	if err := fd.Close(); err != nil {
		panic(err)
	}
}

func TestStopWalk(t *testing.T) {
	// The "infiniteFS" is a tree that is 100 levels deep, with each level
	// containing 100 files (each 1 MB) and 100 directories (in turn
	// containing 100 files and 100 directories, etc). That is, in total
	// 100^100 files and 100^100 directories. It'll take a while to scan,
	// unless the cancel works.

	fs := fs.NewWalkFilesystem(&infiniteFS{100, 100, 1e6})

	ctx, cancel := context.WithCancel(context.Background())
	fchan, err := Walk(ctx, Config{
		Dir:        "testdir",
		BlockSize:  128 * 1024,
		Hashers:    2,
		Filesystem: fs,
	})

	if err != nil {
		t.Fatal(err)
	}

	// Receive a couple of thousand entries to make sure the walker is up
	// and running.
	for i := 0; i < 2000; i++ {
		<-fchan
	}

	// Cancel the walker.
	cancel()

	// Empty out any waiting entries and wait for the channel to close.
	// Count them, they should be zero or very few.
	extra := 0
	for range fchan {
		extra++
	}
	if extra > 4 {
		t.Error("unexpected extra entries received after cancel")
	}
}
