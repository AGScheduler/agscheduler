// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.21.12
// source: cluster.proto

package services

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Node struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EndpointMain      string                 `protobuf:"bytes,1,opt,name=endpoint_main,json=endpointMain,proto3" json:"endpoint_main,omitempty"`
	Endpoint          string                 `protobuf:"bytes,2,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	EndpointGrpc      string                 `protobuf:"bytes,3,opt,name=endpoint_grpc,json=endpointGrpc,proto3" json:"endpoint_grpc,omitempty"`
	EndpointHttp      string                 `protobuf:"bytes,4,opt,name=endpoint_http,json=endpointHttp,proto3" json:"endpoint_http,omitempty"`
	Queue             string                 `protobuf:"bytes,5,opt,name=queue,proto3" json:"queue,omitempty"`
	Mode              string                 `protobuf:"bytes,6,opt,name=mode,proto3" json:"mode,omitempty"`
	Version           string                 `protobuf:"bytes,7,opt,name=version,proto3" json:"version,omitempty"`
	Health            bool                   `protobuf:"varint,8,opt,name=health,proto3" json:"health,omitempty"`
	RegisterTime      *timestamppb.Timestamp `protobuf:"bytes,9,opt,name=register_time,json=registerTime,proto3" json:"register_time,omitempty"`
	LastHeartbeatTime *timestamppb.Timestamp `protobuf:"bytes,10,opt,name=last_heartbeat_time,json=lastHeartbeatTime,proto3" json:"last_heartbeat_time,omitempty"`
}

func (x *Node) Reset() {
	*x = Node{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cluster_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Node) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Node) ProtoMessage() {}

func (x *Node) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Node.ProtoReflect.Descriptor instead.
func (*Node) Descriptor() ([]byte, []int) {
	return file_cluster_proto_rawDescGZIP(), []int{0}
}

func (x *Node) GetEndpointMain() string {
	if x != nil {
		return x.EndpointMain
	}
	return ""
}

func (x *Node) GetEndpoint() string {
	if x != nil {
		return x.Endpoint
	}
	return ""
}

func (x *Node) GetEndpointGrpc() string {
	if x != nil {
		return x.EndpointGrpc
	}
	return ""
}

func (x *Node) GetEndpointHttp() string {
	if x != nil {
		return x.EndpointHttp
	}
	return ""
}

func (x *Node) GetQueue() string {
	if x != nil {
		return x.Queue
	}
	return ""
}

func (x *Node) GetMode() string {
	if x != nil {
		return x.Mode
	}
	return ""
}

func (x *Node) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *Node) GetHealth() bool {
	if x != nil {
		return x.Health
	}
	return false
}

func (x *Node) GetRegisterTime() *timestamppb.Timestamp {
	if x != nil {
		return x.RegisterTime
	}
	return nil
}

func (x *Node) GetLastHeartbeatTime() *timestamppb.Timestamp {
	if x != nil {
		return x.LastHeartbeatTime
	}
	return nil
}

type NodesResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Nodes map[string]*Node `protobuf:"bytes,1,rep,name=nodes,proto3" json:"nodes,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *NodesResp) Reset() {
	*x = NodesResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cluster_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodesResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodesResp) ProtoMessage() {}

func (x *NodesResp) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodesResp.ProtoReflect.Descriptor instead.
func (*NodesResp) Descriptor() ([]byte, []int) {
	return file_cluster_proto_rawDescGZIP(), []int{1}
}

func (x *NodesResp) GetNodes() map[string]*Node {
	if x != nil {
		return x.Nodes
	}
	return nil
}

var File_cluster_proto protoreflect.FileDescriptor

var file_cluster_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x08, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xfa, 0x02, 0x0a, 0x04, 0x4e, 0x6f, 0x64, 0x65,
	0x12, 0x23, 0x0a, 0x0d, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x6d, 0x61, 0x69,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x4d, 0x61, 0x69, 0x6e, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x12, 0x23, 0x0a, 0x0d, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x67, 0x72,
	0x70, 0x63, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69,
	0x6e, 0x74, 0x47, 0x72, 0x70, 0x63, 0x12, 0x23, 0x0a, 0x0d, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69,
	0x6e, 0x74, 0x5f, 0x68, 0x74, 0x74, 0x70, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x65,
	0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x48, 0x74, 0x74, 0x70, 0x12, 0x14, 0x0a, 0x05, 0x71,
	0x75, 0x65, 0x75, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x71, 0x75, 0x65, 0x75,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6d, 0x6f, 0x64, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12,
	0x16, 0x0a, 0x06, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x18, 0x08, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x06, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x12, 0x3f, 0x0a, 0x0d, 0x72, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x65, 0x72, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0c, 0x72, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x65, 0x72, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x4a, 0x0a, 0x13, 0x6c, 0x61, 0x73, 0x74,
	0x5f, 0x68, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x52, 0x11, 0x6c, 0x61, 0x73, 0x74, 0x48, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74,
	0x54, 0x69, 0x6d, 0x65, 0x22, 0x8b, 0x01, 0x0a, 0x09, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x12, 0x34, 0x0a, 0x05, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x1e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2e, 0x4e, 0x6f, 0x64,
	0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x05, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x1a, 0x48, 0x0a, 0x0a, 0x4e, 0x6f, 0x64, 0x65,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x24, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x73, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x32, 0x44, 0x0a, 0x07, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x12, 0x39, 0x0a,
	0x08, 0x47, 0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x1a, 0x13, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2e, 0x4e, 0x6f, 0x64,
	0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x42, 0x0d, 0x5a, 0x0b, 0x2e, 0x2f, 0x3b, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cluster_proto_rawDescOnce sync.Once
	file_cluster_proto_rawDescData = file_cluster_proto_rawDesc
)

func file_cluster_proto_rawDescGZIP() []byte {
	file_cluster_proto_rawDescOnce.Do(func() {
		file_cluster_proto_rawDescData = protoimpl.X.CompressGZIP(file_cluster_proto_rawDescData)
	})
	return file_cluster_proto_rawDescData
}

var file_cluster_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_cluster_proto_goTypes = []any{
	(*Node)(nil),                  // 0: services.Node
	(*NodesResp)(nil),             // 1: services.NodesResp
	nil,                           // 2: services.NodesResp.NodesEntry
	(*timestamppb.Timestamp)(nil), // 3: google.protobuf.Timestamp
	(*emptypb.Empty)(nil),         // 4: google.protobuf.Empty
}
var file_cluster_proto_depIdxs = []int32{
	3, // 0: services.Node.register_time:type_name -> google.protobuf.Timestamp
	3, // 1: services.Node.last_heartbeat_time:type_name -> google.protobuf.Timestamp
	2, // 2: services.NodesResp.nodes:type_name -> services.NodesResp.NodesEntry
	0, // 3: services.NodesResp.NodesEntry.value:type_name -> services.Node
	4, // 4: services.Cluster.GetNodes:input_type -> google.protobuf.Empty
	1, // 5: services.Cluster.GetNodes:output_type -> services.NodesResp
	5, // [5:6] is the sub-list for method output_type
	4, // [4:5] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_cluster_proto_init() }
func file_cluster_proto_init() {
	if File_cluster_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cluster_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Node); i {
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
		file_cluster_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*NodesResp); i {
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
			RawDescriptor: file_cluster_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_cluster_proto_goTypes,
		DependencyIndexes: file_cluster_proto_depIdxs,
		MessageInfos:      file_cluster_proto_msgTypes,
	}.Build()
	File_cluster_proto = out.File
	file_cluster_proto_rawDesc = nil
	file_cluster_proto_goTypes = nil
	file_cluster_proto_depIdxs = nil
}
