import datetime

from google.protobuf import empty_pb2 as _empty_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Node(_message.Message):
    __slots__ = ("endpoint_main", "endpoint", "endpoint_grpc", "endpoint_http", "queue", "mode", "version", "health", "register_time", "last_heartbeat_time")
    ENDPOINT_MAIN_FIELD_NUMBER: _ClassVar[int]
    ENDPOINT_FIELD_NUMBER: _ClassVar[int]
    ENDPOINT_GRPC_FIELD_NUMBER: _ClassVar[int]
    ENDPOINT_HTTP_FIELD_NUMBER: _ClassVar[int]
    QUEUE_FIELD_NUMBER: _ClassVar[int]
    MODE_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    HEALTH_FIELD_NUMBER: _ClassVar[int]
    REGISTER_TIME_FIELD_NUMBER: _ClassVar[int]
    LAST_HEARTBEAT_TIME_FIELD_NUMBER: _ClassVar[int]
    endpoint_main: str
    endpoint: str
    endpoint_grpc: str
    endpoint_http: str
    queue: str
    mode: str
    version: str
    health: bool
    register_time: _timestamp_pb2.Timestamp
    last_heartbeat_time: _timestamp_pb2.Timestamp
    def __init__(self, endpoint_main: _Optional[str] = ..., endpoint: _Optional[str] = ..., endpoint_grpc: _Optional[str] = ..., endpoint_http: _Optional[str] = ..., queue: _Optional[str] = ..., mode: _Optional[str] = ..., version: _Optional[str] = ..., health: bool = ..., register_time: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., last_heartbeat_time: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class NodesResp(_message.Message):
    __slots__ = ("nodes",)
    class NodesEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: Node
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[Node, _Mapping]] = ...) -> None: ...
    NODES_FIELD_NUMBER: _ClassVar[int]
    nodes: _containers.MessageMap[str, Node]
    def __init__(self, nodes: _Optional[_Mapping[str, Node]] = ...) -> None: ...
