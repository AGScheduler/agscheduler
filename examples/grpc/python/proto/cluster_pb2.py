# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: cluster.proto
# Protobuf Python Version: 5.26.1
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import empty_pb2 as google_dot_protobuf_dot_empty__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\rcluster.proto\x12\x08services\x1a\x1bgoogle/protobuf/empty.proto\x1a\x1fgoogle/protobuf/timestamp.proto\"\x87\x02\n\x04Node\x12\x15\n\rendpoint_main\x18\x01 \x01(\t\x12\x10\n\x08\x65ndpoint\x18\x02 \x01(\t\x12\x15\n\rendpoint_grpc\x18\x03 \x01(\t\x12\x15\n\rendpoint_http\x18\x04 \x01(\t\x12\r\n\x05queue\x18\x05 \x01(\t\x12\x0c\n\x04mode\x18\x06 \x01(\t\x12\x0f\n\x07version\x18\x07 \x01(\t\x12\x0e\n\x06health\x18\x08 \x01(\x08\x12\x31\n\rregister_time\x18\t \x01(\x0b\x32\x1a.google.protobuf.Timestamp\x12\x37\n\x13last_heartbeat_time\x18\n \x01(\x0b\x32\x1a.google.protobuf.Timestamp\"x\n\tNodesResp\x12-\n\x05nodes\x18\x01 \x03(\x0b\x32\x1e.services.NodesResp.NodesEntry\x1a<\n\nNodesEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\x1d\n\x05value\x18\x02 \x01(\x0b\x32\x0e.services.Node:\x02\x38\x01\x32\x44\n\x07\x43luster\x12\x39\n\x08GetNodes\x12\x16.google.protobuf.Empty\x1a\x13.services.NodesResp\"\x00\x42\rZ\x0b./;servicesb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'cluster_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z\013./;services'
  _globals['_NODESRESP_NODESENTRY']._loaded_options = None
  _globals['_NODESRESP_NODESENTRY']._serialized_options = b'8\001'
  _globals['_NODE']._serialized_start=90
  _globals['_NODE']._serialized_end=353
  _globals['_NODESRESP']._serialized_start=355
  _globals['_NODESRESP']._serialized_end=475
  _globals['_NODESRESP_NODESENTRY']._serialized_start=415
  _globals['_NODESRESP_NODESENTRY']._serialized_end=475
  _globals['_CLUSTER']._serialized_start=477
  _globals['_CLUSTER']._serialized_end=545
# @@protoc_insertion_point(module_scope)
