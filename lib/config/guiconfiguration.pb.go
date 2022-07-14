// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: lib/config/guiconfiguration.proto

package config

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/syncthing/syncthing/proto/ext"
	io "io"
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

type GUIConfiguration struct {
	Enabled                   bool                 `protobuf:"varint,1,opt,name=enabled,proto3" json:"enabled" xml:"enabled,attr" default:"true"`
	RawAddress                string               `protobuf:"bytes,2,opt,name=address,proto3" json:"address" xml:"address" default:"127.0.0.1:8384"`
	RawUnixSocketPermissions  string               `protobuf:"bytes,3,opt,name=unix_socket_permissions,json=unixSocketPermissions,proto3" json:"unixSocketPermissions" xml:"unixSocketPermissions,omitempty"`
	User                      string               `protobuf:"bytes,4,opt,name=user,proto3" json:"user" xml:"user,omitempty"`
	Password                  string               `protobuf:"bytes,5,opt,name=password,proto3" json:"password" xml:"password,omitempty"`
	AuthMode                  AuthMode             `protobuf:"varint,6,opt,name=auth_mode,json=authMode,proto3,enum=config.AuthMode" json:"authMode" xml:"authMode,omitempty"`
	RawUseTLS                 bool                 `protobuf:"varint,7,opt,name=use_tls,json=useTls,proto3" json:"useTLS" xml:"tls,attr"`
	APIKey                    string               `protobuf:"bytes,8,opt,name=api_key,json=apiKey,proto3" json:"apiKey" xml:"apikey,omitempty"`
	InsecureAdminAccess       bool                 `protobuf:"varint,9,opt,name=insecure_admin_access,json=insecureAdminAccess,proto3" json:"insecureAdminAccess" xml:"insecureAdminAccess,omitempty"`
	Theme                     string               `protobuf:"bytes,10,opt,name=theme,proto3" json:"theme" xml:"theme" default:"default"`
	Debugging                 bool                 `protobuf:"varint,11,opt,name=debugging,proto3" json:"debugging" xml:"debugging,attr"`
	InsecureSkipHostCheck     bool                 `protobuf:"varint,12,opt,name=insecure_skip_host_check,json=insecureSkipHostCheck,proto3" json:"insecureSkipHostcheck" xml:"insecureSkipHostcheck,omitempty"`
	InsecureAllowFrameLoading bool                 `protobuf:"varint,13,opt,name=insecure_allow_frame_loading,json=insecureAllowFrameLoading,proto3" json:"insecureAllowFrameLoading" xml:"insecureAllowFrameLoading,omitempty"`
	WebauthnRpId              string               `protobuf:"bytes,14,opt,name=webauthn_rp_id,json=webauthnRpId,proto3" json:"webauthnRpId" xml:"webauthnRpId"`
	WebauthnOrigin            string               `protobuf:"bytes,15,opt,name=webauthn_origin,json=webauthnOrigin,proto3" json:"webauthnOrigin" xml:"webauthnOrigin"`
	WebauthnCredentials       []WebauthnCredential `protobuf:"bytes,16,rep,name=webauthn_credentials,json=webauthnCredentials,proto3" json:"webauthnCredentials" xml:"webauthnCredential"`
}

func (m *GUIConfiguration) Reset()         { *m = GUIConfiguration{} }
func (m *GUIConfiguration) String() string { return proto.CompactTextString(m) }
func (*GUIConfiguration) ProtoMessage()    {}
func (*GUIConfiguration) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9586d611855d64, []int{0}
}
func (m *GUIConfiguration) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GUIConfiguration) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GUIConfiguration.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GUIConfiguration) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GUIConfiguration.Merge(m, src)
}
func (m *GUIConfiguration) XXX_Size() int {
	return m.ProtoSize()
}
func (m *GUIConfiguration) XXX_DiscardUnknown() {
	xxx_messageInfo_GUIConfiguration.DiscardUnknown(m)
}

var xxx_messageInfo_GUIConfiguration proto.InternalMessageInfo

type WebauthnCredential struct {
	ID            string `protobuf:"bytes,1,opt,name=id,proto3" json:"id" xml:"id,attr"`
	Nickname      string `protobuf:"bytes,2,opt,name=nickname,proto3" json:"nickname" xml:"nickname,attr"`
	PublicKeyCose string `protobuf:"bytes,3,opt,name=public_key_cose,json=publicKeyCose,proto3" json:"publicKeyCose" xml:"publicKeyCose,attr"`
	SignCount     uint32 `protobuf:"varint,4,opt,name=sign_count,json=signCount,proto3" json:"signCount" xml:"signCount,attr"`
}

func (m *WebauthnCredential) Reset()         { *m = WebauthnCredential{} }
func (m *WebauthnCredential) String() string { return proto.CompactTextString(m) }
func (*WebauthnCredential) ProtoMessage()    {}
func (*WebauthnCredential) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9586d611855d64, []int{1}
}
func (m *WebauthnCredential) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *WebauthnCredential) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_WebauthnCredential.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *WebauthnCredential) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WebauthnCredential.Merge(m, src)
}
func (m *WebauthnCredential) XXX_Size() int {
	return m.ProtoSize()
}
func (m *WebauthnCredential) XXX_DiscardUnknown() {
	xxx_messageInfo_WebauthnCredential.DiscardUnknown(m)
}

var xxx_messageInfo_WebauthnCredential proto.InternalMessageInfo

func init() {
	proto.RegisterType((*GUIConfiguration)(nil), "config.GUIConfiguration")
	proto.RegisterType((*WebauthnCredential)(nil), "config.WebauthnCredential")
}

func init() { proto.RegisterFile("lib/config/guiconfiguration.proto", fileDescriptor_2a9586d611855d64) }

var fileDescriptor_2a9586d611855d64 = []byte{
	// 1110 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x56, 0xcf, 0x6e, 0xdb, 0xc6,
	0x13, 0x16, 0x95, 0xc4, 0xb2, 0x36, 0xb6, 0x6c, 0x6c, 0x92, 0xdf, 0x8f, 0x31, 0x1a, 0xad, 0xa3,
	0xa8, 0x85, 0x03, 0x04, 0x72, 0xe2, 0xa4, 0x4d, 0x60, 0x14, 0x01, 0x6c, 0x05, 0x69, 0x0c, 0x3b,
	0xa8, 0xb1, 0xae, 0x5b, 0x20, 0x28, 0x40, 0x50, 0xe4, 0x5a, 0x5a, 0x88, 0x22, 0x59, 0x2e, 0x09,
	0x59, 0x87, 0xf6, 0x19, 0x0a, 0xf7, 0xda, 0x02, 0x7d, 0x86, 0x5e, 0x7a, 0xe9, 0x03, 0xf8, 0x26,
	0x9d, 0x8a, 0x9e, 0x16, 0x88, 0x7c, 0x63, 0x6f, 0x3c, 0xe6, 0x54, 0xec, 0xf2, 0x8f, 0x44, 0x4b,
	0x69, 0x7a, 0xdb, 0xfd, 0xe6, 0x9b, 0xf9, 0x66, 0x97, 0x33, 0xb3, 0x04, 0x77, 0x2d, 0xda, 0xda,
	0x34, 0x1c, 0xfb, 0x84, 0xb6, 0x37, 0xdb, 0x01, 0x8d, 0x57, 0x81, 0xa7, 0xfb, 0xd4, 0xb1, 0x1b,
	0xae, 0xe7, 0xf8, 0x0e, 0x5c, 0x88, 0xc1, 0xb5, 0xdb, 0x53, 0x54, 0x3d, 0xf0, 0x3b, 0x3d, 0xc7,
	0x24, 0x31, 0x65, 0xad, 0x4c, 0x4e, 0xfd, 0x78, 0x59, 0xfb, 0x63, 0x05, 0xac, 0x7e, 0x71, 0xbc,
	0xd7, 0x9c, 0x0e, 0x04, 0x5b, 0xa0, 0x44, 0x6c, 0xbd, 0x65, 0x11, 0x53, 0x55, 0xd6, 0x95, 0x8d,
	0xc5, 0xdd, 0x57, 0x21, 0x47, 0x29, 0x14, 0x71, 0x74, 0xf7, 0xb4, 0x67, 0x6d, 0xd7, 0x92, 0xfd,
	0x03, 0xdd, 0xf7, 0xbd, 0xda, 0xba, 0x49, 0x4e, 0xf4, 0xc0, 0xf2, 0xb7, 0x6b, 0xbe, 0x17, 0x90,
	0x5a, 0x38, 0xac, 0x2f, 0x4d, 0xdb, 0xdf, 0x0d, 0xeb, 0x57, 0x85, 0x01, 0xa7, 0x51, 0xe0, 0xf7,
	0xa0, 0xa4, 0x9b, 0xa6, 0x47, 0x18, 0x53, 0x8b, 0xeb, 0xca, 0x46, 0x79, 0xd7, 0x18, 0x73, 0x04,
	0xb0, 0xde, 0xdf, 0x89, 0x51, 0xa1, 0x98, 0x10, 0x22, 0x8e, 0x3e, 0x91, 0x8a, 0xc9, 0x7e, 0x4a,
	0xec, 0xd1, 0xd6, 0xd3, 0xc6, 0xc3, 0xc6, 0xc3, 0xc6, 0xa3, 0xed, 0x67, 0x8f, 0x9f, 0x3d, 0xa9,
	0xbd, 0x1b, 0xd6, 0x2b, 0x79, 0xe8, 0x6c, 0x54, 0x9f, 0x0a, 0x8a, 0xd3, 0x90, 0xf0, 0x4f, 0x05,
	0xfc, 0x3f, 0xb0, 0xe9, 0xa9, 0xc6, 0x1c, 0xa3, 0x4b, 0x7c, 0xcd, 0x25, 0x5e, 0x8f, 0x32, 0x46,
	0x1d, 0x9b, 0xa9, 0x57, 0x64, 0x3e, 0xbf, 0x28, 0x63, 0x8e, 0x54, 0xac, 0xf7, 0x8f, 0x6d, 0x7a,
	0x7a, 0x24, 0x59, 0x87, 0x13, 0x52, 0xc8, 0xd1, 0xad, 0x60, 0x9e, 0x21, 0xe2, 0xe8, 0x63, 0x99,
	0xec, 0x5c, 0xeb, 0x03, 0xa7, 0x47, 0x7d, 0xd2, 0x73, 0xfd, 0x81, 0xb8, 0x22, 0xf4, 0x01, 0xce,
	0xd9, 0xa8, 0xfe, 0xde, 0x04, 0xf0, 0x7c, 0x79, 0xf8, 0x12, 0x5c, 0x0d, 0x18, 0xf1, 0xd4, 0xab,
	0xf2, 0x10, 0x5b, 0x21, 0x47, 0x72, 0x1f, 0x71, 0x74, 0x33, 0x4e, 0x8b, 0x11, 0x2f, 0x9f, 0x45,
	0x25, 0x0f, 0x61, 0xc9, 0x87, 0x6f, 0xc0, 0xa2, 0xab, 0x33, 0xd6, 0x77, 0x3c, 0x53, 0xbd, 0x26,
	0x63, 0x3d, 0x0f, 0x39, 0xca, 0xb0, 0x88, 0x23, 0x55, 0xc6, 0x4b, 0x81, 0x7c, 0x4c, 0x38, 0x0b,
	0xe3, 0xcc, 0x17, 0xf6, 0x40, 0x59, 0x54, 0xa4, 0x26, 0x4a, 0x52, 0x5d, 0x58, 0x57, 0x36, 0x2a,
	0x5b, 0xab, 0x8d, 0xb8, 0x54, 0x1b, 0x3b, 0x81, 0xdf, 0x79, 0xed, 0x98, 0x24, 0x96, 0xd3, 0x93,
	0x5d, 0x26, 0x97, 0x02, 0x97, 0xe4, 0x66, 0x61, 0x9c, 0xf9, 0x42, 0x02, 0x4a, 0x01, 0x23, 0x9a,
	0x6f, 0x31, 0xb5, 0x24, 0xcb, 0xf9, 0x60, 0xcc, 0x51, 0x59, 0x5c, 0x2c, 0x23, 0x5f, 0x1d, 0x1c,
	0x85, 0x1c, 0x2d, 0x04, 0x72, 0x15, 0x71, 0x54, 0x91, 0x2a, 0xbe, 0xc5, 0xe2, 0xb2, 0x0e, 0x87,
	0xf5, 0xc5, 0x74, 0x13, 0x0d, 0xeb, 0x09, 0xef, 0x6c, 0x54, 0x9f, 0xb8, 0x63, 0x09, 0x5a, 0x4c,
	0xc8, 0xe8, 0x2e, 0xd5, 0xba, 0x64, 0xa0, 0x2e, 0xca, 0x0b, 0x13, 0x32, 0x0b, 0x3b, 0x87, 0x7b,
	0xfb, 0x64, 0x20, 0x34, 0x74, 0x97, 0xee, 0x93, 0x41, 0xc4, 0xd1, 0xff, 0xe2, 0x93, 0xb8, 0xb4,
	0x4b, 0x06, 0xf9, 0x73, 0xac, 0x5e, 0x06, 0xcf, 0x46, 0xf5, 0x24, 0x02, 0x4e, 0xfc, 0xe1, 0x4f,
	0x0a, 0xb8, 0x45, 0x6d, 0x46, 0x8c, 0xc0, 0x23, 0x9a, 0x6e, 0xf6, 0xa8, 0xad, 0xe9, 0x86, 0x21,
	0xfa, 0xa8, 0x2c, 0x0f, 0xa7, 0x85, 0x1c, 0xdd, 0x48, 0x09, 0x3b, 0xc2, 0xbe, 0x23, 0xcd, 0x11,
	0x47, 0xf7, 0xa4, 0xf0, 0x1c, 0x5b, 0x3e, 0x8b, 0x3b, 0xff, 0xca, 0xc0, 0xf3, 0x82, 0xc3, 0x7d,
	0x70, 0xcd, 0xef, 0x90, 0x1e, 0x51, 0x81, 0x3c, 0xfa, 0xa7, 0x21, 0x47, 0x31, 0x10, 0x71, 0x74,
	0x27, 0xbe, 0x53, 0xb1, 0x9b, 0x6a, 0xdd, 0x64, 0x21, 0x7a, 0xb6, 0x94, 0xac, 0x71, 0xec, 0x02,
	0x8f, 0x41, 0xd9, 0x24, 0xad, 0xa0, 0xdd, 0xa6, 0x76, 0x5b, 0xbd, 0x2e, 0x4f, 0xf5, 0x34, 0xe4,
	0x68, 0x02, 0x66, 0xd5, 0x9c, 0x21, 0xd9, 0xe7, 0xaa, 0xe4, 0x21, 0x3c, 0x71, 0x82, 0xbf, 0x2b,
	0x40, 0xcd, 0x6e, 0x8e, 0x75, 0xa9, 0xab, 0x75, 0x1c, 0xe6, 0x6b, 0x46, 0x87, 0x18, 0x5d, 0x75,
	0x49, 0xca, 0xfc, 0x20, 0xfa, 0x3a, 0xe5, 0x1c, 0x75, 0xa9, 0xfb, 0xca, 0x61, 0xbe, 0x24, 0x64,
	0x7d, 0x3d, 0xd7, 0x7a, 0xa9, 0xaf, 0x3f, 0xc0, 0x89, 0x86, 0xf5, 0xf9, 0x22, 0x78, 0x06, 0x6e,
	0x0a, 0x18, 0xfe, 0xa6, 0x80, 0x8f, 0x26, 0xdf, 0xdc, 0xb2, 0x9c, 0xbe, 0x76, 0xe2, 0xe9, 0x3d,
	0xa2, 0x59, 0x8e, 0x6e, 0x8a, 0x4b, 0x5a, 0x96, 0xd9, 0x7f, 0x17, 0x72, 0x74, 0x3b, 0xfb, 0x3a,
	0x82, 0xf6, 0x52, 0xb0, 0x0e, 0x62, 0x52, 0xc4, 0xd1, 0xfd, 0x7c, 0x01, 0x5c, 0x66, 0xe4, 0x4f,
	0x71, 0xef, 0x3f, 0xf0, 0xf0, 0xfb, 0xe5, 0xe0, 0xb7, 0xa0, 0xd2, 0x27, 0x2d, 0xd1, 0x85, 0xb6,
	0xe6, 0xb9, 0x1a, 0x35, 0xd5, 0x8a, 0xac, 0x8d, 0xcf, 0x42, 0x8e, 0x96, 0x52, 0x0b, 0x76, 0xf7,
	0xc4, 0x2c, 0x81, 0x32, 0xb1, 0x69, 0x50, 0x3e, 0x21, 0xd3, 0x00, 0xce, 0xed, 0x20, 0x01, 0x2b,
	0x59, 0x74, 0xc7, 0xa3, 0x6d, 0x6a, 0xab, 0x2b, 0x32, 0xfc, 0xe7, 0x21, 0x47, 0x99, 0xf0, 0x97,
	0xd2, 0x92, 0x95, 0x4b, 0x1e, 0x96, 0xe5, 0x92, 0x87, 0xf0, 0xa5, 0x3d, 0xfc, 0x59, 0x01, 0x37,
	0x33, 0x1d, 0xc3, 0x23, 0x26, 0xb1, 0x7d, 0xaa, 0x5b, 0x4c, 0x5d, 0x5d, 0xbf, 0xb2, 0x71, 0x7d,
	0x6b, 0x2d, 0x1d, 0x5b, 0xdf, 0x24, 0x9c, 0x66, 0x46, 0xd9, 0x7d, 0x7d, 0xce, 0x51, 0x41, 0x34,
	0x63, 0x7f, 0xc6, 0xc6, 0xb2, 0x79, 0x36, 0x6b, 0x93, 0xf3, 0x6c, 0x16, 0xc6, 0xf3, 0xc2, 0xd4,
	0xfe, 0x2e, 0x02, 0x38, 0x2b, 0x0d, 0x9f, 0x83, 0x22, 0x8d, 0xdf, 0xee, 0xf2, 0x6e, 0x63, 0xcc,
	0x51, 0x71, 0xef, 0x45, 0xc8, 0x51, 0x91, 0x8a, 0xab, 0x5e, 0x8e, 0x6b, 0xc0, 0xcc, 0x3a, 0xa6,
	0x94, 0xac, 0xcf, 0x46, 0xf5, 0xe2, 0xde, 0x0b, 0x5c, 0xa4, 0x26, 0x3c, 0x04, 0x8b, 0x36, 0x35,
	0xba, 0xb6, 0xde, 0x23, 0xc9, 0xeb, 0xfc, 0x44, 0x4c, 0xe3, 0x14, 0x8b, 0x38, 0xba, 0x21, 0xa3,
	0xa4, 0x40, 0x16, 0x6b, 0x39, 0x87, 0xe0, 0xcc, 0x03, 0x76, 0xc1, 0x8a, 0x1b, 0xb4, 0x2c, 0x6a,
	0x88, 0xf9, 0xa8, 0x19, 0x0e, 0x23, 0xc9, 0x33, 0xdb, 0x0c, 0x39, 0x5a, 0x8e, 0x4d, 0xfb, 0x64,
	0xd0, 0x74, 0xd8, 0x64, 0xd6, 0xe7, 0xd0, 0x4c, 0x02, 0xce, 0xc2, 0x38, 0x1f, 0x00, 0x7e, 0x0d,
	0x00, 0xa3, 0x6d, 0x5b, 0x33, 0x9c, 0xc0, 0xf6, 0xe5, 0x4b, 0xb8, 0x1c, 0x0f, 0x10, 0x81, 0x36,
	0x05, 0x98, 0x55, 0x44, 0x86, 0x4c, 0x06, 0x48, 0x1e, 0xc2, 0x13, 0xa7, 0xdd, 0xfd, 0xf3, 0xb7,
	0xd5, 0xc2, 0xe8, 0x6d, 0xb5, 0x70, 0x3e, 0xae, 0x2a, 0xa3, 0x71, 0x55, 0xf9, 0xf1, 0xa2, 0x5a,
	0xf8, 0xf5, 0xa2, 0xaa, 0x8c, 0x2e, 0xaa, 0x85, 0xbf, 0x2e, 0xaa, 0x85, 0x37, 0xf7, 0xdb, 0xd4,
	0xef, 0x04, 0xad, 0x86, 0xe1, 0xf4, 0x36, 0xd9, 0xc0, 0x36, 0xfc, 0x0e, 0xb5, 0xdb, 0x53, 0xab,
	0xc9, 0x3f, 0x59, 0x6b, 0x41, 0xfe, 0x80, 0x3d, 0xfe, 0x27, 0x00, 0x00, 0xff, 0xff, 0x7b, 0x3d,
	0x50, 0xc3, 0xd3, 0x09, 0x00, 0x00,
}

func (m *GUIConfiguration) Marshal() (dAtA []byte, err error) {
	size := m.ProtoSize()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GUIConfiguration) MarshalTo(dAtA []byte) (int, error) {
	size := m.ProtoSize()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GUIConfiguration) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.WebauthnCredentials) > 0 {
		for iNdEx := len(m.WebauthnCredentials) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.WebauthnCredentials[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGuiconfiguration(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1
			i--
			dAtA[i] = 0x82
		}
	}
	if len(m.WebauthnOrigin) > 0 {
		i -= len(m.WebauthnOrigin)
		copy(dAtA[i:], m.WebauthnOrigin)
		i = encodeVarintGuiconfiguration(dAtA, i, uint64(len(m.WebauthnOrigin)))
		i--
		dAtA[i] = 0x7a
	}
	if len(m.WebauthnRpId) > 0 {
		i -= len(m.WebauthnRpId)
		copy(dAtA[i:], m.WebauthnRpId)
		i = encodeVarintGuiconfiguration(dAtA, i, uint64(len(m.WebauthnRpId)))
		i--
		dAtA[i] = 0x72
	}
	if m.InsecureAllowFrameLoading {
		i--
		if m.InsecureAllowFrameLoading {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x68
	}
	if m.InsecureSkipHostCheck {
		i--
		if m.InsecureSkipHostCheck {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x60
	}
	if m.Debugging {
		i--
		if m.Debugging {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x58
	}
	if len(m.Theme) > 0 {
		i -= len(m.Theme)
		copy(dAtA[i:], m.Theme)
		i = encodeVarintGuiconfiguration(dAtA, i, uint64(len(m.Theme)))
		i--
		dAtA[i] = 0x52
	}
	if m.InsecureAdminAccess {
		i--
		if m.InsecureAdminAccess {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x48
	}
	if len(m.APIKey) > 0 {
		i -= len(m.APIKey)
		copy(dAtA[i:], m.APIKey)
		i = encodeVarintGuiconfiguration(dAtA, i, uint64(len(m.APIKey)))
		i--
		dAtA[i] = 0x42
	}
	if m.RawUseTLS {
		i--
		if m.RawUseTLS {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x38
	}
	if m.AuthMode != 0 {
		i = encodeVarintGuiconfiguration(dAtA, i, uint64(m.AuthMode))
		i--
		dAtA[i] = 0x30
	}
	if len(m.Password) > 0 {
		i -= len(m.Password)
		copy(dAtA[i:], m.Password)
		i = encodeVarintGuiconfiguration(dAtA, i, uint64(len(m.Password)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.User) > 0 {
		i -= len(m.User)
		copy(dAtA[i:], m.User)
		i = encodeVarintGuiconfiguration(dAtA, i, uint64(len(m.User)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.RawUnixSocketPermissions) > 0 {
		i -= len(m.RawUnixSocketPermissions)
		copy(dAtA[i:], m.RawUnixSocketPermissions)
		i = encodeVarintGuiconfiguration(dAtA, i, uint64(len(m.RawUnixSocketPermissions)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.RawAddress) > 0 {
		i -= len(m.RawAddress)
		copy(dAtA[i:], m.RawAddress)
		i = encodeVarintGuiconfiguration(dAtA, i, uint64(len(m.RawAddress)))
		i--
		dAtA[i] = 0x12
	}
	if m.Enabled {
		i--
		if m.Enabled {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *WebauthnCredential) Marshal() (dAtA []byte, err error) {
	size := m.ProtoSize()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *WebauthnCredential) MarshalTo(dAtA []byte) (int, error) {
	size := m.ProtoSize()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *WebauthnCredential) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.SignCount != 0 {
		i = encodeVarintGuiconfiguration(dAtA, i, uint64(m.SignCount))
		i--
		dAtA[i] = 0x20
	}
	if len(m.PublicKeyCose) > 0 {
		i -= len(m.PublicKeyCose)
		copy(dAtA[i:], m.PublicKeyCose)
		i = encodeVarintGuiconfiguration(dAtA, i, uint64(len(m.PublicKeyCose)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Nickname) > 0 {
		i -= len(m.Nickname)
		copy(dAtA[i:], m.Nickname)
		i = encodeVarintGuiconfiguration(dAtA, i, uint64(len(m.Nickname)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ID) > 0 {
		i -= len(m.ID)
		copy(dAtA[i:], m.ID)
		i = encodeVarintGuiconfiguration(dAtA, i, uint64(len(m.ID)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintGuiconfiguration(dAtA []byte, offset int, v uint64) int {
	offset -= sovGuiconfiguration(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GUIConfiguration) ProtoSize() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Enabled {
		n += 2
	}
	l = len(m.RawAddress)
	if l > 0 {
		n += 1 + l + sovGuiconfiguration(uint64(l))
	}
	l = len(m.RawUnixSocketPermissions)
	if l > 0 {
		n += 1 + l + sovGuiconfiguration(uint64(l))
	}
	l = len(m.User)
	if l > 0 {
		n += 1 + l + sovGuiconfiguration(uint64(l))
	}
	l = len(m.Password)
	if l > 0 {
		n += 1 + l + sovGuiconfiguration(uint64(l))
	}
	if m.AuthMode != 0 {
		n += 1 + sovGuiconfiguration(uint64(m.AuthMode))
	}
	if m.RawUseTLS {
		n += 2
	}
	l = len(m.APIKey)
	if l > 0 {
		n += 1 + l + sovGuiconfiguration(uint64(l))
	}
	if m.InsecureAdminAccess {
		n += 2
	}
	l = len(m.Theme)
	if l > 0 {
		n += 1 + l + sovGuiconfiguration(uint64(l))
	}
	if m.Debugging {
		n += 2
	}
	if m.InsecureSkipHostCheck {
		n += 2
	}
	if m.InsecureAllowFrameLoading {
		n += 2
	}
	l = len(m.WebauthnRpId)
	if l > 0 {
		n += 1 + l + sovGuiconfiguration(uint64(l))
	}
	l = len(m.WebauthnOrigin)
	if l > 0 {
		n += 1 + l + sovGuiconfiguration(uint64(l))
	}
	if len(m.WebauthnCredentials) > 0 {
		for _, e := range m.WebauthnCredentials {
			l = e.ProtoSize()
			n += 2 + l + sovGuiconfiguration(uint64(l))
		}
	}
	return n
}

func (m *WebauthnCredential) ProtoSize() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ID)
	if l > 0 {
		n += 1 + l + sovGuiconfiguration(uint64(l))
	}
	l = len(m.Nickname)
	if l > 0 {
		n += 1 + l + sovGuiconfiguration(uint64(l))
	}
	l = len(m.PublicKeyCose)
	if l > 0 {
		n += 1 + l + sovGuiconfiguration(uint64(l))
	}
	if m.SignCount != 0 {
		n += 1 + sovGuiconfiguration(uint64(m.SignCount))
	}
	return n
}

func sovGuiconfiguration(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGuiconfiguration(x uint64) (n int) {
	return sovGuiconfiguration(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GUIConfiguration) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGuiconfiguration
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GUIConfiguration: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GUIConfiguration: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Enabled", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Enabled = bool(v != 0)
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RawAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RawAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RawUnixSocketPermissions", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RawUnixSocketPermissions = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field User", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.User = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Password", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Password = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AuthMode", wireType)
			}
			m.AuthMode = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.AuthMode |= AuthMode(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RawUseTLS", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.RawUseTLS = bool(v != 0)
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field APIKey", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.APIKey = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field InsecureAdminAccess", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.InsecureAdminAccess = bool(v != 0)
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Theme", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Theme = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 11:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Debugging", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Debugging = bool(v != 0)
		case 12:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field InsecureSkipHostCheck", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.InsecureSkipHostCheck = bool(v != 0)
		case 13:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field InsecureAllowFrameLoading", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.InsecureAllowFrameLoading = bool(v != 0)
		case 14:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field WebauthnRpId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.WebauthnRpId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 15:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field WebauthnOrigin", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.WebauthnOrigin = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 16:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field WebauthnCredentials", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.WebauthnCredentials = append(m.WebauthnCredentials, WebauthnCredential{})
			if err := m.WebauthnCredentials[len(m.WebauthnCredentials)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGuiconfiguration(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *WebauthnCredential) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGuiconfiguration
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: WebauthnCredential: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: WebauthnCredential: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Nickname", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Nickname = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PublicKeyCose", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PublicKeyCose = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SignCount", wireType)
			}
			m.SignCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SignCount |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipGuiconfiguration(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGuiconfiguration
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipGuiconfiguration(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGuiconfiguration
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGuiconfiguration
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthGuiconfiguration
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGuiconfiguration
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGuiconfiguration
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGuiconfiguration        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGuiconfiguration          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGuiconfiguration = fmt.Errorf("proto: unexpected end of group")
)
