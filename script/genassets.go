// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

// +build ignore

package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var tpl = template.Must(template.New("assets").Parse(`// Code generated by genassets.go - DO NOT EDIT.

package auto

const Generated int64 = {{.Generated}}

func Assets() map[string][]byte {
	var assets = make(map[string][]byte, {{.Assets | len}})
{{range $asset := .Assets}}
	assets["{{$asset.Name}}"] = {{$asset.Data}}{{end}}
	return assets
}

`))

type asset struct {
	Name string
	Data string
}

var assets []asset

func walkerFor(basePath string) filepath.WalkFunc {
	return func(name string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(filepath.Base(name), ".") {
			// Skip dotfiles
			return nil
		}

		if info.Mode().IsRegular() {
			fd, err := os.Open(name)
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			gw := gzip.NewWriter(&buf)
			io.Copy(gw, fd)
			fd.Close()
			gw.Flush()
			gw.Close()

			name, _ = filepath.Rel(basePath, name)
			assets = append(assets, asset{
				Name: filepath.ToSlash(name),
				Data: fmt.Sprintf("[]byte(%q)", buf.String()),
			})
		}

		return nil
	}
}

type templateVars struct {
	Assets    []asset
	Generated int64
}

func main() {
	outfile := flag.String("o", "", "Name of output file (default stdout)")
	flag.Parse()

	filepath.Walk(flag.Arg(0), walkerFor(flag.Arg(0)))
	var buf bytes.Buffer

	// Generated time is now, except if the SOURCE_DATE_EPOCH environment
	// variable is set (for reproducible builds).
	generated := time.Now().Unix()
	if s, _ := strconv.ParseInt(os.Getenv("SOURCE_DATE_EPOCH"), 10, 64); s > 0 {
		generated = s
	}

	tpl.Execute(&buf, templateVars{
		Assets:    assets,
		Generated: generated,
	})
	bs, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	out := io.Writer(os.Stdout)
	if *outfile != "" {
		out, err = os.Create(*outfile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	out.Write(bs)
}
