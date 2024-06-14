from google.protobuf import empty_pb2 as _empty_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Queue(_message.Message):
    __slots__ = ("name", "type", "count", "workers")
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    COUNT_FIELD_NUMBER: _ClassVar[int]
    WORKERS_FIELD_NUMBER: _ClassVar[int]
    name: str
    type: str
    count: int
    workers: int
    def __init__(self, name: _Optional[str] = ..., type: _Optional[str] = ..., count: _Optional[int] = ..., workers: _Optional[int] = ...) -> None: ...

class QueuesResp(_message.Message):
    __slots__ = ("queues",)
    QUEUES_FIELD_NUMBER: _ClassVar[int]
    queues: _containers.RepeatedCompositeFieldContainer[Queue]
    def __init__(self, queues: _Optional[_Iterable[_Union[Queue, _Mapping]]] = ...) -> None: ...
