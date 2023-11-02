from google.protobuf import empty_pb2 as _empty_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class JobId(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    def __init__(self, id: _Optional[str] = ...) -> None: ...

class Job(_message.Message):
    __slots__ = ["id", "name", "type", "start_at", "end_at", "interval", "cron_expr", "timezone", "func_name", "args", "timeout", "queues", "last_run_time", "next_run_time", "status", "scheduled"]
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    START_AT_FIELD_NUMBER: _ClassVar[int]
    END_AT_FIELD_NUMBER: _ClassVar[int]
    INTERVAL_FIELD_NUMBER: _ClassVar[int]
    CRON_EXPR_FIELD_NUMBER: _ClassVar[int]
    TIMEZONE_FIELD_NUMBER: _ClassVar[int]
    FUNC_NAME_FIELD_NUMBER: _ClassVar[int]
    ARGS_FIELD_NUMBER: _ClassVar[int]
    TIMEOUT_FIELD_NUMBER: _ClassVar[int]
    QUEUES_FIELD_NUMBER: _ClassVar[int]
    LAST_RUN_TIME_FIELD_NUMBER: _ClassVar[int]
    NEXT_RUN_TIME_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    SCHEDULED_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    type: str
    start_at: str
    end_at: str
    interval: str
    cron_expr: str
    timezone: str
    func_name: str
    args: _struct_pb2.Struct
    timeout: str
    queues: _containers.RepeatedScalarFieldContainer[str]
    last_run_time: _timestamp_pb2.Timestamp
    next_run_time: _timestamp_pb2.Timestamp
    status: str
    scheduled: bool
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., type: _Optional[str] = ..., start_at: _Optional[str] = ..., end_at: _Optional[str] = ..., interval: _Optional[str] = ..., cron_expr: _Optional[str] = ..., timezone: _Optional[str] = ..., func_name: _Optional[str] = ..., args: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., timeout: _Optional[str] = ..., queues: _Optional[_Iterable[str]] = ..., last_run_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., next_run_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., status: _Optional[str] = ..., scheduled: bool = ...) -> None: ...

class Jobs(_message.Message):
    __slots__ = ["Jobs"]
    JOBS_FIELD_NUMBER: _ClassVar[int]
    Jobs: _containers.RepeatedCompositeFieldContainer[Job]
    def __init__(self, Jobs: _Optional[_Iterable[_Union[Job, _Mapping]]] = ...) -> None: ...
