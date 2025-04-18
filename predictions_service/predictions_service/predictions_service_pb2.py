# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: predictions_service/predictions_service.proto
# Protobuf Python Version: 5.28.1
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    5,
    28,
    1,
    '',
    'predictions_service/predictions_service.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n-predictions_service/predictions_service.proto\x12\x0bpredictions\"2\n\x0ePredictRequest\x12\x0b\n\x03UID\x18\x01 \x01(\x03\x12\x13\n\x0bPlannedTime\x18\x02 \x01(\x01\"5\n\x0fPredictResponse\x12\x12\n\nActualTime\x18\x01 \x01(\x01\x12\x0e\n\x06Status\x18\x02 \x01(\t\"5\n\x0cUserWithTime\x12\n\n\x02ID\x18\x01 \x01(\x03\x12\x0b\n\x03UID\x18\x02 \x01(\x03\x12\x0c\n\x04Time\x18\x03 \x01(\x01\"H\n\x12PredictListRequest\x12\x32\n\x0fPlannedUserTime\x18\x01 \x03(\x0b\x32\x19.predictions.UserWithTime\"d\n\x13PredictListResponse\x12\x34\n\x11PredictedUserTime\x18\x01 \x03(\x0b\x32\x19.predictions.UserWithTime\x12\x17\n\x0fUnpredictedUIDs\x18\x02 \x03(\x03\x32\xa5\x01\n\x0bPredictions\x12\x44\n\x07Predict\x12\x1b.predictions.PredictRequest\x1a\x1c.predictions.PredictResponse\x12P\n\x0bPredictList\x12\x1f.predictions.PredictListRequest\x1a .predictions.PredictListResponseB Z\x1e\x63ontrol_system.api;predictionsb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'predictions_service.predictions_service_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z\036control_system.api;predictions'
  _globals['_PREDICTREQUEST']._serialized_start=62
  _globals['_PREDICTREQUEST']._serialized_end=112
  _globals['_PREDICTRESPONSE']._serialized_start=114
  _globals['_PREDICTRESPONSE']._serialized_end=167
  _globals['_USERWITHTIME']._serialized_start=169
  _globals['_USERWITHTIME']._serialized_end=222
  _globals['_PREDICTLISTREQUEST']._serialized_start=224
  _globals['_PREDICTLISTREQUEST']._serialized_end=296
  _globals['_PREDICTLISTRESPONSE']._serialized_start=298
  _globals['_PREDICTLISTRESPONSE']._serialized_end=398
  _globals['_PREDICTIONS']._serialized_start=401
  _globals['_PREDICTIONS']._serialized_end=566
# @@protoc_insertion_point(module_scope)
