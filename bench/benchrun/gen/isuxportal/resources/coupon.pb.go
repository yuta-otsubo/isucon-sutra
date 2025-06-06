// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: isuxportal/resources/coupon.proto

package resources

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Coupon struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	TeamId        int64                  `protobuf:"varint,2,opt,name=team_id,json=teamId,proto3" json:"team_id,omitempty"`
	Code          []string               `protobuf:"bytes,3,rep,name=code,proto3" json:"code,omitempty"`
	Activate      bool                   `protobuf:"varint,4,opt,name=activate,proto3" json:"activate,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Coupon) Reset() {
	*x = Coupon{}
	mi := &file_isuxportal_resources_coupon_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Coupon) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Coupon) ProtoMessage() {}

func (x *Coupon) ProtoReflect() protoreflect.Message {
	mi := &file_isuxportal_resources_coupon_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Coupon.ProtoReflect.Descriptor instead.
func (*Coupon) Descriptor() ([]byte, []int) {
	return file_isuxportal_resources_coupon_proto_rawDescGZIP(), []int{0}
}

func (x *Coupon) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Coupon) GetTeamId() int64 {
	if x != nil {
		return x.TeamId
	}
	return 0
}

func (x *Coupon) GetCode() []string {
	if x != nil {
		return x.Code
	}
	return nil
}

func (x *Coupon) GetActivate() bool {
	if x != nil {
		return x.Activate
	}
	return false
}

var File_isuxportal_resources_coupon_proto protoreflect.FileDescriptor

const file_isuxportal_resources_coupon_proto_rawDesc = "" +
	"\n" +
	"!isuxportal/resources/coupon.proto\x12\x1aisuxportal.proto.resources\"a\n" +
	"\x06Coupon\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\x12\x17\n" +
	"\ateam_id\x18\x02 \x01(\x03R\x06teamId\x12\x12\n" +
	"\x04code\x18\x03 \x03(\tR\x04code\x12\x1a\n" +
	"\bactivate\x18\x04 \x01(\bR\bactivateB\x84\x02\n" +
	"\x1ecom.isuxportal.proto.resourcesB\vCouponProtoP\x01ZKgithub.com/yuta-otsubo/isucon-sutra/bench/benchrun/gen/isuxportal/resources\xa2\x02\x03IPR\xaa\x02\x1aIsuxportal.Proto.Resources\xca\x02\x1aIsuxportal\\Proto\\Resources\xe2\x02&Isuxportal\\Proto\\Resources\\GPBMetadata\xea\x02\x1cIsuxportal::Proto::Resourcesb\x06proto3"

var (
	file_isuxportal_resources_coupon_proto_rawDescOnce sync.Once
	file_isuxportal_resources_coupon_proto_rawDescData []byte
)

func file_isuxportal_resources_coupon_proto_rawDescGZIP() []byte {
	file_isuxportal_resources_coupon_proto_rawDescOnce.Do(func() {
		file_isuxportal_resources_coupon_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_isuxportal_resources_coupon_proto_rawDesc), len(file_isuxportal_resources_coupon_proto_rawDesc)))
	})
	return file_isuxportal_resources_coupon_proto_rawDescData
}

var file_isuxportal_resources_coupon_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_isuxportal_resources_coupon_proto_goTypes = []any{
	(*Coupon)(nil), // 0: isuxportal.proto.resources.Coupon
}
var file_isuxportal_resources_coupon_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_isuxportal_resources_coupon_proto_init() }
func file_isuxportal_resources_coupon_proto_init() {
	if File_isuxportal_resources_coupon_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_isuxportal_resources_coupon_proto_rawDesc), len(file_isuxportal_resources_coupon_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_isuxportal_resources_coupon_proto_goTypes,
		DependencyIndexes: file_isuxportal_resources_coupon_proto_depIdxs,
		MessageInfos:      file_isuxportal_resources_coupon_proto_msgTypes,
	}.Build()
	File_isuxportal_resources_coupon_proto = out.File
	file_isuxportal_resources_coupon_proto_goTypes = nil
	file_isuxportal_resources_coupon_proto_depIdxs = nil
}
