// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v6.30.1
// source: predictions_service/predictions_service.proto

package predictions

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

type PredictRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UID         int64   `protobuf:"varint,1,opt,name=UID,proto3" json:"UID,omitempty"`
	PlannedTime float64 `protobuf:"fixed64,2,opt,name=PlannedTime,proto3" json:"PlannedTime,omitempty"`
}

func (x *PredictRequest) Reset() {
	*x = PredictRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_predictions_service_predictions_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PredictRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PredictRequest) ProtoMessage() {}

func (x *PredictRequest) ProtoReflect() protoreflect.Message {
	mi := &file_predictions_service_predictions_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PredictRequest.ProtoReflect.Descriptor instead.
func (*PredictRequest) Descriptor() ([]byte, []int) {
	return file_predictions_service_predictions_service_proto_rawDescGZIP(), []int{0}
}

func (x *PredictRequest) GetUID() int64 {
	if x != nil {
		return x.UID
	}
	return 0
}

func (x *PredictRequest) GetPlannedTime() float64 {
	if x != nil {
		return x.PlannedTime
	}
	return 0
}

type PredictResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ActualTime float64 `protobuf:"fixed64,1,opt,name=ActualTime,proto3" json:"ActualTime,omitempty"`
	Status     string  `protobuf:"bytes,2,opt,name=Status,proto3" json:"Status,omitempty"`
}

func (x *PredictResponse) Reset() {
	*x = PredictResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_predictions_service_predictions_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PredictResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PredictResponse) ProtoMessage() {}

func (x *PredictResponse) ProtoReflect() protoreflect.Message {
	mi := &file_predictions_service_predictions_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PredictResponse.ProtoReflect.Descriptor instead.
func (*PredictResponse) Descriptor() ([]byte, []int) {
	return file_predictions_service_predictions_service_proto_rawDescGZIP(), []int{1}
}

func (x *PredictResponse) GetActualTime() float64 {
	if x != nil {
		return x.ActualTime
	}
	return 0
}

func (x *PredictResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

type UserWithTime struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ID   int64   `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	UID  int64   `protobuf:"varint,2,opt,name=UID,proto3" json:"UID,omitempty"`
	Time float64 `protobuf:"fixed64,3,opt,name=Time,proto3" json:"Time,omitempty"`
}

func (x *UserWithTime) Reset() {
	*x = UserWithTime{}
	if protoimpl.UnsafeEnabled {
		mi := &file_predictions_service_predictions_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserWithTime) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserWithTime) ProtoMessage() {}

func (x *UserWithTime) ProtoReflect() protoreflect.Message {
	mi := &file_predictions_service_predictions_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserWithTime.ProtoReflect.Descriptor instead.
func (*UserWithTime) Descriptor() ([]byte, []int) {
	return file_predictions_service_predictions_service_proto_rawDescGZIP(), []int{2}
}

func (x *UserWithTime) GetID() int64 {
	if x != nil {
		return x.ID
	}
	return 0
}

func (x *UserWithTime) GetUID() int64 {
	if x != nil {
		return x.UID
	}
	return 0
}

func (x *UserWithTime) GetTime() float64 {
	if x != nil {
		return x.Time
	}
	return 0
}

type PredictListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PlannedUserTime []*UserWithTime `protobuf:"bytes,1,rep,name=PlannedUserTime,proto3" json:"PlannedUserTime,omitempty"`
}

func (x *PredictListRequest) Reset() {
	*x = PredictListRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_predictions_service_predictions_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PredictListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PredictListRequest) ProtoMessage() {}

func (x *PredictListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_predictions_service_predictions_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PredictListRequest.ProtoReflect.Descriptor instead.
func (*PredictListRequest) Descriptor() ([]byte, []int) {
	return file_predictions_service_predictions_service_proto_rawDescGZIP(), []int{3}
}

func (x *PredictListRequest) GetPlannedUserTime() []*UserWithTime {
	if x != nil {
		return x.PlannedUserTime
	}
	return nil
}

type PredictListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PredictedUserTime []*UserWithTime `protobuf:"bytes,1,rep,name=PredictedUserTime,proto3" json:"PredictedUserTime,omitempty"`
	UnpredictedUIDs   []int64         `protobuf:"varint,2,rep,packed,name=UnpredictedUIDs,proto3" json:"UnpredictedUIDs,omitempty"` // users IDs which doesn't have tasks
}

func (x *PredictListResponse) Reset() {
	*x = PredictListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_predictions_service_predictions_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PredictListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PredictListResponse) ProtoMessage() {}

func (x *PredictListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_predictions_service_predictions_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PredictListResponse.ProtoReflect.Descriptor instead.
func (*PredictListResponse) Descriptor() ([]byte, []int) {
	return file_predictions_service_predictions_service_proto_rawDescGZIP(), []int{4}
}

func (x *PredictListResponse) GetPredictedUserTime() []*UserWithTime {
	if x != nil {
		return x.PredictedUserTime
	}
	return nil
}

func (x *PredictListResponse) GetUnpredictedUIDs() []int64 {
	if x != nil {
		return x.UnpredictedUIDs
	}
	return nil
}

var File_predictions_service_predictions_service_proto protoreflect.FileDescriptor

var file_predictions_service_predictions_service_proto_rawDesc = []byte{
	0x0a, 0x2d, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x5f, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x0b, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x44, 0x0a, 0x0e,
	0x50, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10,
	0x0a, 0x03, 0x55, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x55, 0x49, 0x44,
	0x12, 0x20, 0x0a, 0x0b, 0x50, 0x6c, 0x61, 0x6e, 0x6e, 0x65, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x0b, 0x50, 0x6c, 0x61, 0x6e, 0x6e, 0x65, 0x64, 0x54, 0x69,
	0x6d, 0x65, 0x22, 0x49, 0x0a, 0x0f, 0x50, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x54,
	0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x0a, 0x41, 0x63, 0x74, 0x75, 0x61,
	0x6c, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x44, 0x0a,
	0x0c, 0x55, 0x73, 0x65, 0x72, 0x57, 0x69, 0x74, 0x68, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x0e, 0x0a,
	0x02, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x49, 0x44, 0x12, 0x10, 0x0a,
	0x03, 0x55, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x55, 0x49, 0x44, 0x12,
	0x12, 0x0a, 0x04, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x52, 0x04, 0x54,
	0x69, 0x6d, 0x65, 0x22, 0x59, 0x0a, 0x12, 0x50, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x4c, 0x69,
	0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x43, 0x0a, 0x0f, 0x50, 0x6c, 0x61,
	0x6e, 0x6e, 0x65, 0x64, 0x55, 0x73, 0x65, 0x72, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x19, 0x2e, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x55, 0x73, 0x65, 0x72, 0x57, 0x69, 0x74, 0x68, 0x54, 0x69, 0x6d, 0x65, 0x52, 0x0f, 0x50,
	0x6c, 0x61, 0x6e, 0x6e, 0x65, 0x64, 0x55, 0x73, 0x65, 0x72, 0x54, 0x69, 0x6d, 0x65, 0x22, 0x88,
	0x01, 0x0a, 0x13, 0x50, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x47, 0x0a, 0x11, 0x50, 0x72, 0x65, 0x64, 0x69, 0x63,
	0x74, 0x65, 0x64, 0x55, 0x73, 0x65, 0x72, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x19, 0x2e, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x55, 0x73, 0x65, 0x72, 0x57, 0x69, 0x74, 0x68, 0x54, 0x69, 0x6d, 0x65, 0x52, 0x11, 0x50, 0x72,
	0x65, 0x64, 0x69, 0x63, 0x74, 0x65, 0x64, 0x55, 0x73, 0x65, 0x72, 0x54, 0x69, 0x6d, 0x65, 0x12,
	0x28, 0x0a, 0x0f, 0x55, 0x6e, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x65, 0x64, 0x55, 0x49,
	0x44, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x03, 0x52, 0x0f, 0x55, 0x6e, 0x70, 0x72, 0x65, 0x64,
	0x69, 0x63, 0x74, 0x65, 0x64, 0x55, 0x49, 0x44, 0x73, 0x32, 0xa5, 0x01, 0x0a, 0x0b, 0x50, 0x72,
	0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x44, 0x0a, 0x07, 0x50, 0x72, 0x65,
	0x64, 0x69, 0x63, 0x74, 0x12, 0x1b, 0x2e, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x50, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1c, 0x2e, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x50, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x50, 0x0a, 0x0b, 0x50, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x1f,
	0x2e, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x50, 0x72, 0x65,
	0x64, 0x69, 0x63, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x20, 0x2e, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x50, 0x72,
	0x65, 0x64, 0x69, 0x63, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x42, 0x20, 0x5a, 0x1e, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x5f, 0x73, 0x79, 0x73,
	0x74, 0x65, 0x6d, 0x2e, 0x61, 0x70, 0x69, 0x3b, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_predictions_service_predictions_service_proto_rawDescOnce sync.Once
	file_predictions_service_predictions_service_proto_rawDescData = file_predictions_service_predictions_service_proto_rawDesc
)

func file_predictions_service_predictions_service_proto_rawDescGZIP() []byte {
	file_predictions_service_predictions_service_proto_rawDescOnce.Do(func() {
		file_predictions_service_predictions_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_predictions_service_predictions_service_proto_rawDescData)
	})
	return file_predictions_service_predictions_service_proto_rawDescData
}

var file_predictions_service_predictions_service_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_predictions_service_predictions_service_proto_goTypes = []interface{}{
	(*PredictRequest)(nil),      // 0: predictions.PredictRequest
	(*PredictResponse)(nil),     // 1: predictions.PredictResponse
	(*UserWithTime)(nil),        // 2: predictions.UserWithTime
	(*PredictListRequest)(nil),  // 3: predictions.PredictListRequest
	(*PredictListResponse)(nil), // 4: predictions.PredictListResponse
}
var file_predictions_service_predictions_service_proto_depIdxs = []int32{
	2, // 0: predictions.PredictListRequest.PlannedUserTime:type_name -> predictions.UserWithTime
	2, // 1: predictions.PredictListResponse.PredictedUserTime:type_name -> predictions.UserWithTime
	0, // 2: predictions.Predictions.Predict:input_type -> predictions.PredictRequest
	3, // 3: predictions.Predictions.PredictList:input_type -> predictions.PredictListRequest
	1, // 4: predictions.Predictions.Predict:output_type -> predictions.PredictResponse
	4, // 5: predictions.Predictions.PredictList:output_type -> predictions.PredictListResponse
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_predictions_service_predictions_service_proto_init() }
func file_predictions_service_predictions_service_proto_init() {
	if File_predictions_service_predictions_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_predictions_service_predictions_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PredictRequest); i {
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
		file_predictions_service_predictions_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PredictResponse); i {
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
		file_predictions_service_predictions_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserWithTime); i {
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
		file_predictions_service_predictions_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PredictListRequest); i {
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
		file_predictions_service_predictions_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PredictListResponse); i {
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
			RawDescriptor: file_predictions_service_predictions_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_predictions_service_predictions_service_proto_goTypes,
		DependencyIndexes: file_predictions_service_predictions_service_proto_depIdxs,
		MessageInfos:      file_predictions_service_predictions_service_proto_msgTypes,
	}.Build()
	File_predictions_service_predictions_service_proto = out.File
	file_predictions_service_predictions_service_proto_rawDesc = nil
	file_predictions_service_predictions_service_proto_goTypes = nil
	file_predictions_service_predictions_service_proto_depIdxs = nil
}
