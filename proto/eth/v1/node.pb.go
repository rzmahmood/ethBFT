// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.23.3
// source: proto/eth/v1/node.proto

package v1

import (
	reflect "reflect"
	sync "sync"

	_ "github.com/golang/protobuf/protoc-gen-go/descriptor"
	_ "github.com/prysmaticlabs/prysm/v4/proto/eth/ext"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/descriptorpb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PeerDirection int32

const (
	PeerDirection_INBOUND  PeerDirection = 0
	PeerDirection_OUTBOUND PeerDirection = 1
)

// Enum value maps for PeerDirection.
var (
	PeerDirection_name = map[int32]string{
		0: "INBOUND",
		1: "OUTBOUND",
	}
	PeerDirection_value = map[string]int32{
		"INBOUND":  0,
		"OUTBOUND": 1,
	}
)

func (x PeerDirection) Enum() *PeerDirection {
	p := new(PeerDirection)
	*p = x
	return p
}

func (x PeerDirection) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PeerDirection) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_eth_v1_node_proto_enumTypes[0].Descriptor()
}

func (PeerDirection) Type() protoreflect.EnumType {
	return &file_proto_eth_v1_node_proto_enumTypes[0]
}

func (x PeerDirection) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PeerDirection.Descriptor instead.
func (PeerDirection) EnumDescriptor() ([]byte, []int) {
	return file_proto_eth_v1_node_proto_rawDescGZIP(), []int{0}
}

type ConnectionState int32

const (
	ConnectionState_DISCONNECTED  ConnectionState = 0
	ConnectionState_CONNECTING    ConnectionState = 1
	ConnectionState_CONNECTED     ConnectionState = 2
	ConnectionState_DISCONNECTING ConnectionState = 3
)

// Enum value maps for ConnectionState.
var (
	ConnectionState_name = map[int32]string{
		0: "DISCONNECTED",
		1: "CONNECTING",
		2: "CONNECTED",
		3: "DISCONNECTING",
	}
	ConnectionState_value = map[string]int32{
		"DISCONNECTED":  0,
		"CONNECTING":    1,
		"CONNECTED":     2,
		"DISCONNECTING": 3,
	}
)

func (x ConnectionState) Enum() *ConnectionState {
	p := new(ConnectionState)
	*p = x
	return p
}

func (x ConnectionState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ConnectionState) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_eth_v1_node_proto_enumTypes[1].Descriptor()
}

func (ConnectionState) Type() protoreflect.EnumType {
	return &file_proto_eth_v1_node_proto_enumTypes[1]
}

func (x ConnectionState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ConnectionState.Descriptor instead.
func (ConnectionState) EnumDescriptor() ([]byte, []int) {
	return file_proto_eth_v1_node_proto_rawDescGZIP(), []int{1}
}

type Peer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PeerId             string          `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	Enr                string          `protobuf:"bytes,2,opt,name=enr,proto3" json:"enr,omitempty"`
	LastSeenP2PAddress string          `protobuf:"bytes,3,opt,name=last_seen_p2p_address,json=lastSeenP2pAddress,proto3" json:"last_seen_p2p_address,omitempty"`
	State              ConnectionState `protobuf:"varint,4,opt,name=state,proto3,enum=ethereum.eth.v1.ConnectionState" json:"state,omitempty"`
	Direction          PeerDirection   `protobuf:"varint,5,opt,name=direction,proto3,enum=ethereum.eth.v1.PeerDirection" json:"direction,omitempty"`
}

func (x *Peer) Reset() {
	*x = Peer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_eth_v1_node_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Peer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Peer) ProtoMessage() {}

func (x *Peer) ProtoReflect() protoreflect.Message {
	mi := &file_proto_eth_v1_node_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Peer.ProtoReflect.Descriptor instead.
func (*Peer) Descriptor() ([]byte, []int) {
	return file_proto_eth_v1_node_proto_rawDescGZIP(), []int{0}
}

func (x *Peer) GetPeerId() string {
	if x != nil {
		return x.PeerId
	}
	return ""
}

func (x *Peer) GetEnr() string {
	if x != nil {
		return x.Enr
	}
	return ""
}

func (x *Peer) GetLastSeenP2PAddress() string {
	if x != nil {
		return x.LastSeenP2PAddress
	}
	return ""
}

func (x *Peer) GetState() ConnectionState {
	if x != nil {
		return x.State
	}
	return ConnectionState_DISCONNECTED
}

func (x *Peer) GetDirection() PeerDirection {
	if x != nil {
		return x.Direction
	}
	return PeerDirection_INBOUND
}

var File_proto_eth_v1_node_proto protoreflect.FileDescriptor

var file_proto_eth_v1_node_proto_rawDesc = []byte{
	0x0a, 0x17, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x74, 0x68, 0x2f, 0x76, 0x31, 0x2f, 0x6e,
	0x6f, 0x64, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x65, 0x74, 0x68, 0x65, 0x72,
	0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x74, 0x68, 0x2f, 0x65, 0x78, 0x74, 0x2f, 0x6f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xda, 0x01, 0x0a, 0x04, 0x50, 0x65,
	0x65, 0x72, 0x12, 0x17, 0x0a, 0x07, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x65, 0x65, 0x72, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x65,
	0x6e, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x65, 0x6e, 0x72, 0x12, 0x31, 0x0a,
	0x15, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x73, 0x65, 0x65, 0x6e, 0x5f, 0x70, 0x32, 0x70, 0x5f, 0x61,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x6c, 0x61,
	0x73, 0x74, 0x53, 0x65, 0x65, 0x6e, 0x50, 0x32, 0x70, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x12, 0x36, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x20, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76,
	0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x3c, 0x0a, 0x09, 0x64, 0x69, 0x72, 0x65,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1e, 0x2e, 0x65, 0x74,
	0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x65,
	0x65, 0x72, 0x44, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x09, 0x64, 0x69, 0x72,
	0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2a, 0x2a, 0x0a, 0x0d, 0x50, 0x65, 0x65, 0x72, 0x44, 0x69,
	0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x0b, 0x0a, 0x07, 0x49, 0x4e, 0x42, 0x4f, 0x55,
	0x4e, 0x44, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x4f, 0x55, 0x54, 0x42, 0x4f, 0x55, 0x4e, 0x44,
	0x10, 0x01, 0x2a, 0x55, 0x0a, 0x0f, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x10, 0x0a, 0x0c, 0x44, 0x49, 0x53, 0x43, 0x4f, 0x4e, 0x4e,
	0x45, 0x43, 0x54, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a, 0x43, 0x4f, 0x4e, 0x4e, 0x45,
	0x43, 0x54, 0x49, 0x4e, 0x47, 0x10, 0x01, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x4f, 0x4e, 0x4e, 0x45,
	0x43, 0x54, 0x45, 0x44, 0x10, 0x02, 0x12, 0x11, 0x0a, 0x0d, 0x44, 0x49, 0x53, 0x43, 0x4f, 0x4e,
	0x4e, 0x45, 0x43, 0x54, 0x49, 0x4e, 0x47, 0x10, 0x03, 0x42, 0x7c, 0x0a, 0x13, 0x6f, 0x72, 0x67,
	0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x31,
	0x42, 0x0f, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x4e, 0x6f, 0x64, 0x65, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x50, 0x01, 0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x70, 0x72, 0x79, 0x73, 0x6d, 0x61, 0x74, 0x69, 0x63, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x70, 0x72,
	0x79, 0x73, 0x6d, 0x2f, 0x76, 0x34, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x74, 0x68,
	0x2f, 0x76, 0x31, 0xaa, 0x02, 0x0f, 0x45, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x45,
	0x74, 0x68, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x0f, 0x45, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d,
	0x5c, 0x45, 0x74, 0x68, 0x5c, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_eth_v1_node_proto_rawDescOnce sync.Once
	file_proto_eth_v1_node_proto_rawDescData = file_proto_eth_v1_node_proto_rawDesc
)

func file_proto_eth_v1_node_proto_rawDescGZIP() []byte {
	file_proto_eth_v1_node_proto_rawDescOnce.Do(func() {
		file_proto_eth_v1_node_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_eth_v1_node_proto_rawDescData)
	})
	return file_proto_eth_v1_node_proto_rawDescData
}

var file_proto_eth_v1_node_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_proto_eth_v1_node_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_proto_eth_v1_node_proto_goTypes = []interface{}{
	(PeerDirection)(0),   // 0: ethereum.eth.v1.PeerDirection
	(ConnectionState)(0), // 1: ethereum.eth.v1.ConnectionState
	(*Peer)(nil),         // 2: ethereum.eth.v1.Peer
}
var file_proto_eth_v1_node_proto_depIdxs = []int32{
	1, // 0: ethereum.eth.v1.Peer.state:type_name -> ethereum.eth.v1.ConnectionState
	0, // 1: ethereum.eth.v1.Peer.direction:type_name -> ethereum.eth.v1.PeerDirection
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_proto_eth_v1_node_proto_init() }
func file_proto_eth_v1_node_proto_init() {
	if File_proto_eth_v1_node_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_eth_v1_node_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Peer); i {
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
			RawDescriptor: file_proto_eth_v1_node_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_eth_v1_node_proto_goTypes,
		DependencyIndexes: file_proto_eth_v1_node_proto_depIdxs,
		EnumInfos:         file_proto_eth_v1_node_proto_enumTypes,
		MessageInfos:      file_proto_eth_v1_node_proto_msgTypes,
	}.Build()
	File_proto_eth_v1_node_proto = out.File
	file_proto_eth_v1_node_proto_rawDesc = nil
	file_proto_eth_v1_node_proto_goTypes = nil
	file_proto_eth_v1_node_proto_depIdxs = nil
}
