// To regenerate generator.pb.go run
//   protoc --proto_path=./src/generator/proto --go_out=. --go-grpc_out=. generator.proto

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.20.3
// source: generator.proto

package proto

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

type GenerateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success  bool     `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	ConfList [][]byte `protobuf:"bytes,2,rep,name=confList,proto3" json:"confList,omitempty"`
	Message  string   `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *GenerateResponse) Reset() {
	*x = GenerateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_generator_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GenerateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenerateResponse) ProtoMessage() {}

func (x *GenerateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_generator_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GenerateResponse.ProtoReflect.Descriptor instead.
func (*GenerateResponse) Descriptor() ([]byte, []int) {
	return file_generator_proto_rawDescGZIP(), []int{0}
}

func (x *GenerateResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *GenerateResponse) GetConfList() [][]byte {
	if x != nil {
		return x.ConfList
	}
	return nil
}

func (x *GenerateResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type ConfigRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PodName      string `protobuf:"bytes,1,opt,name=podName,proto3" json:"podName,omitempty"`
	PodNamespace string `protobuf:"bytes,2,opt,name=podNamespace,proto3" json:"podNamespace,omitempty"`
	HostName     string `protobuf:"bytes,3,opt,name=hostName,proto3" json:"hostName,omitempty"`
	Data         []byte `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *ConfigRequest) Reset() {
	*x = ConfigRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_generator_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConfigRequest) ProtoMessage() {}

func (x *ConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_generator_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConfigRequest.ProtoReflect.Descriptor instead.
func (*ConfigRequest) Descriptor() ([]byte, []int) {
	return file_generator_proto_rawDescGZIP(), []int{1}
}

func (x *ConfigRequest) GetPodName() string {
	if x != nil {
		return x.PodName
	}
	return ""
}

func (x *ConfigRequest) GetPodNamespace() string {
	if x != nil {
		return x.PodNamespace
	}
	return ""
}

func (x *ConfigRequest) GetHostName() string {
	if x != nil {
		return x.HostName
	}
	return ""
}

func (x *ConfigRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_generator_proto protoreflect.FileDescriptor

var file_generator_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x09, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x22, 0x62, 0x0a, 0x10,
	0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f,
	0x6e, 0x66, 0x4c, 0x69, 0x73, 0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x08, 0x63, 0x6f,
	0x6e, 0x66, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x22, 0x7d, 0x0a, 0x0d, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x6f, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x70, 0x6f, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x22, 0x0a, 0x0c, 0x70,
	0x6f, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0c, 0x70, 0x6f, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12,
	0x1a, 0x0a, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x32,
	0x94, 0x01, 0x0a, 0x09, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x43, 0x0a,
	0x08, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x12, 0x18, 0x2e, 0x67, 0x65, 0x6e, 0x65,
	0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e,
	0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x42, 0x0a, 0x07, 0x43, 0x6c, 0x65, 0x61, 0x6e, 0x75, 0x70, 0x12, 0x18, 0x2e,
	0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61,
	0x74, 0x6f, 0x72, 0x2e, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x17, 0x5a, 0x15, 0x2e, 0x2f, 0x73, 0x72, 0x63, 0x2f,
	0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_generator_proto_rawDescOnce sync.Once
	file_generator_proto_rawDescData = file_generator_proto_rawDesc
)

func file_generator_proto_rawDescGZIP() []byte {
	file_generator_proto_rawDescOnce.Do(func() {
		file_generator_proto_rawDescData = protoimpl.X.CompressGZIP(file_generator_proto_rawDescData)
	})
	return file_generator_proto_rawDescData
}

var file_generator_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_generator_proto_goTypes = []interface{}{
	(*GenerateResponse)(nil), // 0: generator.GenerateResponse
	(*ConfigRequest)(nil),    // 1: generator.ConfigRequest
}
var file_generator_proto_depIdxs = []int32{
	1, // 0: generator.Generator.Generate:input_type -> generator.ConfigRequest
	1, // 1: generator.Generator.Cleanup:input_type -> generator.ConfigRequest
	0, // 2: generator.Generator.Generate:output_type -> generator.GenerateResponse
	0, // 3: generator.Generator.Cleanup:output_type -> generator.GenerateResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_generator_proto_init() }
func file_generator_proto_init() {
	if File_generator_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_generator_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GenerateResponse); i {
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
		file_generator_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConfigRequest); i {
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
			RawDescriptor: file_generator_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_generator_proto_goTypes,
		DependencyIndexes: file_generator_proto_depIdxs,
		MessageInfos:      file_generator_proto_msgTypes,
	}.Build()
	File_generator_proto = out.File
	file_generator_proto_rawDesc = nil
	file_generator_proto_goTypes = nil
	file_generator_proto_depIdxs = nil
}
