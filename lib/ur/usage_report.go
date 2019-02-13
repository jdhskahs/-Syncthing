// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package ur

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/syncthing/syncthing/lib/build"
	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/connections"
	"github.com/syncthing/syncthing/lib/db"
	"github.com/syncthing/syncthing/lib/dialer"
	"github.com/syncthing/syncthing/lib/model"
	"github.com/syncthing/syncthing/lib/protocol"
	"github.com/syncthing/syncthing/lib/scanner"
	"github.com/syncthing/syncthing/lib/stats"
	"github.com/syncthing/syncthing/lib/upgrade"
	"github.com/syncthing/syncthing/lib/versioner"
)

// Current version number of the usage report, for acceptance purposes. If
// fields are added or changed this integer must be incremented so that users
// are prompted for acceptance of the new report.
const UsageReportVersion = 3

var StartTime = time.Now()

type Service struct {
	cfg                ConfigIntf
	model              ModelIntf
	connectionsService ConnectionsIntf
	noUpgrade          bool
	forceRun           chan struct{}
	stop               chan struct{}
	stopped            chan struct{}
	stopMut            sync.RWMutex
}

func New(cfg ConfigIntf, model ModelIntf, connectionsService ConnectionsIntf, noUpgrade bool) *Service {
	svc := &Service{
		cfg:                cfg,
		model:              model,
		connectionsService: connectionsService,
		noUpgrade:          noUpgrade,
		forceRun:           make(chan struct{}),
		stop:               make(chan struct{}),
		stopped:            make(chan struct{}),
	}
	close(svc.stopped) // Not yet running, dont block on Stop()
	cfg.Subscribe(svc)
	return svc
}

// ReportData returns the data to be sent in a usage report with the currently
// configured usage reporting version.
func (s *Service) ReportData() map[string]interface{} {
	return s.reportData(UsageReportVersion, false)
}

// ReportDataPreview returns a preview of the data to be sent in a usage report
// with the given version.
func (s *Service) ReportDataPreview(urversion int) map[string]interface{} {
	return s.reportData(urversion, true)
}

func (s *Service) reportData(urversion int, preview bool) map[string]interface{} {
	opts := s.cfg.Options()
	res := make(map[string]interface{})
	res["urVersion"] = urversion
	res["uniqueID"] = opts.URUniqueID
	res["version"] = build.Version
	res["longVersion"] = build.LongVersion
	res["platform"] = runtime.GOOS + "-" + runtime.GOARCH
	res["numFolders"] = len(s.cfg.Folders())
	res["numDevices"] = len(s.cfg.Devices())

	var totFiles, maxFiles int
	var totBytes, maxBytes int64
	for folderID := range s.cfg.Folders() {
		global := s.model.GlobalSize(folderID)
		totFiles += int(global.Files)
		totBytes += global.Bytes
		if int(global.Files) > maxFiles {
			maxFiles = int(global.Files)
		}
		if global.Bytes > maxBytes {
			maxBytes = global.Bytes
		}
	}

	res["totFiles"] = totFiles
	res["folderMaxFiles"] = maxFiles
	res["totMiB"] = totBytes / 1024 / 1024
	res["folderMaxMiB"] = maxBytes / 1024 / 1024

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	res["memoryUsageMiB"] = (mem.Sys - mem.HeapReleased) / 1024 / 1024
	res["sha256Perf"] = CpuBench(5, 125*time.Millisecond, false)
	res["hashPerf"] = CpuBench(5, 125*time.Millisecond, true)

	bytes, err := memorySize()
	if err == nil {
		res["memorySize"] = bytes / 1024 / 1024
	}
	res["numCPU"] = runtime.NumCPU()

	var rescanIntvs []int
	folderUses := map[string]int{
		"sendonly":            0,
		"sendreceive":         0,
		"receiveonly":         0,
		"ignorePerms":         0,
		"ignoreDelete":        0,
		"autoNormalize":       0,
		"simpleVersioning":    0,
		"externalVersioning":  0,
		"staggeredVersioning": 0,
		"trashcanVersioning":  0,
	}
	for _, cfg := range s.cfg.Folders() {
		rescanIntvs = append(rescanIntvs, cfg.RescanIntervalS)

		switch cfg.Type {
		case config.FolderTypeSendOnly:
			folderUses["sendonly"]++
		case config.FolderTypeSendReceive:
			folderUses["sendreceive"]++
		case config.FolderTypeReceiveOnly:
			folderUses["receiveonly"]++
		}
		if cfg.IgnorePerms {
			folderUses["ignorePerms"]++
		}
		if cfg.IgnoreDelete {
			folderUses["ignoreDelete"]++
		}
		if cfg.AutoNormalize {
			folderUses["autoNormalize"]++
		}
		if cfg.Versioning.Type != "" {
			folderUses[cfg.Versioning.Type+"Versioning"]++
		}
	}
	sort.Ints(rescanIntvs)
	res["rescanIntvs"] = rescanIntvs
	res["folderUses"] = folderUses

	deviceUses := map[string]int{
		"introducer":       0,
		"customCertName":   0,
		"compressAlways":   0,
		"compressMetadata": 0,
		"compressNever":    0,
		"dynamicAddr":      0,
		"staticAddr":       0,
	}
	for _, cfg := range s.cfg.Devices() {
		if cfg.Introducer {
			deviceUses["introducer"]++
		}
		if cfg.CertName != "" && cfg.CertName != "syncthing" {
			deviceUses["customCertName"]++
		}
		if cfg.Compression == protocol.CompressAlways {
			deviceUses["compressAlways"]++
		} else if cfg.Compression == protocol.CompressMetadata {
			deviceUses["compressMetadata"]++
		} else if cfg.Compression == protocol.CompressNever {
			deviceUses["compressNever"]++
		}
		for _, addr := range cfg.Addresses {
			if addr == "dynamic" {
				deviceUses["dynamicAddr"]++
			} else {
				deviceUses["staticAddr"]++
			}
		}
	}
	res["deviceUses"] = deviceUses

	defaultAnnounceServersDNS, defaultAnnounceServersIP, otherAnnounceServers := 0, 0, 0
	for _, addr := range opts.GlobalAnnServers {
		if addr == "default" || addr == "default-v4" || addr == "default-v6" {
			defaultAnnounceServersDNS++
		} else {
			otherAnnounceServers++
		}
	}
	res["announce"] = map[string]interface{}{
		"globalEnabled":     opts.GlobalAnnEnabled,
		"localEnabled":      opts.LocalAnnEnabled,
		"defaultServersDNS": defaultAnnounceServersDNS,
		"defaultServersIP":  defaultAnnounceServersIP,
		"otherServers":      otherAnnounceServers,
	}

	defaultRelayServers, otherRelayServers := 0, 0
	for _, addr := range s.cfg.ListenAddresses() {
		switch {
		case addr == "dynamic+https://relays.syncthing.net/endpoint":
			defaultRelayServers++
		case strings.HasPrefix(addr, "relay://") || strings.HasPrefix(addr, "dynamic+http"):
			otherRelayServers++
		}
	}
	res["relays"] = map[string]interface{}{
		"enabled":        defaultRelayServers+otherAnnounceServers > 0,
		"defaultServers": defaultRelayServers,
		"otherServers":   otherRelayServers,
	}

	res["usesRateLimit"] = opts.MaxRecvKbps > 0 || opts.MaxSendKbps > 0

	res["upgradeAllowedManual"] = !(upgrade.DisabledByCompilation || s.noUpgrade)
	res["upgradeAllowedAuto"] = !(upgrade.DisabledByCompilation || s.noUpgrade) && opts.AutoUpgradeIntervalH > 0
	res["upgradeAllowedPre"] = !(upgrade.DisabledByCompilation || s.noUpgrade) && opts.AutoUpgradeIntervalH > 0 && opts.UpgradeToPreReleases

	if urversion >= 3 {
		res["uptime"] = s.UptimeS()
		res["natType"] = s.connectionsService.NATType()
		res["alwaysLocalNets"] = len(opts.AlwaysLocalNets) > 0
		res["cacheIgnoredFiles"] = opts.CacheIgnoredFiles
		res["overwriteRemoteDeviceNames"] = opts.OverwriteRemoteDevNames
		res["progressEmitterEnabled"] = opts.ProgressUpdateIntervalS > -1
		res["customDefaultFolderPath"] = opts.DefaultFolderPath != "~"
		res["customTrafficClass"] = opts.TrafficClass != 0
		res["customTempIndexMinBlocks"] = opts.TempIndexMinBlocks != 10
		res["temporariesDisabled"] = opts.KeepTemporariesH == 0
		res["temporariesCustom"] = opts.KeepTemporariesH != 24
		res["limitBandwidthInLan"] = opts.LimitBandwidthInLan
		res["customReleaseURL"] = opts.ReleasesURL != "https://upgrades.syncthing.net/meta.json"
		res["restartOnWakeup"] = opts.RestartOnWakeup

		folderUsesV3 := map[string]int{
			"scanProgressDisabled":    0,
			"conflictsDisabled":       0,
			"conflictsUnlimited":      0,
			"conflictsOther":          0,
			"disableSparseFiles":      0,
			"disableTempIndexes":      0,
			"alwaysWeakHash":          0,
			"customWeakHashThreshold": 0,
			"fsWatcherEnabled":        0,
		}
		pullOrder := make(map[string]int)
		filesystemType := make(map[string]int)
		var fsWatcherDelays []int
		for _, cfg := range s.cfg.Folders() {
			if cfg.ScanProgressIntervalS < 0 {
				folderUsesV3["scanProgressDisabled"]++
			}
			if cfg.MaxConflicts == 0 {
				folderUsesV3["conflictsDisabled"]++
			} else if cfg.MaxConflicts < 0 {
				folderUsesV3["conflictsUnlimited"]++
			} else {
				folderUsesV3["conflictsOther"]++
			}
			if cfg.DisableSparseFiles {
				folderUsesV3["disableSparseFiles"]++
			}
			if cfg.DisableTempIndexes {
				folderUsesV3["disableTempIndexes"]++
			}
			if cfg.WeakHashThresholdPct < 0 {
				folderUsesV3["alwaysWeakHash"]++
			} else if cfg.WeakHashThresholdPct != 25 {
				folderUsesV3["customWeakHashThreshold"]++
			}
			if cfg.FSWatcherEnabled {
				folderUsesV3["fsWatcherEnabled"]++
			}
			pullOrder[cfg.Order.String()]++
			filesystemType[cfg.FilesystemType.String()]++
			fsWatcherDelays = append(fsWatcherDelays, cfg.FSWatcherDelayS)
		}
		sort.Ints(fsWatcherDelays)
		folderUsesV3Interface := map[string]interface{}{
			"pullOrder":       pullOrder,
			"filesystemType":  filesystemType,
			"fsWatcherDelays": fsWatcherDelays,
		}
		for key, value := range folderUsesV3 {
			folderUsesV3Interface[key] = value
		}
		res["folderUsesV3"] = folderUsesV3Interface

		guiCfg := s.cfg.GUI()
		// Anticipate multiple GUI configs in the future, hence store counts.
		guiStats := map[string]int{
			"enabled":                   0,
			"useTLS":                    0,
			"useAuth":                   0,
			"insecureAdminAccess":       0,
			"debugging":                 0,
			"insecureSkipHostCheck":     0,
			"insecureAllowFrameLoading": 0,
			"listenLocal":               0,
			"listenUnspecified":         0,
		}
		theme := make(map[string]int)
		if guiCfg.Enabled {
			guiStats["enabled"]++
			if guiCfg.UseTLS() {
				guiStats["useTLS"]++
			}
			if len(guiCfg.User) > 0 && len(guiCfg.Password) > 0 {
				guiStats["useAuth"]++
			}
			if guiCfg.InsecureAdminAccess {
				guiStats["insecureAdminAccess"]++
			}
			if guiCfg.Debugging {
				guiStats["debugging"]++
			}
			if guiCfg.InsecureSkipHostCheck {
				guiStats["insecureSkipHostCheck"]++
			}
			if guiCfg.InsecureAllowFrameLoading {
				guiStats["insecureAllowFrameLoading"]++
			}

			addr, err := net.ResolveTCPAddr("tcp", guiCfg.Address())
			if err == nil {
				if addr.IP.IsLoopback() {
					guiStats["listenLocal"]++
				} else if addr.IP.IsUnspecified() {
					guiStats["listenUnspecified"]++
				}
			}

			theme[guiCfg.Theme]++
		}
		guiStatsInterface := map[string]interface{}{
			"theme": theme,
		}
		for key, value := range guiStats {
			guiStatsInterface[key] = value
		}
		res["guiStats"] = guiStatsInterface
	}

	for key, value := range s.model.UsageReportingStats(urversion, preview) {
		res[key] = value
	}

	return res
}

func (s *Service) UptimeS() int {
	return int(time.Since(StartTime).Seconds())
}

func (s *Service) sendUsageReport() error {
	d := s.ReportData()
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(d); err != nil {
		return err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Dial:  dialer.Dial,
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: s.cfg.Options().URPostInsecurely,
			},
		},
	}
	_, err := client.Post(s.cfg.Options().URURL, "application/json", &b)
	return err
}

func (s *Service) Serve() {
	s.stopMut.Lock()
	s.stop = make(chan struct{})
	s.stopped = make(chan struct{})
	s.stopMut.Unlock()
	t := time.NewTimer(time.Duration(s.cfg.Options().URInitialDelayS) * time.Second)
	s.stopMut.RLock()
	defer func() {
		close(s.stopped)
		s.stopMut.RUnlock()
	}()
	for {
		select {
		case <-s.stop:
			return
		case <-s.forceRun:
			t.Reset(0)
		case <-t.C:
			if s.cfg.Options().URAccepted >= 2 {
				err := s.sendUsageReport()
				if err != nil {
					l.Infoln("Usage report:", err)
				} else {
					l.Infof("Sent usage report (version %d)", s.cfg.Options().URAccepted)
				}
			}
			t.Reset(24 * time.Hour) // next report tomorrow
		}
	}
}

func (s *Service) VerifyConfiguration(from, to config.Configuration) error {
	return nil
}

func (s *Service) CommitConfiguration(from, to config.Configuration) bool {
	if from.Options.URAccepted != to.Options.URAccepted || from.Options.URUniqueID != to.Options.URUniqueID || from.Options.URURL != to.Options.URURL {
		s.stopMut.RLock()
		select {
		case s.forceRun <- struct{}{}:
		case <-s.stop:
		}
		s.stopMut.RUnlock()
	}
	return true
}

func (s *Service) Stop() {
	s.stopMut.RLock()
	close(s.stop)
	<-s.stopped
	s.stopMut.RUnlock()
}

func (*Service) String() string {
	return "ur.Service"
}

// CpuBench returns CPU performance as a measure of single threaded SHA-256 MiB/s
func CpuBench(iterations int, duration time.Duration, useWeakHash bool) float64 {
	dataSize := 16 * protocol.MinBlockSize
	bs := make([]byte, dataSize)
	rand.Reader.Read(bs)

	var perf float64
	for i := 0; i < iterations; i++ {
		if v := cpuBenchOnce(duration, useWeakHash, bs); v > perf {
			perf = v
		}
	}
	blocksResult = nil
	return perf
}

var blocksResult []protocol.BlockInfo // so the result is not optimized away

func cpuBenchOnce(duration time.Duration, useWeakHash bool, bs []byte) float64 {
	t0 := time.Now()
	b := 0
	for time.Since(t0) < duration {
		r := bytes.NewReader(bs)
		blocksResult, _ = scanner.Blocks(context.TODO(), r, protocol.MinBlockSize, int64(len(bs)), nil, useWeakHash)
		b += len(bs)
	}
	d := time.Since(t0)
	return float64(int(float64(b)/d.Seconds()/(1<<20)*100)) / 100
}

type ModelIntf interface {
	GlobalDirectoryTree(folder, prefix string, levels int, dirsonly bool) map[string]interface{}
	Completion(device protocol.DeviceID, folder string) model.FolderCompletion
	Override(folder string)
	Revert(folder string)
	NeedFolderFiles(folder string, page, perpage int) ([]db.FileInfoTruncated, []db.FileInfoTruncated, []db.FileInfoTruncated)
	RemoteNeedFolderFiles(device protocol.DeviceID, folder string, page, perpage int) ([]db.FileInfoTruncated, error)
	LocalChangedFiles(folder string, page, perpage int) []db.FileInfoTruncated
	NeedSize(folder string) db.Counts
	ConnectionStats() map[string]interface{}
	DeviceStatistics() map[string]stats.DeviceStatistics
	FolderStatistics() map[string]stats.FolderStatistics
	CurrentFolderFile(folder string, file string) (protocol.FileInfo, bool)
	CurrentGlobalFile(folder string, file string) (protocol.FileInfo, bool)
	ResetFolder(folder string)
	Availability(folder string, file protocol.FileInfo, block protocol.BlockInfo) []model.Availability
	GetIgnores(folder string) ([]string, []string, error)
	GetFolderVersions(folder string) (map[string][]versioner.FileVersion, error)
	RestoreFolderVersions(folder string, versions map[string]time.Time) (map[string]string, error)
	SetIgnores(folder string, content []string) error
	DelayScan(folder string, next time.Duration)
	ScanFolder(folder string) error
	ScanFolders() map[string]error
	ScanFolderSubdirs(folder string, subs []string) error
	BringToFront(folder, file string)
	Connection(deviceID protocol.DeviceID) (connections.Connection, bool)
	GlobalSize(folder string) db.Counts
	LocalSize(folder string) db.Counts
	ReceiveOnlyChangedSize(folder string) db.Counts
	CurrentSequence(folder string) (int64, bool)
	RemoteSequence(folder string) (int64, bool)
	State(folder string) (string, time.Time, error)
	UsageReportingStats(version int, preview bool) map[string]interface{}
	FolderErrors(folder string) ([]model.FileError, error)
	WatchError(folder string) error
}

type ConfigIntf interface {
	GUI() config.GUIConfiguration
	LDAP() config.LDAPConfiguration
	RawCopy() config.Configuration
	Options() config.OptionsConfiguration
	Replace(cfg config.Configuration) (config.Waiter, error)
	Subscribe(c config.Committer)
	Unsubscribe(c config.Committer)
	Folders() map[string]config.FolderConfiguration
	Devices() map[protocol.DeviceID]config.DeviceConfiguration
	SetDevice(config.DeviceConfiguration) (config.Waiter, error)
	SetDevices([]config.DeviceConfiguration) (config.Waiter, error)
	Save() error
	ListenAddresses() []string
	RequiresRestart() bool
}

type ConnectionsIntf interface {
	Status() map[string]interface{}
	NATType() string
}
