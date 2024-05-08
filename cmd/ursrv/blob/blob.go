// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package blob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/syncthing/syncthing/cmd/ursrv/report"
	"github.com/syncthing/syncthing/lib/ur"
	"github.com/syncthing/syncthing/lib/ur/contract"
)

const (
	USAGE_PREFIX      = "UR" // contract.Report
	AGGREGATED_PREFIX = "AR" // report.AggregatedReport
)

func NewBlobStorage(s3Config S3Config) Store {
	// If S3-compatible credentials are provided, use those.
	if s3Config.isSet() {
		s3, err := NewS3(s3Config)
		if err == nil {
			return s3
		}
	}

	// Fall back to local storage.
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Println("Could not get user home directory", "error", err)
		dir = os.TempDir()
	}

	dir = filepath.Join(dir, ".ursrv", "blob")

	return NewDisk(dir)
}

type Store interface {
	Put(_ string, _ []byte) error
	Get(_ string) ([]byte, error)
	Delete(_ string) error
	Iterate(_ context.Context, _ string, _ func([]byte) bool) error
	IterateFromDate(_ context.Context, _ string, _ time.Time, _ func([]byte) bool) error
	CountFromDate(_ string, _ time.Time) (int, error)
}

type UrsrvStore struct {
	Store
}

func NewUrsrvStore(s Store) *UrsrvStore {
	return &UrsrvStore{s}
}

func usageReportKey(when time.Time, uniqueId string) string {
	return fmt.Sprintf("%s/%s-%s.json", USAGE_PREFIX, when.UTC().Format("20060102"), uniqueId)
}

func aggregatedReportKey(when time.Time) string {
	return fmt.Sprintf("%s/%s.json", AGGREGATED_PREFIX, when.UTC().Format("20060102"))
}

// Usage Reports.

func (s *UrsrvStore) PutUsageReport(rep contract.Report, received time.Time) error {
	key := usageReportKey(received, rep.UniqueID)

	// Check if we already have a report for this instance today.
	if data, err := s.Store.Get(key); err == nil && len(data) != 0 {
		return errors.New("already exists")
	}

	bs, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	return s.Store.Put(key, bs)
}

func (s *UrsrvStore) ListUsageReportsForDate(when time.Time) ([]contract.Report, error) {
	ctx := context.Background()
	prefix, _ := strings.CutSuffix(usageReportKey(when, ""), ".json")

	var res []contract.Report
	var rep contract.Report

	err := s.Store.Iterate(ctx, prefix, func(b []byte) bool {
		err := json.Unmarshal(b, &rep)
		if err != nil {
			return true
		}
		res = append(res, rep)
		return true
	})

	return res, err
}

// Aggregated reports.

func (s *UrsrvStore) PutAggregatedReport(rep *ur.Aggregation) error {
	key := aggregatedReportKey(time.Unix(rep.Date, 0))
	bs, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	return s.Store.Put(key, bs)
}

func (s *UrsrvStore) ListAggregatedReports(from time.Time) ([]report.AggregatedReport, error) {
	ctx := context.Background()

	var res []report.AggregatedReport
	var rep report.AggregatedReport
	err := s.Store.IterateFromDate(ctx, AGGREGATED_PREFIX, from, func(b []byte) bool {
		err := json.Unmarshal(b, &rep)
		if err != nil {
			return true
		}
		res = append(res, rep)
		return true
	})

	return res, err
}

func (s *UrsrvStore) LatestAggregatedReport() (report.AggregatedReport, error) {
	var rep report.AggregatedReport

	// Requires an aggregated report of the day before, which in practise should
	// always be the case.
	date := time.Now().UTC().AddDate(0, 0, -1)
	key := aggregatedReportKey(date)
	data, err := s.Store.Get(key)
	if err != nil {
		// In practise this shouldn't happen, but we can look one day further
		// back.
		date := date.AddDate(0, 0, -1)
		key := aggregatedReportKey(date)
		data, err = s.Store.Get(key)
		if err != nil {
			return rep, errors.New("no aggregated report found")
		}
	}

	err = json.Unmarshal(data, &rep)
	return rep, err
}

func (s *UrsrvStore) CountAggregatedReports(from time.Time) (int, error) {
	prefix := AGGREGATED_PREFIX + "/"
	return s.Count(prefix, from)
}

// Common.

func (s *UrsrvStore) Count(prefix string, from time.Time) (int, error) {
	return s.Store.CountFromDate(prefix, from)
}

func hasValidDate(key string, from time.Time) bool {
	key, found := strings.CutSuffix(key, ".json")
	if !found {
		return false
	}
	keyDate, err := time.Parse(time.DateOnly, key)
	if err != nil {
		return false
	}
	if keyDate.Before(from) {
		return false
	}
	return true
}

func commonTimestampPrefix(a, b time.Time) string {
	aFormatted := a.UTC().Format("20060102")
	bFormatted := b.UTC().Format("20060102")
	prefixLen := 0
	for i := range aFormatted {
		if aFormatted[i] != bFormatted[i] {
			break
		}
		prefixLen = i + 1
	}
	return aFormatted[:prefixLen]
}
