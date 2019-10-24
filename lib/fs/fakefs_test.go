// Copyright (C) 2018 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"testing"
)

func TestFakeFS(t *testing.T) {
	// Test some basic aspects of the fakefs

	fs := newFakeFilesystem("/foo/bar/baz")

	// MkdirAll
	err := fs.MkdirAll("dira/dirb", 0755)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fs.Stat("dira/dirb")
	if err != nil {
		t.Fatal(err)
	}

	// Mkdir
	err = fs.Mkdir("dira/dirb/dirc", 0755)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fs.Stat("dira/dirb/dirc")
	if err != nil {
		t.Fatal(err)
	}

	// Create
	fd, err := fs.Create("/dira/dirb/test")
	if err != nil {
		t.Fatal(err)
	}

	// Write
	_, err = fd.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	// Stat on fd
	info, err := fd.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if info.Name() != "test" {
		t.Error("wrong name:", info.Name())
	}
	if info.Size() != 5 {
		t.Error("wrong size:", info.Size())
	}

	// Stat on fs
	info, err = fs.Stat("dira/dirb/test")
	if err != nil {
		t.Fatal(err)
	}
	if info.Name() != "test" {
		t.Error("wrong name:", info.Name())
	}
	if info.Size() != 5 {
		t.Error("wrong size:", info.Size())
	}

	// Seek
	_, err = fd.Seek(1, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}

	// Read
	bs0, err := ioutil.ReadAll(fd)
	if err != nil {
		t.Fatal(err)
	}
	if len(bs0) != 4 {
		t.Error("wrong number of bytes:", len(bs0))
	}

	// Read again, same data hopefully
	_, err = fd.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}
	bs1, err := ioutil.ReadAll(fd)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bs0, bs1[1:]) {
		t.Error("wrong data")
	}

	// Create symlink
	if err := fs.CreateSymlink("foo", "dira/dirb/symlink"); err != nil {
		t.Fatal(err)
	}
	if str, err := fs.ReadSymlink("dira/dirb/symlink"); err != nil {
		t.Fatal(err)
	} else if str != "foo" {
		t.Error("Wrong symlink destination", str)
	}

	// Chown
	if err := fs.Lchown("dira", 1234, 5678); err != nil {
		t.Fatal(err)
	}
	if info, err := fs.Lstat("dira"); err != nil {
		t.Fatal(err)
	} else if info.Owner() != 1234 || info.Group() != 5678 {
		t.Error("Wrong owner/group")
	}
}

func TestFakeFSRead(t *testing.T) {
	// Test some basic aspects of the fakefs

	fs := newFakeFilesystem("/foo/bar/baz")

	// Create
	fd, _ := fs.Create("test")
	fd.Truncate(3 * 1 << randomBlockShift)

	// Read
	fd.Seek(0, io.SeekStart)
	bs0, err := ioutil.ReadAll(fd)
	if err != nil {
		t.Fatal(err)
	}
	if len(bs0) != 3*1<<randomBlockShift {
		t.Error("wrong number of bytes:", len(bs0))
	}

	// Read again, starting at an odd offset
	fd.Seek(0, io.SeekStart)
	buf0 := make([]byte, 12345)
	n, _ := fd.Read(buf0)
	if n != len(buf0) {
		t.Fatal("short read")
	}
	buf1, err := ioutil.ReadAll(fd)
	if err != nil {
		t.Fatal(err)
	}
	if len(buf1) != 3*1<<randomBlockShift-len(buf0) {
		t.Error("wrong number of bytes:", len(buf1))
	}

	bs1 := append(buf0, buf1...)
	if !bytes.Equal(bs0, bs1) {
		t.Error("data mismatch")
	}

	// Read large block with ReadAt
	bs2 := make([]byte, 3*1<<randomBlockShift)
	_, err = fd.ReadAt(bs2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bs0, bs2) {
		t.Error("data mismatch")
	}
}

type testFS struct {
	name string
	fs   Filesystem
}

type test struct {
	name string
	impl func(t *testing.T, fs Filesystem)
}

func TestFakeFSCaseSensitive(t *testing.T) {
	var tests = []test{
		{"OpenFile", testFakeFSOpenFile},
		{"RemoveAll", testFakeFSRemoveAll},
		{"Remove", testFakeFSRemove},
		{"Rename", testFakeFSRename},
		{"Mkdir", testFakeFSMkdir},
		{"SameFile", testFakeFSSameFile},
		{"DirNames", testFakeFSDirNames},
		{"FileName", testFakeFSFileName},
	}
	var filesystems = []testFS{
		{"fakefs", newFakeFilesystem("/foo")},
	}

	if runtime.GOOS == "linux" {
		testDir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatalf("could not create temporary dir for testing: %s", err)
		}

		if fd, err := os.Create(filepath.Join(testDir, ".stfolder")); err != nil {
			t.Fatalf("could not create .stfolder: %s", err)
		} else {
			fd.Close()
		}

		defer func() {
			if err := os.RemoveAll(testDir); err != nil {
				t.Fatalf("could not remove test directory: %s", err)
			}
		}()

		filesystems = append(filesystems, testFS{runtime.GOOS, newBasicFilesystem(testDir)})
	}

	runTests(t, tests, filesystems)
}

func TestFakeFSCaseInsensitive(t *testing.T) {
	var tests = []test{
		{"general", testFakeFSCaseInsensitive},
		{"MkdirAll", testFakeFSCaseInsensitiveMkdirAll},
		{"Stat", testFakeFSStatInsens},
		{"Rename", testFakeFSRenameInsensitive},
		{"Mkdir", testFakeFSMkdirInsens},
		{"DirNames", testFakeFSDirNames},
		{"OpenFile", testFakeFSOpenFileInsens},
		{"RemoveAll", testFakeFSRemoveAllInsens},
		{"Remove", testFakeFSRemoveInsens},
		{"SameFile", testFakeFSSameFileInsens},
		{"Create", testFakeFSCreateInsens},
		{"FileName", testFakeFSFileNameInsens},
	}

	var filesystems = []testFS{
		{"fakefs", newFakeFilesystem("/foobar?insens=true")},
	}

	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		testDir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatalf("could not create temporary dir for testing: %s", err)
		}

		if fd, err := os.Create(filepath.Join(testDir, ".stfolder")); err != nil {
			t.Fatalf("could not create .stfolder: %s", err)
		} else {
			fd.Close()
		}

		defer func() {
			if err := os.RemoveAll(testDir); err != nil {
				t.Fatalf("could not remove test directory: %s", err)
			}
		}()

		filesystems = append(filesystems, testFS{runtime.GOOS, newBasicFilesystem(testDir)})
	}

	runTests(t, tests, filesystems)
}

func runTests(t *testing.T, tests []test, filesystems []testFS) {
	for _, filesystem := range filesystems {
		for _, test := range tests {
			name := fmt.Sprintf("%s_%s", test.name, filesystem.name)
			t.Run(name, func(t *testing.T) {
				test.impl(t, filesystem.fs)
				if err := cleanup(filesystem.fs); err != nil {
					t.Errorf("cleanup failed: %s", err)
				}
			})
		}
	}
}

func testFakeFSCaseInsensitive(t *testing.T, fs Filesystem) {
	bs1 := []byte("test")

	err := fs.Mkdir("/fUbar", 0755)
	if err != nil {
		t.Fatal(err)
	}

	fd1, err := fs.Create("fuBAR/SISYPHOS")
	if err != nil {
		t.Fatalf("could not create file: %s", err)
	}

	_, err = fd1.Write(bs1)
	if err != nil {
		t.Fatal(err)
	}

	fd1.Close()

	// Try reading from the same file with different filenames
	fd2, err := fs.Open("Fubar/Sisyphos")
	if err != nil {
		t.Fatalf("could not open file by its case-differing filename: %s", err)
	}

	if _, err := fd2.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}

	bs2, err := ioutil.ReadAll(fd2)
	if err != nil {
		t.Fatal(err)
	}

	fd2.Close()

	if len(bs1) != len(bs2) {
		t.Errorf("wrong number of bytes, expected %d, got %d", len(bs1), len(bs2))
	}
}

func testFakeFSCaseInsensitiveMkdirAll(t *testing.T, fs Filesystem) {
	err := fs.MkdirAll("/fOO/Bar/bAz", 0755)
	if err != nil {
		t.Fatal(err)
	}

	fd, err := fs.OpenFile("/foo/BaR/BaZ/tESt", os.O_CREATE, 0644)
	if err != nil {
		t.Fatal(err)
	}

	if err = fd.Close(); err != nil {
		t.Fatal(err)
	}

	if err = fs.Rename("/FOO/BAR/baz/tesT", "/foo/baR/BAZ/Qux"); err != nil {
		t.Fatal(err)
	}
}

func testFakeFSDirNames(t *testing.T, fs Filesystem) {
	testDirNames(t, fs)
}

func testDirNames(t *testing.T, fs Filesystem) {
	t.Helper()
	filenames := []string{"fOO", "Bar", "baz"}
	for _, filename := range filenames {
		fd, err := fs.Create("/" + filename)
		if err != nil {
			t.Errorf("Could not create %s: %s", filename, err)
		}
		fd.Close()
	}

	assertDir(t, fs, "/", filenames)
}

func assertDir(t *testing.T, fs Filesystem, directory string, filenames []string) {
	t.Helper()
	got, err := fs.DirNames(directory)
	if err != nil {
		t.Fatal(err)
	}

	if path.Clean(directory) == "/" {
		filenames = append(filenames, ".stfolder")
	}
	sort.Strings(filenames)
	sort.Strings(got)

	if !reflect.DeepEqual(got, filenames) {
		t.Errorf("want %s, got %s", filenames, got)
	}
}

func testFakeFSStatInsens(t *testing.T, fs Filesystem) {
	if err := fs.Mkdir("/foo", 0755); err != nil {
		t.Fatal(err)
	}

	fd, err := fs.Create("/Foo/aaa")
	if err != nil {
		t.Fatal(err)
	}

	fd.Close()

	info, err := fs.Stat("/FOO/AAA")
	if err != nil {
		t.Fatal(err)
	}

	if _, err = fs.Stat("/fOO/aAa"); err != nil {
		t.Fatal(err)
	}

	if info.Name() != "AAA" {
		t.Errorf("want AAA, got %s", info.Name())
	}

	fd1, err := fs.Open("/FOO/AAA")
	if err != nil {
		t.Fatal(err)
	}

	if info, err = fd1.Stat(); err != nil {
		t.Fatal(err)
	}

	fd2, err := fs.Open("Foo/aAa")
	if err != nil {
		t.Fatal(err)
	}

	if _, err = fd2.Stat(); err != nil {
		t.Fatal(err)
	}

	if info.Name() != "AAA" {
		t.Errorf("want AAA, got %s", info.Name())
	}

	fd1.Close()
	fd2.Close()

	assertDir(t, fs, "/", []string{"foo"})
	assertDir(t, fs, "/foo", []string{"aaa"})
}

func testFakeFSFileName(t *testing.T, fs Filesystem) {
	var testCases = []struct {
		create string
		open   string
	}{
		{"bar", "bar"},
	}

	for _, testCase := range testCases {
		if _, err := fs.Create(testCase.create); err != nil {
			t.Fatal(err)
		}

		fd, err := fs.Open(testCase.open)
		if err != nil {
			t.Fatal(err)
		}

		if got := fd.Name(); got != testCase.open {
			t.Errorf("want %s, got %s", testCase.open, got)
		}
	}
}

func testFakeFSFileNameInsens(t *testing.T, fs Filesystem) {
	var testCases = []struct {
		create string
		open   string
	}{
		{"BaZ", "bAz"},
	}

	for _, testCase := range testCases {
		fd, err := fs.Create(testCase.create)
		if err != nil {
			t.Fatal(err)
		}
		fd.Close()

		fd, err = fs.Open(testCase.open)
		if err != nil {
			t.Fatal(err)
		}

		defer fd.Close()

		if got := fd.Name(); got != testCase.open {
			t.Errorf("want %s, got %s", testCase.open, got)
		}
	}
}

func testFakeFSRename(t *testing.T, fs Filesystem) {
	if err := fs.MkdirAll("/foo/bar/baz", 0755); err != nil {
		t.Fatal(err)
	}

	fd, err := fs.Create("/foo/bar/baz/qux")
	if err != nil {
		t.Fatal(err)
	}
	fd.Close()

	if err := fs.Rename("/foo/bar/baz/qux", "/foo/baz/bar/qux"); err == nil {
		t.Errorf("rename to non-existent dir gave no error")
	}

	if err := fs.MkdirAll("/baz/bar/foo", 0755); err != nil {
		t.Fatal(err)
	}

	if err := fs.Rename("/foo/bar/baz/qux", "/baz/bar/foo/qux"); err != nil {
		t.Fatal(err)
	}

	var dirs = []struct {
		dir   string
		files []string
	}{
		{dir: "/", files: []string{"foo", "baz"}},
		{dir: "/foo", files: []string{"bar"}},
		{dir: "/foo/bar/baz", files: []string{}},
		{dir: "/baz/bar/foo", files: []string{"qux"}},
	}

	for _, dir := range dirs {
		assertDir(t, fs, dir.dir, dir.files)
	}

	if err := fs.Rename("/baz/bar/foo", "/baz/bar/FOO"); err != nil {
		t.Fatal(err)
	}

	assertDir(t, fs, "/baz/bar", []string{"FOO"})
	assertDir(t, fs, "/baz/bar/FOO", []string{"qux"})
}

func testFakeFSRenameInsensitive(t *testing.T, fs Filesystem) {
	if err := fs.MkdirAll("/baz/bar/foo", 0755); err != nil {
		t.Fatal(err)
	}

	if err := fs.MkdirAll("/foO/baR/baZ", 0755); err != nil {
		t.Fatal(err)
	}

	fd, err := fs.Create("/BAZ/BAR/FOO/QUX")
	if err != nil {
		t.Fatal(err)
	}

	fd.Close()

	if err := fs.Rename("/Baz/bAr/foO/QuX", "/Foo/Bar/Baz/qUUx"); err != nil {
		t.Fatal(err)
	}

	var dirs = []struct {
		dir   string
		files []string
	}{
		{dir: "/", files: []string{"foO", "baz"}},
		{dir: "/foo", files: []string{"baR"}},
		{dir: "/foo/bar/baz", files: []string{"qUUx"}},
		{dir: "/baz/bar/foo", files: []string{}},
	}

	for _, dir := range dirs {
		assertDir(t, fs, dir.dir, dir.files)
	}

	// the next rename can be done on Windows, and maybe elsewhere, but not on OS X
	// so we're shooting for the lowest common denominator
	if _, ok := fs.(*fakefs); ok || runtime.GOOS == "darwin" {
		if err := fs.Rename("/foo/bar/BAZ", "/FOO/BAR/bAz"); err == nil {
			t.Errorf("In-place case-only directory renames fail on OS X, should fail here, too: %s", err)
		}

		assertDir(t, fs, "/foo/bar", []string{"baZ"})
		assertDir(t, fs, "/fOO/bAr/baz", []string{"qUUx"})
	}

	if err := fs.Rename("foo/bar/baz/quux", "foo/bar/baz/Qux"); err != nil {
		t.Errorf("File rename failed: %s", err)
	}

	assertDir(t, fs, "/FOO/BAR/BAZ", []string{"Qux"})
}

func testFakeFSMkdir(t *testing.T, fs Filesystem) {
	if err := fs.Mkdir("/foo", 0755); err != nil {
		t.Fatal(err)
	}

	if _, err := fs.Stat("/foo"); err != nil {
		t.Fatal(err)
	}

	if err := fs.Mkdir("/foo", 0755); err == nil {
		t.Errorf("got no error while creating existing directory")
	}
}

func testFakeFSMkdirInsens(t *testing.T, fs Filesystem) {
	if err := fs.Mkdir("/foo", 0755); err != nil {
		t.Fatal(err)
	}

	if _, err := fs.Stat("/Foo"); err != nil {
		t.Fatal(err)
	}

	if err := fs.Mkdir("/FOO", 0755); err == nil {
		t.Errorf("got no error while creating existing directory")
	}
}

func testFakeFSOpenFile(t *testing.T, fs Filesystem) {
	if _, err := fs.OpenFile("foobar", os.O_RDONLY, 0664); err == nil {
		t.Errorf("got no error opening a non-existing file")
	}

	if _, err := fs.OpenFile("foobar", os.O_RDWR|os.O_CREATE, 0664); err != nil {
		t.Fatal(err)
	}

	if _, err := fs.OpenFile("foobar", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0664); err == nil {
		t.Errorf("created an existing file while told not to")
	}

	if _, err := fs.OpenFile("foobar", os.O_RDWR|os.O_CREATE, 0664); err != nil {
		t.Fatal(err)
	}

	if _, err := fs.OpenFile("foobar", os.O_RDWR, 0664); err != nil {
		t.Fatal(err)
	}
}

func testFakeFSOpenFileInsens(t *testing.T, fs Filesystem) {
	fd, err := fs.OpenFile("FooBar", os.O_RDONLY, 0664)
	if err == nil {
		t.Errorf("got no error opening a non-existing file")
	}
	if fd != nil {
		fd.Close()
	}

	fd, err = fs.OpenFile("fOObar", os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		t.Fatal(err)
	}
	if fd != nil {
		fd.Close()
	}

	fd, err = fs.OpenFile("fOoBaR", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0664)
	if err == nil {
		t.Errorf("created an existing file while told not to")
	}
	if fd != nil {
		fd.Close()
	}

	fd, err = fs.OpenFile("FoObAr", os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		t.Fatal(err)
	}
	if fd != nil {
		fd.Close()
	}

	fd, err = fs.OpenFile("FOOBAR", os.O_RDWR, 0664)
	if err != nil {
		t.Fatal(err)
	}
	if fd != nil {
		fd.Close()
	}
}

func testFakeFSRemoveAll(t *testing.T, fs Filesystem) {
	if err := fs.Mkdir("/foo", 0755); err != nil {
		t.Fatal(err)
	}

	filenames := []string{"bar", "baz", "qux"}
	for _, filename := range filenames {
		if _, err := fs.Create("/foo/" + filename); err != nil {
			t.Fatalf("Could not create %s: %s", filename, err)
		}
	}

	if err := fs.RemoveAll("/foo"); err != nil {
		t.Fatal(err)
	}

	if _, err := fs.Stat("/foo"); err == nil {
		t.Errorf("this should be an error, as file doesn not exist anymore")
	}

	if err := fs.RemoveAll("/foo/bar"); err != nil {
		t.Errorf("real systems don't return error here")
	}
}

func testFakeFSRemoveAllInsens(t *testing.T, fs Filesystem) {
	if err := fs.Mkdir("/Foo", 0755); err != nil {
		t.Fatal(err)
	}

	filenames := []string{"bar", "baz", "qux"}
	for _, filename := range filenames {
		fd, err := fs.Create("/FOO/" + filename)
		if err != nil {
			t.Fatalf("Could not create %s: %s", filename, err)
		}
		fd.Close()
	}

	if err := fs.RemoveAll("/fOo"); err != nil {
		t.Errorf("Could not remove dir: %s", err)
	}

	if _, err := fs.Stat("/foo"); err == nil {
		t.Errorf("this should be an error, as file doesn not exist anymore")
	}

	if err := fs.RemoveAll("/foO/bAr"); err != nil {
		t.Errorf("real systems don't return error here")
	}
}

func testFakeFSRemove(t *testing.T, fs Filesystem) {
	if err := fs.Mkdir("/Foo", 0755); err != nil {
		t.Fatal(err)
	}

	if _, err := fs.Create("/Foo/Bar"); err != nil {
		t.Fatal(err)
	}

	if err := fs.Remove("/Foo"); err == nil {
		t.Errorf("not empty, should give error")
	}

	if err := fs.Remove("/Foo/Bar"); err != nil {
		t.Fatal(err)
	}

	if err := fs.Remove("/Foo"); err != nil {
		t.Fatal(err)
	}
}

func testFakeFSRemoveInsens(t *testing.T, fs Filesystem) {
	if err := fs.Mkdir("/Foo", 0755); err != nil {
		t.Fatal(err)
	}

	fd, err := fs.Create("/Foo/Bar")
	if err != nil {
		t.Fatal(err)
	}
	fd.Close()

	if err := fs.Remove("/FOO"); err == nil || err == os.ErrNotExist {
		t.Errorf("not empty, should give error")
	}

	if err := fs.Remove("/Foo/BaR"); err != nil {
		t.Fatal(err)
	}

	if err := fs.Remove("/FoO"); err != nil {
		t.Fatal(err)
	}
}

func testFakeFSSameFile(t *testing.T, fs Filesystem) {
	if runtime.GOOS == "windows" {
		// windows time in not precise enough
		t.SkipNow()
	}

	if err := fs.Mkdir("/Foo", 0755); err != nil {
		t.Fatal(err)
	}

	filenames := []string{"Bar", "Baz", "/Foo/Bar"}
	for _, filename := range filenames {
		if _, err := fs.Create(filename); err != nil {
			t.Fatalf("Could not create %s: %s", filename, err)
		}
	}

	testCases := []struct {
		f1   string
		f2   string
		want bool
	}{
		{f1: "Bar", f2: "Baz", want: false},
		{f1: "Bar", f2: "/Foo/Bar", want: false},
		{"Bar", "Bar", true},
	}

	for _, test := range testCases {
		assertSameFile(t, fs, test.f1, test.f2, test.want)
	}
}

func testFakeFSSameFileInsens(t *testing.T, fs Filesystem) {
	if err := fs.Mkdir("/Foo", 0755); err != nil {
		t.Fatal(err)
	}

	filenames := []string{"Bar", "Baz"}
	for _, filename := range filenames {
		fd, err := fs.Create(filename)
		if err != nil {
			t.Errorf("Could not create %s: %s", filename, err)
		}
		fd.Close()
	}

	testCases := []struct {
		f1   string
		f2   string
		want bool
	}{
		{f1: "bAr", f2: "baZ", want: false},
		{"baz", "BAZ", true},
	}

	for _, test := range testCases {
		assertSameFile(t, fs, test.f1, test.f2, test.want)
	}
}

func assertSameFile(t *testing.T, fs Filesystem, f1, f2 string, want bool) {
	t.Helper()

	fi1, err := fs.Stat(f1)
	if err != nil {
		t.Fatal(err)
	}

	fi2, err := fs.Stat(f2)
	if err != nil {
		t.Fatal(err)
	}

	got := fs.SameFile(fi1, fi2)
	if got != want {
		t.Errorf("for \"%s\" and \"%s\" want SameFile %v, got %v", f1, f2, want, got)
	}
}

func testFakeFSCreateInsens(t *testing.T, fs Filesystem) {
	fd1, err := fs.Create("FOO")
	if err != nil {
		t.Fatal(err)
	}

	defer fd1.Close()

	fd2, err := fs.Create("fOo")
	if err != nil {
		t.Fatal(err)
	}

	defer fd2.Close()

	if fd1.Name() != "FOO" {
		t.Errorf("name of the file created as \"FOO\" is %s", fd1.Name())
	}

	if fd2.Name() != "fOo" {
		t.Errorf("name of created file \"fOo\" is %s", fd2.Name())
	}

	// one would expect DirNames to show the last variant, but in fact it shows
	// the original one
	assertDir(t, fs, "/", []string{"FOO"})
}

func cleanup(fs Filesystem) error {
	filenames, _ := fs.DirNames("/")
	for _, filename := range filenames {
		if filename != ".stfolder" {
			if err := fs.RemoveAll(filename); err != nil {
				return err
			}
		}
	}

	return nil
}
