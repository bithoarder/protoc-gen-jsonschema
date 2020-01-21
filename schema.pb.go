// Code generated by protoc-gen-go. DO NOT EDIT.
// source: schema.proto

package main

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Schema struct {
	Description string `protobuf:"bytes,10,opt,name=description,proto3" json:"description,omitempty"`
	// Types that are valid to be assigned to Rule:
	//	*Schema_Ref
	//	*Schema_Pattern
	//	*Schema_Min
	//	*Schema_Max
	//	*Schema_Range_
	Rule                 isSchema_Rule `protobuf_oneof:"rule"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Schema) Reset()         { *m = Schema{} }
func (m *Schema) String() string { return proto.CompactTextString(m) }
func (*Schema) ProtoMessage()    {}
func (*Schema) Descriptor() ([]byte, []int) {
	return fileDescriptor_1c5fb4d8cc22d66a, []int{0}
}

func (m *Schema) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Schema.Unmarshal(m, b)
}
func (m *Schema) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Schema.Marshal(b, m, deterministic)
}
func (m *Schema) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Schema.Merge(m, src)
}
func (m *Schema) XXX_Size() int {
	return xxx_messageInfo_Schema.Size(m)
}
func (m *Schema) XXX_DiscardUnknown() {
	xxx_messageInfo_Schema.DiscardUnknown(m)
}

var xxx_messageInfo_Schema proto.InternalMessageInfo

func (m *Schema) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

type isSchema_Rule interface {
	isSchema_Rule()
}

type Schema_Ref struct {
	Ref string `protobuf:"bytes,1,opt,name=ref,proto3,oneof"`
}

type Schema_Pattern struct {
	Pattern string `protobuf:"bytes,2,opt,name=pattern,proto3,oneof"`
}

type Schema_Min struct {
	Min int64 `protobuf:"varint,4,opt,name=min,proto3,oneof"`
}

type Schema_Max struct {
	Max int64 `protobuf:"varint,5,opt,name=max,proto3,oneof"`
}

type Schema_Range_ struct {
	Range *Schema_Range `protobuf:"bytes,6,opt,name=range,proto3,oneof"`
}

func (*Schema_Ref) isSchema_Rule() {}

func (*Schema_Pattern) isSchema_Rule() {}

func (*Schema_Min) isSchema_Rule() {}

func (*Schema_Max) isSchema_Rule() {}

func (*Schema_Range_) isSchema_Rule() {}

func (m *Schema) GetRule() isSchema_Rule {
	if m != nil {
		return m.Rule
	}
	return nil
}

func (m *Schema) GetRef() string {
	if x, ok := m.GetRule().(*Schema_Ref); ok {
		return x.Ref
	}
	return ""
}

func (m *Schema) GetPattern() string {
	if x, ok := m.GetRule().(*Schema_Pattern); ok {
		return x.Pattern
	}
	return ""
}

func (m *Schema) GetMin() int64 {
	if x, ok := m.GetRule().(*Schema_Min); ok {
		return x.Min
	}
	return 0
}

func (m *Schema) GetMax() int64 {
	if x, ok := m.GetRule().(*Schema_Max); ok {
		return x.Max
	}
	return 0
}

func (m *Schema) GetRange() *Schema_Range {
	if x, ok := m.GetRule().(*Schema_Range_); ok {
		return x.Range
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Schema) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Schema_Ref)(nil),
		(*Schema_Pattern)(nil),
		(*Schema_Min)(nil),
		(*Schema_Max)(nil),
		(*Schema_Range_)(nil),
	}
}

type Schema_Range struct {
	Min                  int64    `protobuf:"varint,1,opt,name=min,proto3" json:"min,omitempty"`
	Max                  int64    `protobuf:"varint,2,opt,name=max,proto3" json:"max,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Schema_Range) Reset()         { *m = Schema_Range{} }
func (m *Schema_Range) String() string { return proto.CompactTextString(m) }
func (*Schema_Range) ProtoMessage()    {}
func (*Schema_Range) Descriptor() ([]byte, []int) {
	return fileDescriptor_1c5fb4d8cc22d66a, []int{0, 0}
}

func (m *Schema_Range) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Schema_Range.Unmarshal(m, b)
}
func (m *Schema_Range) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Schema_Range.Marshal(b, m, deterministic)
}
func (m *Schema_Range) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Schema_Range.Merge(m, src)
}
func (m *Schema_Range) XXX_Size() int {
	return xxx_messageInfo_Schema_Range.Size(m)
}
func (m *Schema_Range) XXX_DiscardUnknown() {
	xxx_messageInfo_Schema_Range.DiscardUnknown(m)
}

var xxx_messageInfo_Schema_Range proto.InternalMessageInfo

func (m *Schema_Range) GetMin() int64 {
	if m != nil {
		return m.Min
	}
	return 0
}

func (m *Schema_Range) GetMax() int64 {
	if m != nil {
		return m.Max
	}
	return 0
}

var E_JsonSchema = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.FieldOptions)(nil),
	ExtensionType: (*Schema)(nil),
	Field:         51000,
	Name:          "main.json_schema",
	Tag:           "bytes,51000,opt,name=json_schema",
	Filename:      "schema.proto",
}

func init() {
	proto.RegisterType((*Schema)(nil), "main.Schema")
	proto.RegisterType((*Schema_Range)(nil), "main.Schema.Range")
	proto.RegisterExtension(E_JsonSchema)
}

func init() { proto.RegisterFile("schema.proto", fileDescriptor_1c5fb4d8cc22d66a) }

var fileDescriptor_1c5fb4d8cc22d66a = []byte{
	// 261 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x90, 0x4d, 0x6a, 0xc3, 0x30,
	0x14, 0x84, 0xe3, 0xf8, 0xa7, 0xf4, 0x39, 0x8b, 0xf0, 0x56, 0xc2, 0x50, 0x10, 0x5d, 0x99, 0x16,
	0x14, 0x68, 0x77, 0x5d, 0x76, 0x51, 0xb2, 0x6a, 0x41, 0x3d, 0x40, 0x51, 0x12, 0xc5, 0x55, 0xb1,
	0x25, 0x23, 0x3b, 0xe0, 0x53, 0xf4, 0x3c, 0xbd, 0x4f, 0x2f, 0x52, 0xa4, 0x17, 0x97, 0xec, 0xec,
	0x6f, 0xe6, 0x0d, 0xa3, 0x81, 0xd5, 0xb0, 0xff, 0xd4, 0x9d, 0x12, 0xbd, 0x77, 0xa3, 0xc3, 0xac,
	0x53, 0xc6, 0x56, 0xbc, 0x71, 0xae, 0x69, 0xf5, 0x26, 0xb2, 0xdd, 0xe9, 0xb8, 0x39, 0xe8, 0x61,
	0xef, 0x4d, 0x3f, 0x3a, 0x4f, 0xbe, 0xdb, 0xdf, 0x04, 0x8a, 0xf7, 0x78, 0x88, 0x1c, 0xca, 0x59,
	0x36, 0xce, 0x32, 0xe0, 0x49, 0x7d, 0x2d, 0x2f, 0x11, 0x22, 0xa4, 0x5e, 0x1f, 0x59, 0x12, 0x94,
	0xed, 0x42, 0x86, 0x1f, 0xac, 0xe0, 0xaa, 0x57, 0xe3, 0xa8, 0xbd, 0x65, 0xcb, 0x33, 0x9f, 0x41,
	0xf0, 0x77, 0xc6, 0xb2, 0x8c, 0x27, 0x75, 0x1a, 0xfc, 0x9d, 0x21, 0xa6, 0x26, 0x96, 0xff, 0x33,
	0x35, 0xe1, 0x1d, 0xe4, 0x5e, 0xd9, 0x46, 0xb3, 0x82, 0x27, 0x75, 0xf9, 0x80, 0x22, 0x94, 0x17,
	0x54, 0x4b, 0xc8, 0xa0, 0x6c, 0x17, 0x92, 0x2c, 0xd5, 0x3d, 0xe4, 0x91, 0xe0, 0x9a, 0xc2, 0x43,
	0x99, 0x94, 0xa2, 0xd7, 0x14, 0xbd, 0x3c, 0x13, 0x35, 0x3d, 0x17, 0x90, 0xf9, 0x53, 0xab, 0x9f,
	0x5e, 0xa1, 0xfc, 0x1a, 0x9c, 0xfd, 0xa0, 0x89, 0xf0, 0x46, 0xd0, 0x2e, 0x62, 0xde, 0x45, 0xbc,
	0x18, 0xdd, 0x1e, 0xde, 0xe2, 0x2b, 0x07, 0xf6, 0xf3, 0x9d, 0xc6, 0x1e, 0xab, 0xcb, 0x1e, 0x12,
	0x42, 0x02, 0x7d, 0xef, 0x8a, 0x78, 0xf8, 0xf8, 0x17, 0x00, 0x00, 0xff, 0xff, 0x9a, 0xb7, 0x03,
	0xed, 0x74, 0x01, 0x00, 0x00,
}
