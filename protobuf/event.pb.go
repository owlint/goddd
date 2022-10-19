// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.21.7
// source: protobuf/event.proto

package protobuf

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

type EventSourcingEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Timestamp int64  `protobuf:"varint,1,opt,name=Timestamp,proto3" json:"Timestamp,omitempty"`
	ObjectID  string `protobuf:"bytes,2,opt,name=ObjectID,proto3" json:"ObjectID,omitempty"`
	Name      string `protobuf:"bytes,3,opt,name=Name,proto3" json:"Name,omitempty"`
	Payload   string `protobuf:"bytes,4,opt,name=Payload,proto3" json:"Payload,omitempty"`
	Version   int32  `protobuf:"varint,6,opt,name=Version,proto3" json:"Version,omitempty"`
}

func (x *EventSourcingEvent) Reset() {
	*x = EventSourcingEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protobuf_event_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventSourcingEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventSourcingEvent) ProtoMessage() {}

func (x *EventSourcingEvent) ProtoReflect() protoreflect.Message {
	mi := &file_protobuf_event_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventSourcingEvent.ProtoReflect.Descriptor instead.
func (*EventSourcingEvent) Descriptor() ([]byte, []int) {
	return file_protobuf_event_proto_rawDescGZIP(), []int{0}
}

func (x *EventSourcingEvent) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *EventSourcingEvent) GetObjectID() string {
	if x != nil {
		return x.ObjectID
	}
	return ""
}

func (x *EventSourcingEvent) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *EventSourcingEvent) GetPayload() string {
	if x != nil {
		return x.Payload
	}
	return ""
}

func (x *EventSourcingEvent) GetVersion() int32 {
	if x != nil {
		return x.Version
	}
	return 0
}

var File_protobuf_event_proto protoreflect.FileDescriptor

var file_protobuf_event_proto_rawDesc = []byte{
	0x0a, 0x14, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x22, 0x9c, 0x01, 0x0a, 0x12, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x69,
	0x6e, 0x67, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x1a, 0x0a, 0x08, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49,
	0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49,
	0x44, 0x12, 0x12, 0x0a, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12,
	0x18, 0x0a, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x4a, 0x04, 0x08, 0x05, 0x10, 0x06, 0x42,
	0x22, 0x5a, 0x20, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6f, 0x77,
	0x6c, 0x69, 0x6e, 0x74, 0x2f, 0x67, 0x6f, 0x64, 0x64, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protobuf_event_proto_rawDescOnce sync.Once
	file_protobuf_event_proto_rawDescData = file_protobuf_event_proto_rawDesc
)

func file_protobuf_event_proto_rawDescGZIP() []byte {
	file_protobuf_event_proto_rawDescOnce.Do(func() {
		file_protobuf_event_proto_rawDescData = protoimpl.X.CompressGZIP(file_protobuf_event_proto_rawDescData)
	})
	return file_protobuf_event_proto_rawDescData
}

var file_protobuf_event_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protobuf_event_proto_goTypes = []interface{}{
	(*EventSourcingEvent)(nil), // 0: protobuf.EventSourcingEvent
}
var file_protobuf_event_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_protobuf_event_proto_init() }
func file_protobuf_event_proto_init() {
	if File_protobuf_event_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protobuf_event_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventSourcingEvent); i {
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
			RawDescriptor: file_protobuf_event_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protobuf_event_proto_goTypes,
		DependencyIndexes: file_protobuf_event_proto_depIdxs,
		MessageInfos:      file_protobuf_event_proto_msgTypes,
	}.Build()
	File_protobuf_event_proto = out.File
	file_protobuf_event_proto_rawDesc = nil
	file_protobuf_event_proto_goTypes = nil
	file_protobuf_event_proto_depIdxs = nil
}
