from google.protobuf import empty_pb2 as _empty_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Info(_message.Message):
    __slots__ = ("info",)
    INFO_FIELD_NUMBER: _ClassVar[int]
    info: _struct_pb2.Struct
    def __init__(self, info: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class Func(_message.Message):
    __slots__ = ("name", "info")
    NAME_FIELD_NUMBER: _ClassVar[int]
    INFO_FIELD_NUMBER: _ClassVar[int]
    name: str
    info: str
    def __init__(self, name: _Optional[str] = ..., info: _Optional[str] = ...) -> None: ...

class Funcs(_message.Message):
    __slots__ = ("funcs",)
    FUNCS_FIELD_NUMBER: _ClassVar[int]
    funcs: _containers.RepeatedCompositeFieldContainer[Func]
    def __init__(self, funcs: _Optional[_Iterable[_Union[Func, _Mapping]]] = ...) -> None: ...
