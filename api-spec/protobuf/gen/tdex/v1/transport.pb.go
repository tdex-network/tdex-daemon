// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: tdex/v1/transport.proto

package tdexv1

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type SupportedContentTypesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SupportedContentTypesRequest) Reset() {
	*x = SupportedContentTypesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tdex_v1_transport_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SupportedContentTypesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SupportedContentTypesRequest) ProtoMessage() {}

func (x *SupportedContentTypesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_tdex_v1_transport_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SupportedContentTypesRequest.ProtoReflect.Descriptor instead.
func (*SupportedContentTypesRequest) Descriptor() ([]byte, []int) {
	return file_tdex_v1_transport_proto_rawDescGZIP(), []int{0}
}

type SupportedContentTypesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AcceptedTypes []ContentType `protobuf:"varint,1,rep,packed,name=accepted_types,json=acceptedTypes,proto3,enum=tdex.v1.ContentType" json:"accepted_types,omitempty"`
}

func (x *SupportedContentTypesResponse) Reset() {
	*x = SupportedContentTypesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tdex_v1_transport_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SupportedContentTypesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SupportedContentTypesResponse) ProtoMessage() {}

func (x *SupportedContentTypesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_tdex_v1_transport_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SupportedContentTypesResponse.ProtoReflect.Descriptor instead.
func (*SupportedContentTypesResponse) Descriptor() ([]byte, []int) {
	return file_tdex_v1_transport_proto_rawDescGZIP(), []int{1}
}

func (x *SupportedContentTypesResponse) GetAcceptedTypes() []ContentType {
	if x != nil {
		return x.AcceptedTypes
	}
	return nil
}

var File_tdex_v1_transport_proto protoreflect.FileDescriptor

var file_tdex_v1_transport_proto_rawDesc = []byte{
	0x0a, 0x17, 0x74, 0x64, 0x65, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70,
	0x6f, 0x72, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x74, 0x64, 0x65, 0x78, 0x2e,
	0x76, 0x31, 0x1a, 0x13, 0x74, 0x64, 0x65, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x79, 0x70, 0x65,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x1e, 0x0a, 0x1c, 0x53, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74,
	0x65, 0x64, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x5c, 0x0a, 0x1d, 0x53, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74,
	0x65, 0x64, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3b, 0x0a, 0x0e, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74,
	0x65, 0x64, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x14,
	0x2e, 0x74, 0x64, 0x65, 0x78, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x54, 0x79, 0x70, 0x65, 0x52, 0x0d, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x54, 0x79,
	0x70, 0x65, 0x73, 0x32, 0x91, 0x01, 0x0a, 0x10, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72,
	0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x7d, 0x0a, 0x15, 0x53, 0x75, 0x70, 0x70,
	0x6f, 0x72, 0x74, 0x65, 0x64, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65,
	0x73, 0x12, 0x25, 0x2e, 0x74, 0x64, 0x65, 0x78, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x75, 0x70, 0x70,
	0x6f, 0x72, 0x74, 0x65, 0x64, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x74, 0x64, 0x65, 0x78, 0x2e,
	0x76, 0x31, 0x2e, 0x53, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x65, 0x64, 0x43, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x15, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0f, 0x12, 0x0d, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x72,
	0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x42, 0xa4, 0x01, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x2e,
	0x74, 0x64, 0x65, 0x78, 0x2e, 0x76, 0x31, 0x42, 0x0e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f,
	0x72, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x48, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x64, 0x65, 0x78, 0x2d, 0x6e, 0x65, 0x74, 0x77, 0x6f,
	0x72, 0x6b, 0x2f, 0x74, 0x64, 0x65, 0x78, 0x2d, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2f, 0x61,
	0x70, 0x69, 0x2d, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x74, 0x64, 0x65, 0x78, 0x2f, 0x76, 0x31, 0x3b, 0x74, 0x64, 0x65,
	0x78, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x54, 0x58, 0x58, 0xaa, 0x02, 0x07, 0x54, 0x64, 0x65, 0x78,
	0x2e, 0x56, 0x31, 0xca, 0x02, 0x07, 0x54, 0x64, 0x65, 0x78, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x13,
	0x54, 0x64, 0x65, 0x78, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0xea, 0x02, 0x08, 0x54, 0x64, 0x65, 0x78, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_tdex_v1_transport_proto_rawDescOnce sync.Once
	file_tdex_v1_transport_proto_rawDescData = file_tdex_v1_transport_proto_rawDesc
)

func file_tdex_v1_transport_proto_rawDescGZIP() []byte {
	file_tdex_v1_transport_proto_rawDescOnce.Do(func() {
		file_tdex_v1_transport_proto_rawDescData = protoimpl.X.CompressGZIP(file_tdex_v1_transport_proto_rawDescData)
	})
	return file_tdex_v1_transport_proto_rawDescData
}

var file_tdex_v1_transport_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_tdex_v1_transport_proto_goTypes = []interface{}{
	(*SupportedContentTypesRequest)(nil),  // 0: tdex.v1.SupportedContentTypesRequest
	(*SupportedContentTypesResponse)(nil), // 1: tdex.v1.SupportedContentTypesResponse
	(ContentType)(0),                      // 2: tdex.v1.ContentType
}
var file_tdex_v1_transport_proto_depIdxs = []int32{
	2, // 0: tdex.v1.SupportedContentTypesResponse.accepted_types:type_name -> tdex.v1.ContentType
	0, // 1: tdex.v1.TransportService.SupportedContentTypes:input_type -> tdex.v1.SupportedContentTypesRequest
	1, // 2: tdex.v1.TransportService.SupportedContentTypes:output_type -> tdex.v1.SupportedContentTypesResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_tdex_v1_transport_proto_init() }
func file_tdex_v1_transport_proto_init() {
	if File_tdex_v1_transport_proto != nil {
		return
	}
	file_tdex_v1_types_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_tdex_v1_transport_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SupportedContentTypesRequest); i {
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
		file_tdex_v1_transport_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SupportedContentTypesResponse); i {
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
			RawDescriptor: file_tdex_v1_transport_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_tdex_v1_transport_proto_goTypes,
		DependencyIndexes: file_tdex_v1_transport_proto_depIdxs,
		MessageInfos:      file_tdex_v1_transport_proto_msgTypes,
	}.Build()
	File_tdex_v1_transport_proto = out.File
	file_tdex_v1_transport_proto_rawDesc = nil
	file_tdex_v1_transport_proto_goTypes = nil
	file_tdex_v1_transport_proto_depIdxs = nil
}
