// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.23.0--rc1
// source: template/v1/common.proto

package v1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type BoolReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code int32           `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Msg  string          `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
	Data *BoolReply_Data `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *BoolReply) Reset() {
	*x = BoolReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_template_v1_common_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BoolReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BoolReply) ProtoMessage() {}

func (x *BoolReply) ProtoReflect() protoreflect.Message {
	mi := &file_template_v1_common_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BoolReply.ProtoReflect.Descriptor instead.
func (*BoolReply) Descriptor() ([]byte, []int) {
	return file_template_v1_common_proto_rawDescGZIP(), []int{0}
}

func (x *BoolReply) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *BoolReply) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

func (x *BoolReply) GetData() *BoolReply_Data {
	if x != nil {
		return x.Data
	}
	return nil
}

type BoolReply_Data struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ok bool `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
}

func (x *BoolReply_Data) Reset() {
	*x = BoolReply_Data{}
	if protoimpl.UnsafeEnabled {
		mi := &file_template_v1_common_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BoolReply_Data) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BoolReply_Data) ProtoMessage() {}

func (x *BoolReply_Data) ProtoReflect() protoreflect.Message {
	mi := &file_template_v1_common_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BoolReply_Data.ProtoReflect.Descriptor instead.
func (*BoolReply_Data) Descriptor() ([]byte, []int) {
	return file_template_v1_common_proto_rawDescGZIP(), []int{0, 0}
}

func (x *BoolReply_Data) GetOk() bool {
	if x != nil {
		return x.Ok
	}
	return false
}

var File_template_v1_common_proto protoreflect.FileDescriptor

var file_template_v1_common_proto_rawDesc = []byte{
	0x0a, 0x18, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x61, 0x70, 0x69, 0x2e,
	0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x22, 0x7e, 0x0a, 0x09, 0x42,
	0x6f, 0x6f, 0x6c, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x10, 0x0a, 0x03,
	0x6d, 0x73, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x12, 0x33,
	0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x61,
	0x70, 0x69, 0x2e, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x42,
	0x6f, 0x6f, 0x6c, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x52, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x1a, 0x16, 0x0a, 0x04, 0x44, 0x61, 0x74, 0x61, 0x12, 0x0e, 0x0a, 0x02, 0x6f,
	0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x02, 0x6f, 0x6b, 0x42, 0x37, 0x0a, 0x0f, 0x61,
	0x70, 0x69, 0x2e, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x50, 0x01,
	0x5a, 0x22, 0x6d, 0x65, 0x65, 0x74, 0x69, 0x6e, 0x67, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x76,
	0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_template_v1_common_proto_rawDescOnce sync.Once
	file_template_v1_common_proto_rawDescData = file_template_v1_common_proto_rawDesc
)

func file_template_v1_common_proto_rawDescGZIP() []byte {
	file_template_v1_common_proto_rawDescOnce.Do(func() {
		file_template_v1_common_proto_rawDescData = protoimpl.X.CompressGZIP(file_template_v1_common_proto_rawDescData)
	})
	return file_template_v1_common_proto_rawDescData
}

var file_template_v1_common_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_template_v1_common_proto_goTypes = []interface{}{
	(*BoolReply)(nil),      // 0: api.template.v1.BoolReply
	(*BoolReply_Data)(nil), // 1: api.template.v1.BoolReply.Data
}
var file_template_v1_common_proto_depIdxs = []int32{
	1, // 0: api.template.v1.BoolReply.data:type_name -> api.template.v1.BoolReply.Data
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_template_v1_common_proto_init() }
func file_template_v1_common_proto_init() {
	if File_template_v1_common_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_template_v1_common_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BoolReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_template_v1_common_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BoolReply_Data); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_template_v1_common_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_template_v1_common_proto_goTypes,
		DependencyIndexes: file_template_v1_common_proto_depIdxs,
		MessageInfos:      file_template_v1_common_proto_msgTypes,
	}.Build()
	File_template_v1_common_proto = out.File
	file_template_v1_common_proto_rawDesc = nil
	file_template_v1_common_proto_goTypes = nil
	file_template_v1_common_proto_depIdxs = nil
}
