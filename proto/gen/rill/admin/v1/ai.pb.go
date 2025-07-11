// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: rill/admin/v1/ai.proto

package adminv1

import (
	v1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
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

type CompleteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Messages []*v1.CompletionMessage `protobuf:"bytes,1,rep,name=messages,proto3" json:"messages,omitempty"`
	// Optional list of tools that the AI can use during completion
	Tools []*v1.Tool `protobuf:"bytes,2,rep,name=tools,proto3" json:"tools,omitempty"`
}

func (x *CompleteRequest) Reset() {
	*x = CompleteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_admin_v1_ai_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CompleteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CompleteRequest) ProtoMessage() {}

func (x *CompleteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_rill_admin_v1_ai_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CompleteRequest.ProtoReflect.Descriptor instead.
func (*CompleteRequest) Descriptor() ([]byte, []int) {
	return file_rill_admin_v1_ai_proto_rawDescGZIP(), []int{0}
}

func (x *CompleteRequest) GetMessages() []*v1.CompletionMessage {
	if x != nil {
		return x.Messages
	}
	return nil
}

func (x *CompleteRequest) GetTools() []*v1.Tool {
	if x != nil {
		return x.Tools
	}
	return nil
}

type CompleteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message *v1.CompletionMessage `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *CompleteResponse) Reset() {
	*x = CompleteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_admin_v1_ai_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CompleteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CompleteResponse) ProtoMessage() {}

func (x *CompleteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_rill_admin_v1_ai_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CompleteResponse.ProtoReflect.Descriptor instead.
func (*CompleteResponse) Descriptor() ([]byte, []int) {
	return file_rill_admin_v1_ai_proto_rawDescGZIP(), []int{1}
}

func (x *CompleteResponse) GetMessage() *v1.CompletionMessage {
	if x != nil {
		return x.Message
	}
	return nil
}

var File_rill_admin_v1_ai_proto protoreflect.FileDescriptor

var file_rill_admin_v1_ai_proto_rawDesc = []byte{
	0x0a, 0x16, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x76, 0x31, 0x2f,
	0x61, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x61,
	0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x13, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x61, 0x69, 0x2f, 0x76,
	0x31, 0x2f, 0x61, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x74, 0x0a, 0x0f, 0x43, 0x6f,
	0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x39, 0x0a,
	0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1d, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x61, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6d,
	0x70, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x08,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x12, 0x26, 0x0a, 0x05, 0x74, 0x6f, 0x6f, 0x6c,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x61,
	0x69, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x6f, 0x6f, 0x6c, 0x52, 0x05, 0x74, 0x6f, 0x6f, 0x6c, 0x73,
	0x22, 0x4b, 0x0a, 0x10, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x37, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x61, 0x69, 0x2e,
	0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x32, 0x74, 0x0a,
	0x09, 0x41, 0x49, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x67, 0x0a, 0x08, 0x43, 0x6f,
	0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x1e, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1a, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x14, 0x3a,
	0x01, 0x2a, 0x22, 0x0f, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x69, 0x2f, 0x63, 0x6f, 0x6d, 0x70, 0x6c,
	0x65, 0x74, 0x65, 0x42, 0xac, 0x01, 0x0a, 0x11, 0x63, 0x6f, 0x6d, 0x2e, 0x72, 0x69, 0x6c, 0x6c,
	0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x76, 0x31, 0x42, 0x07, 0x41, 0x69, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x38, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x64, 0x61, 0x74, 0x61, 0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2f, 0x76, 0x31, 0x3b, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x76, 0x31, 0xa2, 0x02,
	0x03, 0x52, 0x41, 0x58, 0xaa, 0x02, 0x0d, 0x52, 0x69, 0x6c, 0x6c, 0x2e, 0x41, 0x64, 0x6d, 0x69,
	0x6e, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x0d, 0x52, 0x69, 0x6c, 0x6c, 0x5c, 0x41, 0x64, 0x6d, 0x69,
	0x6e, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x19, 0x52, 0x69, 0x6c, 0x6c, 0x5c, 0x41, 0x64, 0x6d, 0x69,
	0x6e, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0xea, 0x02, 0x0f, 0x52, 0x69, 0x6c, 0x6c, 0x3a, 0x3a, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x3a, 0x3a,
	0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rill_admin_v1_ai_proto_rawDescOnce sync.Once
	file_rill_admin_v1_ai_proto_rawDescData = file_rill_admin_v1_ai_proto_rawDesc
)

func file_rill_admin_v1_ai_proto_rawDescGZIP() []byte {
	file_rill_admin_v1_ai_proto_rawDescOnce.Do(func() {
		file_rill_admin_v1_ai_proto_rawDescData = protoimpl.X.CompressGZIP(file_rill_admin_v1_ai_proto_rawDescData)
	})
	return file_rill_admin_v1_ai_proto_rawDescData
}

var file_rill_admin_v1_ai_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_rill_admin_v1_ai_proto_goTypes = []any{
	(*CompleteRequest)(nil),      // 0: rill.admin.v1.CompleteRequest
	(*CompleteResponse)(nil),     // 1: rill.admin.v1.CompleteResponse
	(*v1.CompletionMessage)(nil), // 2: rill.ai.v1.CompletionMessage
	(*v1.Tool)(nil),              // 3: rill.ai.v1.Tool
}
var file_rill_admin_v1_ai_proto_depIdxs = []int32{
	2, // 0: rill.admin.v1.CompleteRequest.messages:type_name -> rill.ai.v1.CompletionMessage
	3, // 1: rill.admin.v1.CompleteRequest.tools:type_name -> rill.ai.v1.Tool
	2, // 2: rill.admin.v1.CompleteResponse.message:type_name -> rill.ai.v1.CompletionMessage
	0, // 3: rill.admin.v1.AIService.Complete:input_type -> rill.admin.v1.CompleteRequest
	1, // 4: rill.admin.v1.AIService.Complete:output_type -> rill.admin.v1.CompleteResponse
	4, // [4:5] is the sub-list for method output_type
	3, // [3:4] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_rill_admin_v1_ai_proto_init() }
func file_rill_admin_v1_ai_proto_init() {
	if File_rill_admin_v1_ai_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_rill_admin_v1_ai_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*CompleteRequest); i {
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
		file_rill_admin_v1_ai_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*CompleteResponse); i {
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
			RawDescriptor: file_rill_admin_v1_ai_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_rill_admin_v1_ai_proto_goTypes,
		DependencyIndexes: file_rill_admin_v1_ai_proto_depIdxs,
		MessageInfos:      file_rill_admin_v1_ai_proto_msgTypes,
	}.Build()
	File_rill_admin_v1_ai_proto = out.File
	file_rill_admin_v1_ai_proto_rawDesc = nil
	file_rill_admin_v1_ai_proto_goTypes = nil
	file_rill_admin_v1_ai_proto_depIdxs = nil
}
