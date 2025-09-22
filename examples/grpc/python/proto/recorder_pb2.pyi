import datetime

from google.protobuf import empty_pb2 as _empty_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
import scheduler_pb2 as _scheduler_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RecordsReq(_message.Message):
    __slots__ = ("job_id", "page", "page_size")
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    PAGE_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    job_id: str
    page: int
    page_size: int
    def __init__(self, job_id: _Optional[str] = ..., page: _Optional[int] = ..., page_size: _Optional[int] = ...) -> None: ...

class RecordsAllReq(_message.Message):
    __slots__ = ("page", "page_size")
    PAGE_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    page: int
    page_size: int
    def __init__(self, page: _Optional[int] = ..., page_size: _Optional[int] = ...) -> None: ...

class Record(_message.Message):
    __slots__ = ("id", "job_id", "job_name", "status", "result", "start_at", "end_at")
    ID_FIELD_NUMBER: _ClassVar[int]
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    JOB_NAME_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    RESULT_FIELD_NUMBER: _ClassVar[int]
    START_AT_FIELD_NUMBER: _ClassVar[int]
    END_AT_FIELD_NUMBER: _ClassVar[int]
    id: int
    job_id: str
    job_name: str
    status: str
    result: str
    start_at: _timestamp_pb2.Timestamp
    end_at: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[int] = ..., job_id: _Optional[str] = ..., job_name: _Optional[str] = ..., status: _Optional[str] = ..., result: _Optional[str] = ..., start_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., end_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class RecordsResp(_message.Message):
    __slots__ = ("records", "page", "page_size", "total")
    RECORDS_FIELD_NUMBER: _ClassVar[int]
    PAGE_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    TOTAL_FIELD_NUMBER: _ClassVar[int]
    records: _containers.RepeatedCompositeFieldContainer[Record]
    page: int
    page_size: int
    total: int
    def __init__(self, records: _Optional[_Iterable[_Union[Record, _Mapping]]] = ..., page: _Optional[int] = ..., page_size: _Optional[int] = ..., total: _Optional[int] = ...) -> None: ...
