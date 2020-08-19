// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: lib/config/folderconfiguration.proto

package config

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	fs "github.com/syncthing/syncthing/lib/fs"
	github_com_syncthing_syncthing_lib_protocol "github.com/syncthing/syncthing/lib/protocol"
	_ "github.com/syncthing/syncthing/proto/ext"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type FolderDeviceConfiguration struct {
	DeviceID     github_com_syncthing_syncthing_lib_protocol.DeviceID `protobuf:"bytes,1,opt,name=device_id,json=deviceId,proto3,customtype=github.com/syncthing/syncthing/lib/protocol.DeviceID" json:"deviceID" xml:"id,attr"`
	IntroducedBy github_com_syncthing_syncthing_lib_protocol.DeviceID `protobuf:"bytes,2,opt,name=introduced_by,json=introducedBy,proto3,customtype=github.com/syncthing/syncthing/lib/protocol.DeviceID" json:"introducedBy" xml:"introducedBy,attr"`
}

func (m *FolderDeviceConfiguration) Reset()         { *m = FolderDeviceConfiguration{} }
func (m *FolderDeviceConfiguration) String() string { return proto.CompactTextString(m) }
func (*FolderDeviceConfiguration) ProtoMessage()    {}
func (*FolderDeviceConfiguration) Descriptor() ([]byte, []int) {
	return fileDescriptor_44a9785876ed3afa, []int{0}
}
func (m *FolderDeviceConfiguration) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FolderDeviceConfiguration.Unmarshal(m, b)
}
func (m *FolderDeviceConfiguration) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FolderDeviceConfiguration.Marshal(b, m, deterministic)
}
func (m *FolderDeviceConfiguration) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FolderDeviceConfiguration.Merge(m, src)
}
func (m *FolderDeviceConfiguration) XXX_Size() int {
	return xxx_messageInfo_FolderDeviceConfiguration.Size(m)
}
func (m *FolderDeviceConfiguration) XXX_DiscardUnknown() {
	xxx_messageInfo_FolderDeviceConfiguration.DiscardUnknown(m)
}

var xxx_messageInfo_FolderDeviceConfiguration proto.InternalMessageInfo

type FolderConfiguration struct {
	ID                      string                      `protobuf:"bytes,1,opt,name=id,proto3" json:"id" xml:"id,attr"`
	Label                   string                      `protobuf:"bytes,2,opt,name=label,proto3" json:"label" xml:"label,attr" restart:"false"`
	FilesystemType          fs.FilesystemType           `protobuf:"varint,3,opt,name=filesystem_type,json=filesystemType,proto3,enum=fs.FilesystemType" json:"filesystemType" xml:"filesystemType"`
	Path                    string                      `protobuf:"bytes,4,opt,name=path,proto3" json:"path" xml:"path,attr"`
	Type                    FolderType                  `protobuf:"varint,5,opt,name=type,proto3,enum=config.FolderType" json:"type" xml:"type,attr"`
	Devices                 []FolderDeviceConfiguration `protobuf:"bytes,6,rep,name=devices,proto3" json:"devices" xml:"device"`
	RescanIntervalS         int                         `protobuf:"varint,7,opt,name=rescan_interval_s,json=rescanIntervalS,proto3,casttype=int" json:"rescanIntervalS" xml:"rescanIntervalS,attr" default:"3600"`
	FSWatcherEnabled        bool                        `protobuf:"varint,8,opt,name=fs_watcher_enabled,json=fsWatcherEnabled,proto3" json:"fsWatcherEnabled" xml:"fsWatcherEnabled,attr" default:"true"`
	FSWatcherDelayS         int                         `protobuf:"varint,9,opt,name=fs_watcher_delay_s,json=fsWatcherDelayS,proto3,casttype=int" json:"fsWatcherDelayS" xml:"fsWatcherDelayS,attr" default:"10"`
	IgnorePerms             bool                        `protobuf:"varint,10,opt,name=ignore_perms,json=ignorePerms,proto3" json:"ignorePerms" xml:"ignorePerms,attr"`
	AutoNormalize           bool                        `protobuf:"varint,11,opt,name=auto_normalize,json=autoNormalize,proto3" json:"autoNormalize" xml:"autoNormalize,attr" default:"true"`
	MinDiskFree             Size                        `protobuf:"bytes,12,opt,name=min_disk_free,json=minDiskFree,proto3" json:"minDiskFree" xml:"minDiskFree"`
	Versioning              VersioningConfiguration     `protobuf:"bytes,13,opt,name=versioning,proto3" json:"versioning" xml:"versioning"`
	Copiers                 int                         `protobuf:"varint,14,opt,name=copiers,proto3,casttype=int" json:"copiers" xml:"copiers"`
	PullerMaxPendingKiB     int                         `protobuf:"varint,15,opt,name=puller_max_pending_kib,json=pullerMaxPendingKib,proto3,casttype=int" json:"pullerMaxPendingKiB" xml:"pullerMaxPendingKiB"`
	Hashers                 int                         `protobuf:"varint,16,opt,name=hashers,proto3,casttype=int" json:"hashers" xml:"hashers"`
	Order                   PullOrder                   `protobuf:"varint,17,opt,name=order,proto3,enum=config.PullOrder" json:"order" xml:"order"`
	IgnoreDelete            bool                        `protobuf:"varint,18,opt,name=ignore_delete,json=ignoreDelete,proto3" json:"ignoreDelete" xml:"ignoreDelete"`
	ScanProgressIntervalS   int                         `protobuf:"varint,19,opt,name=scan_progress_interval_s,json=scanProgressIntervalS,proto3,casttype=int" json:"scanProgressIntervalS" xml:"scanProgressIntervalS"`
	PullerPauseS            int                         `protobuf:"varint,20,opt,name=puller_pause_s,json=pullerPauseS,proto3,casttype=int" json:"pullerPauseS" xml:"pullerPauseS"`
	MaxConflicts            int                         `protobuf:"varint,21,opt,name=max_conflicts,json=maxConflicts,proto3,casttype=int" json:"maxConflicts" xml:"maxConflicts" default:"-1"`
	DisableSparseFiles      bool                        `protobuf:"varint,22,opt,name=disable_sparse_files,json=disableSparseFiles,proto3" json:"disableSparseFiles" xml:"disableSparseFiles"`
	DisableTempIndexes      bool                        `protobuf:"varint,23,opt,name=disable_temp_indexes,json=disableTempIndexes,proto3" json:"disableTempIndexes" xml:"disableTempIndexes"`
	Paused                  bool                        `protobuf:"varint,24,opt,name=paused,proto3" json:"paused" xml:"paused"`
	WeakHashThresholdPct    int                         `protobuf:"varint,25,opt,name=weak_hash_threshold_pct,json=weakHashThresholdPct,proto3,casttype=int" json:"weakHashThresholdPct" xml:"weakHashThresholdPct"`
	MarkerName              string                      `protobuf:"bytes,26,opt,name=marker_name,json=markerName,proto3" json:"markerName" xml:"markerName"`
	CopyOwnershipFromParent bool                        `protobuf:"varint,27,opt,name=copy_ownership_from_parent,json=copyOwnershipFromParent,proto3" json:"copyOwnershipFromParent" xml:"copyOwnershipFromParent"`
	RawModTimeWindowS       int                         `protobuf:"varint,28,opt,name=mod_time_window_s,json=modTimeWindowS,proto3,casttype=int" json:"modTimeWindowS" xml:"modTimeWindowS"`
	MaxConcurrentWrites     int                         `protobuf:"varint,29,opt,name=max_concurrent_writes,json=maxConcurrentWrites,proto3,casttype=int" json:"maxConcurrentWrites" xml:"maxConcurrentWrites" default:"2"`
	DisableFsync            bool                        `protobuf:"varint,30,opt,name=disable_fsync,json=disableFsync,proto3" json:"disableFsync" xml:"disableFsync"`
	BlockPullOrder          BlockPullOrder              `protobuf:"varint,31,opt,name=block_pull_order,json=blockPullOrder,proto3,enum=config.BlockPullOrder" json:"blockPullOrder" xml:"blockPullOrder"`
	CopyRangeMethod         fs.CopyRangeMethod          `protobuf:"varint,32,opt,name=copy_range_method,json=copyRangeMethod,proto3,enum=fs.CopyRangeMethod" json:"copyRangeMethod" xml:"copyRangeMethod" default:"standard"`
	CaseSensitiveFS         bool                        `protobuf:"varint,33,opt,name=case_sensitive_fs,json=caseSensitiveFs,proto3" json:"caseSensitiveFS" xml:"caseSensitiveFS"`
	// Legacy deprecated
	DeprecatedReadOnly       bool    `protobuf:"varint,9000,opt,name=read_only,json=readOnly,proto3" json:"-" xml:"ro,attr,omitempty"`                       // Deprecated: Do not use.
	DeprecatedMinDiskFreePct float64 `protobuf:"fixed64,9001,opt,name=min_disk_free_pct,json=minDiskFreePct,proto3" json:"-" xml:"minDiskFreePct,omitempty"` // Deprecated: Do not use.
	DeprecatedPullers        int     `protobuf:"varint,9002,opt,name=pullers,proto3,casttype=int" json:"-" xml:"pullers,omitempty"`                          // Deprecated: Do not use.
}

func (m *FolderConfiguration) Reset()         { *m = FolderConfiguration{} }
func (m *FolderConfiguration) String() string { return proto.CompactTextString(m) }
func (*FolderConfiguration) ProtoMessage()    {}
func (*FolderConfiguration) Descriptor() ([]byte, []int) {
	return fileDescriptor_44a9785876ed3afa, []int{1}
}
func (m *FolderConfiguration) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FolderConfiguration.Unmarshal(m, b)
}
func (m *FolderConfiguration) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FolderConfiguration.Marshal(b, m, deterministic)
}
func (m *FolderConfiguration) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FolderConfiguration.Merge(m, src)
}
func (m *FolderConfiguration) XXX_Size() int {
	return xxx_messageInfo_FolderConfiguration.Size(m)
}
func (m *FolderConfiguration) XXX_DiscardUnknown() {
	xxx_messageInfo_FolderConfiguration.DiscardUnknown(m)
}

var xxx_messageInfo_FolderConfiguration proto.InternalMessageInfo

func init() {
	proto.RegisterType((*FolderDeviceConfiguration)(nil), "config.FolderDeviceConfiguration")
	proto.RegisterType((*FolderConfiguration)(nil), "config.FolderConfiguration")
}

func init() {
	proto.RegisterFile("lib/config/folderconfiguration.proto", fileDescriptor_44a9785876ed3afa)
}

var fileDescriptor_44a9785876ed3afa = []byte{
	// 1929 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x57, 0xcd, 0x6f, 0x1c, 0x49,
	0x15, 0x77, 0x3b, 0x5f, 0x76, 0xf9, 0x23, 0x76, 0xd9, 0x4e, 0x2a, 0xce, 0xee, 0x94, 0xb7, 0x19,
	0x96, 0xd9, 0xd5, 0xc6, 0x49, 0xbc, 0x88, 0x43, 0xb4, 0xbb, 0xc0, 0xd8, 0x6b, 0x88, 0x82, 0x37,
	0xa3, 0x76, 0x20, 0x22, 0x20, 0x35, 0x3d, 0xdd, 0x35, 0x33, 0x25, 0xf7, 0x17, 0x55, 0xed, 0xd8,
	0x93, 0x53, 0xb8, 0x20, 0x10, 0x7b, 0x40, 0xe6, 0xc0, 0x95, 0x03, 0x42, 0xb0, 0xff, 0x00, 0x12,
	0x12, 0xf7, 0xbd, 0x20, 0xcf, 0x11, 0x71, 0x28, 0x69, 0x9d, 0x0b, 0x9a, 0x1b, 0x7d, 0xcc, 0x09,
	0x55, 0x55, 0x77, 0x4f, 0xf7, 0xcc, 0x2c, 0x42, 0xda, 0x5b, 0xd7, 0xef, 0xf7, 0xab, 0xf7, 0x5e,
	0xbf, 0xaa, 0xf7, 0xaa, 0x0a, 0xd4, 0x7d, 0xda, 0xbe, 0xeb, 0x46, 0x61, 0x87, 0x76, 0xef, 0x76,
	0x22, 0xdf, 0x23, 0x4c, 0x0f, 0x8e, 0x99, 0x93, 0xd0, 0x28, 0xdc, 0x8e, 0x59, 0x94, 0x44, 0xf0,
	0xaa, 0x06, 0x37, 0x6f, 0x4f, 0xa8, 0x93, 0x7e, 0x4c, 0xb4, 0x68, 0x73, 0xa3, 0x44, 0x72, 0xfa,
	0x22, 0x87, 0x37, 0x4b, 0x70, 0x7c, 0xec, 0xfb, 0x11, 0xf3, 0x08, 0xcb, 0xb8, 0x46, 0x89, 0x7b,
	0x4e, 0x18, 0xa7, 0x51, 0x48, 0xc3, 0xee, 0x94, 0x08, 0x36, 0x71, 0x49, 0xd9, 0xf6, 0x23, 0xf7,
	0x68, 0xdc, 0x14, 0x94, 0x82, 0x0e, 0xbf, 0x2b, 0x03, 0xe2, 0x19, 0xf6, 0x46, 0x86, 0xb9, 0x51,
	0xdc, 0x67, 0x4e, 0xd8, 0x25, 0x01, 0x49, 0x7a, 0x91, 0x97, 0xb1, 0xf3, 0xe4, 0x34, 0xd1, 0x9f,
	0xe6, 0x7f, 0x66, 0xc1, 0xad, 0x7d, 0xf5, 0x3f, 0x7b, 0xe4, 0x39, 0x75, 0xc9, 0x6e, 0x39, 0x02,
	0xf8, 0x99, 0x01, 0xe6, 0x3d, 0x85, 0xdb, 0xd4, 0x43, 0xc6, 0x96, 0xd1, 0x58, 0x6c, 0x7e, 0x6a,
	0x7c, 0x2e, 0xf0, 0xcc, 0xbf, 0x04, 0xfe, 0x66, 0x97, 0x26, 0xbd, 0xe3, 0xf6, 0xb6, 0x1b, 0x05,
	0x77, 0x79, 0x3f, 0x74, 0x93, 0x1e, 0x0d, 0xbb, 0xa5, 0x2f, 0x19, 0x82, 0x72, 0xe2, 0x46, 0xfe,
	0xb6, 0xb6, 0xfe, 0x70, 0xef, 0x42, 0xe0, 0xb9, 0xfc, 0x7b, 0x28, 0xf0, 0x9c, 0x97, 0x7d, 0xa7,
	0x02, 0x2f, 0x9d, 0x06, 0xfe, 0x03, 0x93, 0x7a, 0xef, 0x39, 0x49, 0xc2, 0xcc, 0xe1, 0x79, 0xfd,
	0x5a, 0xf6, 0x9d, 0x9e, 0xd7, 0x0b, 0xdd, 0xaf, 0x06, 0x75, 0xe3, 0x6c, 0x50, 0x2f, 0x6c, 0x58,
	0x39, 0xe3, 0xc1, 0x3f, 0x19, 0x60, 0x89, 0x86, 0x09, 0x8b, 0xbc, 0x63, 0x97, 0x78, 0x76, 0xbb,
	0x8f, 0x66, 0x55, 0xc0, 0x2f, 0xbf, 0x52, 0xc0, 0x43, 0x81, 0x17, 0x47, 0x56, 0x9b, 0xfd, 0x54,
	0xe0, 0x9b, 0x3a, 0xd0, 0x12, 0x58, 0x84, 0xbc, 0x3a, 0x81, 0xca, 0x80, 0xad, 0x8a, 0x05, 0xf3,
	0xef, 0x35, 0xb0, 0xa6, 0x73, 0x5e, 0xcd, 0xf6, 0x47, 0x60, 0x36, 0xcb, 0xf2, 0x7c, 0x73, 0xfb,
	0x42, 0xe0, 0x59, 0xe5, 0x7d, 0x96, 0x7a, 0xff, 0x2b, 0x39, 0x67, 0x83, 0xfa, 0xec, 0xc3, 0x3d,
	0x6b, 0x96, 0x7a, 0xf0, 0x87, 0xe0, 0x8a, 0xef, 0xb4, 0x89, 0xaf, 0xfe, 0x7b, 0xbe, 0xf9, 0xed,
	0xa1, 0xc0, 0x1a, 0x48, 0x05, 0xde, 0x52, 0xf3, 0xd5, 0x48, 0x9b, 0xd8, 0x62, 0x84, 0x27, 0x0e,
	0x4b, 0x1e, 0x98, 0x1d, 0xc7, 0xe7, 0x44, 0x9a, 0x04, 0x23, 0xfa, 0xe5, 0xa0, 0x3e, 0x63, 0xe9,
	0xc9, 0xb0, 0x0b, 0xae, 0x77, 0xa8, 0x4f, 0x78, 0x9f, 0x27, 0x24, 0xb0, 0xe5, 0x2e, 0x43, 0x97,
	0xb6, 0x8c, 0xc6, 0xf2, 0x0e, 0xdc, 0xee, 0xf0, 0xed, 0xfd, 0x82, 0x7a, 0xd2, 0x8f, 0x49, 0xf3,
	0xdd, 0xa1, 0xc0, 0xcb, 0x9d, 0x0a, 0x96, 0x0a, 0xbc, 0xae, 0xbc, 0x57, 0x61, 0xd3, 0x1a, 0xd3,
	0xc1, 0x0f, 0xc0, 0xe5, 0xd8, 0x49, 0x7a, 0xe8, 0xb2, 0x0a, 0xbf, 0x31, 0x14, 0x58, 0x8d, 0x53,
	0x81, 0xaf, 0xab, 0xf9, 0x72, 0x50, 0xfc, 0xff, 0x7c, 0x31, 0xb2, 0x94, 0x0a, 0xb6, 0xc0, 0x65,
	0x15, 0xdb, 0x95, 0x2c, 0x36, 0x5d, 0x32, 0xdb, 0x3a, 0xd1, 0x2a, 0x36, 0x65, 0x31, 0xd1, 0x11,
	0x69, 0x8b, 0x72, 0x30, 0xb2, 0x58, 0x8c, 0x2c, 0xa5, 0x82, 0x3f, 0x05, 0xd7, 0xf4, 0xe6, 0xe2,
	0xe8, 0xea, 0xd6, 0xa5, 0xc6, 0xc2, 0xce, 0x5b, 0x55, 0xa3, 0x53, 0x2a, 0xa6, 0x89, 0xe5, 0x5e,
	0x1b, 0x0a, 0x9c, 0xcf, 0x4c, 0x05, 0x5e, 0x54, 0xae, 0xf4, 0xd8, 0xb4, 0x72, 0x02, 0xfe, 0xce,
	0x00, 0xab, 0x8c, 0x70, 0xd7, 0x09, 0x6d, 0x1a, 0x26, 0x84, 0x3d, 0x77, 0x7c, 0x9b, 0xa3, 0x6b,
	0x5b, 0x46, 0xe3, 0x4a, 0xb3, 0x3b, 0x14, 0xf8, 0xba, 0x26, 0x1f, 0x66, 0xdc, 0x61, 0x2a, 0xf0,
	0x3b, 0xca, 0xd2, 0x18, 0x9e, 0x2d, 0xa7, 0x47, 0x3a, 0xce, 0xb1, 0x9f, 0x3c, 0x30, 0xdf, 0xff,
	0xd6, 0xbd, 0x7b, 0xe6, 0x6b, 0x81, 0x2f, 0xd1, 0x30, 0x19, 0x9e, 0xd7, 0xd7, 0xa7, 0xc9, 0x5f,
	0x9f, 0xd7, 0x2f, 0x4b, 0x9d, 0x35, 0xee, 0x04, 0xfe, 0xcd, 0x00, 0xb0, 0xc3, 0xed, 0x13, 0x27,
	0x71, 0x7b, 0x84, 0xd9, 0x24, 0x74, 0xda, 0x3e, 0xf1, 0xd0, 0xdc, 0x96, 0xd1, 0x98, 0x6b, 0xfe,
	0xc6, 0xb8, 0x10, 0x78, 0x65, 0xff, 0xf0, 0xa9, 0x66, 0x3f, 0xd6, 0xe4, 0x50, 0xe0, 0x95, 0x0e,
	0xaf, 0x62, 0xa9, 0xc0, 0xef, 0xea, 0x35, 0x1f, 0x23, 0xc6, 0xa3, 0x4d, 0xd8, 0xb1, 0xda, 0x7b,
	0x1b, 0x53, 0x85, 0x32, 0x4e, 0xa9, 0x38, 0x1b, 0xd4, 0x27, 0xdc, 0x5a, 0x13, 0x4e, 0xe1, 0x5f,
	0xab, 0xc1, 0x7b, 0xc4, 0x77, 0xfa, 0x36, 0x47, 0xf3, 0x2a, 0xa7, 0xbf, 0x96, 0xc1, 0x5f, 0x2f,
	0xac, 0xec, 0x49, 0xf2, 0x50, 0xe6, 0xb9, 0x30, 0xa3, 0xa1, 0x54, 0xe0, 0x6f, 0x54, 0x43, 0xd7,
	0xf8, 0x78, 0xe4, 0xf7, 0x2b, 0x59, 0x9e, 0x26, 0x7e, 0x7d, 0x5e, 0x9f, 0xbd, 0x7f, 0xef, 0x6c,
	0x50, 0x1f, 0xf7, 0x6a, 0x8d, 0xfb, 0x84, 0x3f, 0x03, 0x8b, 0xb4, 0x1b, 0x46, 0x8c, 0xd8, 0x31,
	0x61, 0x01, 0x47, 0x40, 0xe5, 0xfb, 0xc3, 0xa1, 0xc0, 0x0b, 0x1a, 0x6f, 0x49, 0x38, 0x15, 0xf8,
	0x86, 0xee, 0x03, 0x23, 0xac, 0xd8, 0xbe, 0x2b, 0xe3, 0xa0, 0x55, 0x9e, 0x0a, 0x7f, 0x61, 0x80,
	0x65, 0xe7, 0x38, 0x89, 0xec, 0x30, 0x62, 0x81, 0xe3, 0xd3, 0x17, 0x04, 0x2d, 0x28, 0x27, 0xcf,
	0x86, 0x02, 0x2f, 0x49, 0xe6, 0x93, 0x9c, 0x28, 0x32, 0x50, 0x41, 0xbf, 0x6c, 0xe5, 0xe0, 0xa4,
	0x2a, 0x5f, 0x36, 0xab, 0x6a, 0x17, 0x3e, 0x03, 0x4b, 0x01, 0x0d, 0x6d, 0x8f, 0xf2, 0x23, 0xbb,
	0xc3, 0x08, 0x41, 0x8b, 0x5b, 0x46, 0x63, 0x61, 0x67, 0x31, 0x2f, 0xab, 0x43, 0xfa, 0x82, 0x34,
	0x1b, 0x59, 0x05, 0x2d, 0x04, 0x34, 0xdc, 0xa3, 0xfc, 0x68, 0x9f, 0x11, 0x19, 0xd1, 0xaa, 0x8a,
	0xa8, 0x84, 0x99, 0x56, 0x59, 0x01, 0xbb, 0x00, 0x8c, 0xce, 0x51, 0xb4, 0xa4, 0x0c, 0xe3, 0xdc,
	0xf0, 0x8f, 0x0a, 0xa6, 0x5a, 0xad, 0x6f, 0x67, 0xbe, 0x4a, 0x53, 0x53, 0x81, 0x57, 0x94, 0xab,
	0x11, 0x64, 0x5a, 0x25, 0x1e, 0x7e, 0x08, 0xae, 0xb9, 0x51, 0x4c, 0x09, 0xe3, 0x68, 0x59, 0x6d,
	0xac, 0xaf, 0xc9, 0x72, 0xcf, 0xa0, 0xa2, 0x53, 0x67, 0xe3, 0x7c, 0x8b, 0x58, 0xb9, 0x00, 0xfe,
	0xc3, 0x00, 0x37, 0xe4, 0x09, 0x4e, 0x98, 0x1d, 0x38, 0xa7, 0x76, 0x4c, 0x42, 0x8f, 0x86, 0x5d,
	0xfb, 0x88, 0xb6, 0xd1, 0x75, 0x65, 0xee, 0xf7, 0x72, 0x9f, 0xae, 0xb5, 0x94, 0xe4, 0xc0, 0x39,
	0x6d, 0x69, 0xc1, 0x23, 0xda, 0x1c, 0x0a, 0xbc, 0x16, 0x4f, 0xc2, 0xa9, 0xc0, 0xb7, 0x74, 0x7b,
	0x9c, 0xe4, 0x4a, 0x3b, 0x74, 0xea, 0xd4, 0xe9, 0xf0, 0xd9, 0xa0, 0x3e, 0xcd, 0xbf, 0x35, 0x45,
	0xdb, 0x96, 0xe9, 0xe8, 0x39, 0xbc, 0x27, 0xd3, 0xb1, 0x32, 0x4a, 0x47, 0x06, 0x15, 0xe9, 0xc8,
	0xc6, 0xa3, 0x74, 0x64, 0x00, 0xfc, 0x2e, 0xb8, 0xa2, 0xee, 0x32, 0x68, 0x55, 0xb5, 0xed, 0xd5,
	0x7c, 0xc5, 0xa4, 0xff, 0xc7, 0x92, 0x68, 0x22, 0x79, 0x8c, 0x29, 0x4d, 0x2a, 0xf0, 0x82, 0xb2,
	0xa6, 0x46, 0xa6, 0xa5, 0x51, 0xf8, 0x08, 0x2c, 0x65, 0xb5, 0xe3, 0x11, 0x9f, 0x24, 0x04, 0x41,
	0xb5, 0xaf, 0xdf, 0x56, 0x27, 0xb7, 0x22, 0xf6, 0x14, 0x9e, 0x0a, 0x0c, 0x4b, 0xd5, 0xa3, 0x41,
	0xd3, 0xaa, 0x68, 0xe0, 0x29, 0x40, 0xaa, 0x25, 0xc7, 0x2c, 0xea, 0x32, 0xc2, 0x79, 0xb9, 0x37,
	0xaf, 0xa9, 0xff, 0x93, 0xc7, 0xea, 0x86, 0xd4, 0xb4, 0x32, 0x49, 0xb9, 0x43, 0xdf, 0x56, 0x0e,
	0xa6, 0xb2, 0xc5, 0xbf, 0x4f, 0x9f, 0x0c, 0x0f, 0xc1, 0x72, 0xb6, 0x2f, 0x62, 0xe7, 0x98, 0x13,
	0x9b, 0xa3, 0x75, 0xe5, 0xef, 0x8e, 0xfc, 0x0f, 0xcd, 0xb4, 0x24, 0x71, 0x58, 0xfc, 0x47, 0x19,
	0x2c, 0xac, 0x57, 0xa4, 0x90, 0x80, 0x25, 0xb9, 0xcb, 0x64, 0x52, 0x7d, 0xea, 0x26, 0x1c, 0x6d,
	0x28, 0x9b, 0xdf, 0x91, 0x36, 0x03, 0xe7, 0x74, 0x37, 0xc7, 0x53, 0x81, 0xb1, 0x2e, 0xb0, 0x12,
	0x58, 0x2a, 0xf6, 0x3b, 0xf7, 0x73, 0x07, 0xb2, 0xa9, 0xdd, 0xb9, 0x6f, 0x55, 0x66, 0x43, 0x0f,
	0xac, 0x7b, 0x94, 0xcb, 0x26, 0x6c, 0xf3, 0xd8, 0x61, 0x9c, 0xd8, 0xea, 0x68, 0x47, 0x37, 0xd4,
	0x4a, 0xec, 0x0c, 0x05, 0x86, 0x19, 0x7f, 0xa8, 0x68, 0x75, 0x69, 0x48, 0x05, 0x46, 0xfa, 0x68,
	0x9c, 0xa0, 0x4c, 0x6b, 0x8a, 0xbe, 0xec, 0x25, 0x21, 0x41, 0x6c, 0xd3, 0xd0, 0x23, 0xa7, 0x84,
	0xa3, 0x9b, 0x13, 0x5e, 0x9e, 0x90, 0x20, 0x7e, 0xa8, 0xd9, 0x71, 0x2f, 0x25, 0x6a, 0xe4, 0xa5,
	0x04, 0xc2, 0x1d, 0x70, 0x55, 0x2d, 0x80, 0x87, 0x90, 0xb2, 0xbb, 0x39, 0x14, 0x38, 0x43, 0x8a,
	0xc3, 0x5c, 0x0f, 0x4d, 0x2b, 0xc3, 0x61, 0x02, 0x6e, 0x9e, 0x10, 0xe7, 0xc8, 0x96, 0xbb, 0xda,
	0x4e, 0x7a, 0x8c, 0xf0, 0x5e, 0xe4, 0x7b, 0x76, 0xec, 0x26, 0xe8, 0x96, 0x4a, 0xb8, 0xec, 0xe4,
	0xeb, 0x52, 0xf2, 0x7d, 0x87, 0xf7, 0x9e, 0xe4, 0x82, 0x96, 0x9b, 0xa4, 0x02, 0x6f, 0x2a, 0x93,
	0xd3, 0xc8, 0x62, 0x51, 0xa7, 0x4e, 0x85, 0xbb, 0x60, 0x21, 0x70, 0xd8, 0x11, 0x61, 0x76, 0xe8,
	0x04, 0x04, 0x6d, 0xaa, 0x6b, 0x93, 0x29, 0xdb, 0x99, 0x86, 0x3f, 0x71, 0x02, 0x52, 0xb4, 0xb3,
	0x11, 0x64, 0x5a, 0x25, 0x1e, 0xf6, 0xc1, 0xa6, 0x7c, 0x24, 0xd8, 0xd1, 0x49, 0x48, 0x18, 0xef,
	0xd1, 0xd8, 0xee, 0xb0, 0x28, 0xb0, 0x63, 0x87, 0x91, 0x30, 0x41, 0xb7, 0x55, 0x0a, 0x3e, 0x18,
	0x0a, 0x7c, 0x53, 0xaa, 0x1e, 0xe7, 0xa2, 0x7d, 0x16, 0x05, 0x2d, 0x25, 0x49, 0x05, 0x7e, 0x33,
	0xef, 0x78, 0xd3, 0x78, 0xd3, 0xfa, 0xb2, 0x99, 0xf0, 0x97, 0x06, 0x58, 0x0d, 0x22, 0xcf, 0x4e,
	0x68, 0x40, 0xec, 0x13, 0x1a, 0x7a, 0xd1, 0x89, 0xcd, 0xd1, 0x1b, 0x2a, 0x61, 0x3f, 0xb9, 0x10,
	0x78, 0xd5, 0x72, 0x4e, 0x0e, 0x22, 0xef, 0x09, 0x0d, 0xc8, 0x53, 0xc5, 0xca, 0xe3, 0x7a, 0x39,
	0xa8, 0x20, 0xc5, 0xe5, 0xb2, 0x0a, 0xe7, 0x99, 0x3b, 0x1b, 0xd4, 0x27, 0xad, 0x58, 0x63, 0x36,
	0xe0, 0x4b, 0x03, 0x6c, 0x64, 0x65, 0xe2, 0x1e, 0x33, 0x19, 0x9b, 0x7d, 0xc2, 0x68, 0x42, 0x38,
	0x7a, 0x53, 0x05, 0xf3, 0x03, 0xd9, 0x7a, 0xf5, 0x86, 0xcf, 0xf8, 0xa7, 0x8a, 0x4e, 0x05, 0xfe,
	0x7a, 0xa9, 0x6a, 0x2a, 0x5c, 0xa9, 0x78, 0x76, 0x4a, 0xb5, 0x63, 0xec, 0x58, 0xd3, 0x2c, 0xc9,
	0x26, 0x96, 0xef, 0xed, 0x8e, 0x7c, 0x91, 0xa0, 0xda, 0xa8, 0x89, 0x65, 0xc4, 0xbe, 0xc4, 0x8b,
	0xe2, 0x2f, 0x83, 0xa6, 0x55, 0xd1, 0x40, 0x1f, 0xac, 0xa8, 0x97, 0xa2, 0x2d, 0x7b, 0x81, 0xad,
	0xfb, 0x2b, 0x56, 0xfd, 0xf5, 0x46, 0xde, 0x5f, 0x9b, 0x92, 0x1f, 0x35, 0x59, 0x75, 0x6d, 0x6f,
	0x57, 0xb0, 0x22, 0xb3, 0x55, 0xd8, 0xb4, 0xc6, 0x74, 0xf0, 0x53, 0x03, 0xac, 0xaa, 0x2d, 0xa4,
	0x1e, 0x9a, 0xb6, 0x7e, 0x69, 0xa2, 0x2d, 0xe5, 0x6f, 0x4d, 0x3e, 0x11, 0x76, 0xa3, 0xb8, 0x6f,
	0x49, 0xee, 0x40, 0x51, 0xcd, 0x47, 0xf2, 0xd6, 0xe5, 0x56, 0xc1, 0x54, 0xe0, 0x46, 0xb1, 0x8d,
	0x4a, 0x78, 0x29, 0x8d, 0x3c, 0x71, 0x42, 0xcf, 0x61, 0x9e, 0xf9, 0xfa, 0xbc, 0x3e, 0x97, 0x0f,
	0xac, 0x71, 0x43, 0xf0, 0x8f, 0x32, 0x1c, 0x47, 0x36, 0x50, 0x12, 0x72, 0x9a, 0xd0, 0xe7, 0x32,
	0xa3, 0xe8, 0x2d, 0x95, 0xce, 0x53, 0x79, 0x05, 0xdc, 0x75, 0x38, 0x39, 0xcc, 0xb9, 0x7d, 0x75,
	0x05, 0x74, 0xab, 0x50, 0x2a, 0xf0, 0x86, 0x0e, 0xa6, 0x8a, 0xcb, 0xeb, 0xce, 0x84, 0x76, 0x12,
	0x92, 0x37, 0xbe, 0x31, 0x27, 0xd6, 0x98, 0x86, 0xc3, 0x23, 0x30, 0xcf, 0x88, 0xe3, 0xd9, 0x51,
	0xe8, 0xf7, 0xd1, 0x9f, 0xf7, 0x55, 0x78, 0x07, 0x17, 0x02, 0xc3, 0x3d, 0x12, 0x33, 0xe2, 0x3a,
	0x09, 0xf1, 0x2c, 0xe2, 0x78, 0x8f, 0x43, 0xbf, 0x3f, 0x14, 0xd8, 0xb8, 0x53, 0xbc, 0x3b, 0x59,
	0xa4, 0xee, 0x58, 0xef, 0x45, 0x01, 0x95, 0x5d, 0x30, 0xe9, 0xab, 0x77, 0xe7, 0x04, 0x8a, 0x0c,
	0x6b, 0x8e, 0x65, 0x06, 0xe0, 0xcf, 0xc1, 0x6a, 0xe5, 0xe2, 0xa5, 0x3a, 0xd3, 0x5f, 0xa4, 0x53,
	0xa3, 0xf9, 0xf1, 0x85, 0xc0, 0x68, 0xe4, 0xf4, 0x60, 0x74, 0xa7, 0x6a, 0xb9, 0x49, 0xee, 0xba,
	0x36, 0x7e, 0xfb, 0x6a, 0xb9, 0x49, 0x29, 0x02, 0x64, 0x58, 0xcb, 0x55, 0x12, 0xfe, 0x18, 0x5c,
	0xd3, 0x27, 0x11, 0x47, 0x9f, 0xed, 0xab, 0x2a, 0xfa, 0x48, 0x96, 0xf4, 0xc8, 0x91, 0xbe, 0x61,
	0xf0, 0xea, 0xcf, 0x65, 0x53, 0x4a, 0xa6, 0xb3, 0xd2, 0x41, 0x86, 0x95, 0xdb, 0x6b, 0x7e, 0xef,
	0xf3, 0x2f, 0x6a, 0x33, 0x83, 0x2f, 0x6a, 0x33, 0xff, 0xbe, 0xa8, 0xcd, 0xfc, 0xf6, 0x55, 0x6d,
	0xe6, 0x0f, 0xaf, 0x6a, 0xc6, 0xe0, 0x55, 0x6d, 0xe6, 0x9f, 0xaf, 0x6a, 0x33, 0xcf, 0xde, 0xf9,
	0x3f, 0x5e, 0xf9, 0xba, 0x08, 0xda, 0x57, 0xd5, 0x6b, 0xff, 0xfd, 0xff, 0x06, 0x00, 0x00, 0xff,
	0xff, 0x3d, 0xce, 0xbc, 0x07, 0x0b, 0x12, 0x00, 0x00,
}

func (m *FolderDeviceConfiguration) ProtoSize() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.DeviceID.ProtoSize()
	n += 1 + l + sovFolderconfiguration(uint64(l))
	l = m.IntroducedBy.ProtoSize()
	n += 1 + l + sovFolderconfiguration(uint64(l))
	return n
}

func (m *FolderConfiguration) ProtoSize() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ID)
	if l > 0 {
		n += 1 + l + sovFolderconfiguration(uint64(l))
	}
	l = len(m.Label)
	if l > 0 {
		n += 1 + l + sovFolderconfiguration(uint64(l))
	}
	if m.FilesystemType != 0 {
		n += 1 + sovFolderconfiguration(uint64(m.FilesystemType))
	}
	l = len(m.Path)
	if l > 0 {
		n += 1 + l + sovFolderconfiguration(uint64(l))
	}
	if m.Type != 0 {
		n += 1 + sovFolderconfiguration(uint64(m.Type))
	}
	if len(m.Devices) > 0 {
		for _, e := range m.Devices {
			l = e.ProtoSize()
			n += 1 + l + sovFolderconfiguration(uint64(l))
		}
	}
	if m.RescanIntervalS != 0 {
		n += 1 + sovFolderconfiguration(uint64(m.RescanIntervalS))
	}
	if m.FSWatcherEnabled {
		n += 2
	}
	if m.FSWatcherDelayS != 0 {
		n += 1 + sovFolderconfiguration(uint64(m.FSWatcherDelayS))
	}
	if m.IgnorePerms {
		n += 2
	}
	if m.AutoNormalize {
		n += 2
	}
	l = m.MinDiskFree.ProtoSize()
	n += 1 + l + sovFolderconfiguration(uint64(l))
	l = m.Versioning.ProtoSize()
	n += 1 + l + sovFolderconfiguration(uint64(l))
	if m.Copiers != 0 {
		n += 1 + sovFolderconfiguration(uint64(m.Copiers))
	}
	if m.PullerMaxPendingKiB != 0 {
		n += 1 + sovFolderconfiguration(uint64(m.PullerMaxPendingKiB))
	}
	if m.Hashers != 0 {
		n += 2 + sovFolderconfiguration(uint64(m.Hashers))
	}
	if m.Order != 0 {
		n += 2 + sovFolderconfiguration(uint64(m.Order))
	}
	if m.IgnoreDelete {
		n += 3
	}
	if m.ScanProgressIntervalS != 0 {
		n += 2 + sovFolderconfiguration(uint64(m.ScanProgressIntervalS))
	}
	if m.PullerPauseS != 0 {
		n += 2 + sovFolderconfiguration(uint64(m.PullerPauseS))
	}
	if m.MaxConflicts != 0 {
		n += 2 + sovFolderconfiguration(uint64(m.MaxConflicts))
	}
	if m.DisableSparseFiles {
		n += 3
	}
	if m.DisableTempIndexes {
		n += 3
	}
	if m.Paused {
		n += 3
	}
	if m.WeakHashThresholdPct != 0 {
		n += 2 + sovFolderconfiguration(uint64(m.WeakHashThresholdPct))
	}
	l = len(m.MarkerName)
	if l > 0 {
		n += 2 + l + sovFolderconfiguration(uint64(l))
	}
	if m.CopyOwnershipFromParent {
		n += 3
	}
	if m.RawModTimeWindowS != 0 {
		n += 2 + sovFolderconfiguration(uint64(m.RawModTimeWindowS))
	}
	if m.MaxConcurrentWrites != 0 {
		n += 2 + sovFolderconfiguration(uint64(m.MaxConcurrentWrites))
	}
	if m.DisableFsync {
		n += 3
	}
	if m.BlockPullOrder != 0 {
		n += 2 + sovFolderconfiguration(uint64(m.BlockPullOrder))
	}
	if m.CopyRangeMethod != 0 {
		n += 2 + sovFolderconfiguration(uint64(m.CopyRangeMethod))
	}
	if m.CaseSensitiveFS {
		n += 3
	}
	if m.DeprecatedReadOnly {
		n += 4
	}
	if m.DeprecatedMinDiskFreePct != 0 {
		n += 11
	}
	if m.DeprecatedPullers != 0 {
		n += 3 + sovFolderconfiguration(uint64(m.DeprecatedPullers))
	}
	return n
}

func sovFolderconfiguration(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozFolderconfiguration(x uint64) (n int) {
	return sovFolderconfiguration(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
