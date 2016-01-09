// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package ignore

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/syncthing/syncthing/lib/fnmatch"
	"github.com/syncthing/syncthing/lib/sync"
)

type Result int

const (
	DontIgnore Result = iota
	Nuke
	Preserve
)

type Pattern struct {
	match  *regexp.Regexp
	result Result
}

func (p Pattern) String() string {
	switch p.result {
	case DontIgnore:
		return "(?exclude)" + p.match.String()
	case Nuke:
		return p.match.String()
	case Preserve:
		return "(?preserve)" + p.match.String()
	default:
		panic("Unsupported ignore result")
	}
}

type Matcher struct {
	patterns  []Pattern
	withCache bool
	matches   *cache
	curHash   string
	stop      chan struct{}
	mut       sync.Mutex
}

func New(withCache bool) *Matcher {
	m := &Matcher{
		withCache: withCache,
		stop:      make(chan struct{}),
		mut:       sync.NewMutex(),
	}
	if withCache {
		go m.clean(2 * time.Hour)
	}
	return m
}

func (m *Matcher) Load(file string, includePath string) error {
	// No locking, Parse() does the locking

	fd, err := os.Open(file)
	if err != nil {
		// We do a parse with empty patterns to clear out the hash, cache etc.
		m.Parse(&bytes.Buffer{}, file, includePath)
		return err
	}
	defer fd.Close()

	return m.Parse(fd, file, includePath)
}

func (m *Matcher) Parse(r io.Reader, file, includePath string) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	seen := map[string]bool{file: true}
	patterns, err := parseIgnoreFile(r, file, includePath, seen)
	// Error is saved and returned at the end. We process the patterns
	// (possibly blank) anyway.

	newHash := hashPatterns(patterns)
	if newHash == m.curHash {
		// We've already loaded exactly these patterns.
		return err
	}

	m.curHash = newHash
	m.patterns = patterns
	if m.withCache {
		m.matches = newCache(patterns)
	}

	return err
}

func (m *Matcher) Nuke(file string) (result bool) {
	return m.match(file) == Nuke
}

func (m *Matcher) Preserve(file string) (result bool) {
	return m.match(file) == Preserve
}

func (m *Matcher) Ignore(file string) (result bool) {
	return m.match(file) != DontIgnore
}

func (m *Matcher) match(file string) (result Result) {
	if m == nil {
		return DontIgnore
	}

	m.mut.Lock()
	defer m.mut.Unlock()

	if len(m.patterns) == 0 {
		return DontIgnore
	}

	if m.matches != nil {
		// Check the cache for a known result.
		res, ok := m.matches.get(file)
		if ok {
			return res
		}

		// Update the cache with the result at return time
		defer func() {
			m.matches.set(file, result)
		}()
	}

	// Check all the patterns for a match.
	for _, pattern := range m.patterns {
		if pattern.match.MatchString(file) {
			return pattern.result
		}
	}

	// Default to DontIgnore.
	return DontIgnore
}

// Patterns return a list of the loaded regexp patterns, as strings
func (m *Matcher) Patterns() []string {
	if m == nil {
		return nil
	}

	m.mut.Lock()
	defer m.mut.Unlock()

	patterns := make([]string, len(m.patterns))
	for i, pat := range m.patterns {
		patterns[i] = pat.String()
	}
	return patterns
}

func (m *Matcher) Hash() string {
	m.mut.Lock()
	defer m.mut.Unlock()
	return m.curHash
}

func (m *Matcher) Stop() {
	close(m.stop)
}

func (m *Matcher) clean(d time.Duration) {
	t := time.NewTimer(d / 2)
	for {
		select {
		case <-m.stop:
			return
		case <-t.C:
			m.mut.Lock()
			if m.matches != nil {
				m.matches.clean(d)
			}
			t.Reset(d / 2)
			m.mut.Unlock()
		}
	}
}

func hashPatterns(patterns []Pattern) string {
	h := md5.New()
	for _, pat := range patterns {
		h.Write([]byte(pat.String()))
		h.Write([]byte("\n"))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func loadIgnoreFile(file string, includePath string, seen map[string]bool) ([]Pattern, error) {
	if seen[file] {
		return nil, fmt.Errorf("Multiple include of ignore file %q", file)
	}
	seen[file] = true

	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	return parseIgnoreFile(fd, file, includePath, seen)
}

func parseIgnoreFile(fd io.Reader, currentFile, includePath string, seen map[string]bool) ([]Pattern, error) {
	var patterns []Pattern

	addPattern := func(line string) error {
		result := Nuke
		if strings.HasPrefix(line, "!") {
			line = line[1:]
			result = DontIgnore
		}

		if strings.HasPrefix(line, "(?preserve)") {
			line = line[11:]
			if result == Nuke {
				result = Preserve
			}
		}

		flags := fnmatch.PathName
		if strings.HasPrefix(line, "(?i)") {
			line = line[4:]
			flags |= fnmatch.CaseFold
		}

		if strings.HasPrefix(line, "/") {
			// Pattern is rooted in the current dir only
			exp, err := fnmatch.Convert(line[1:], flags)
			if err != nil {
				return fmt.Errorf("invalid pattern %q in ignore file", line)
			}
			patterns = append(patterns, Pattern{exp, result})
		} else if strings.HasPrefix(line, "**/") {
			// Add the pattern as is, and without **/ so it matches in current dir
			exp, err := fnmatch.Convert(line, flags)
			if err != nil {
				return fmt.Errorf("invalid pattern %q in ignore file", line)
			}
			patterns = append(patterns, Pattern{exp, result})

			exp, err = fnmatch.Convert(line[3:], flags)
			if err != nil {
				return fmt.Errorf("invalid pattern %q in ignore file", line)
			}
			patterns = append(patterns, Pattern{exp, result})
		} else if strings.HasPrefix(line, "#include ") {
			includeRel := line[len("#include "):]
			includeFile := filepath.Join(filepath.Dir(includePath), includeRel)
			includes, err := loadIgnoreFile(includeFile, filepath.Dir(includeFile), seen)
			if err != nil {
				return fmt.Errorf("include of %q: %v", includeRel, err)
			}
			patterns = append(patterns, includes...)
		} else {
			// Path name or pattern, add it so it matches files both in
			// current directory and subdirs.
			exp, err := fnmatch.Convert(line, flags)
			if err != nil {
				return fmt.Errorf("invalid pattern %q in ignore file", line)
			}
			patterns = append(patterns, Pattern{exp, result})

			exp, err = fnmatch.Convert("**/"+line, flags)
			if err != nil {
				return fmt.Errorf("invalid pattern %q in ignore file", line)
			}
			patterns = append(patterns, Pattern{exp, result})
		}
		return nil
	}

	scanner := bufio.NewScanner(fd)
	var err error
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case line == "":
			continue
		case strings.HasPrefix(line, "//"):
			continue
		case strings.HasPrefix(line, "#"):
			err = addPattern(line)
		case strings.HasSuffix(line, "/**"):
			err = addPattern(line)
		case strings.HasSuffix(line, "/"):
			err = addPattern(line)
			if err == nil {
				err = addPattern(line + "**")
			}
		default:
			err = addPattern(line)
			if err == nil {
				err = addPattern(line + "/**")
			}
		}
		if err != nil {
			return nil, err
		}
	}

	return patterns, nil
}
