# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: broker.proto
# Protobuf Python Version: 5.26.1
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import empty_pb2 as google_dot_protobuf_dot_empty__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x0c\x62roker.proto\x12\x08services\x1a\x1bgoogle/protobuf/empty.proto\"C\n\x05Queue\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\x0c\n\x04type\x18\x02 \x01(\t\x12\r\n\x05\x63ount\x18\x03 \x01(\x03\x12\x0f\n\x07workers\x18\x04 \x01(\x05\"-\n\nQueuesResp\x12\x1f\n\x06queues\x18\x01 \x03(\x0b\x32\x0f.services.Queue2E\n\x06\x42roker\x12;\n\tGetQueues\x12\x16.google.protobuf.Empty\x1a\x14.services.QueuesResp\"\x00\x42\rZ\x0b./;servicesb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'broker_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z\013./;services'
  _globals['_QUEUE']._serialized_start=55
  _globals['_QUEUE']._serialized_end=122
  _globals['_QUEUESRESP']._serialized_start=124
  _globals['_QUEUESRESP']._serialized_end=169
  _globals['_BROKER']._serialized_start=171
  _globals['_BROKER']._serialized_end=240
# @@protoc_insertion_point(module_scope)
