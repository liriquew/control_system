from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class PredictRequest(_message.Message):
    __slots__ = ("UID", "PlannedTime")
    UID_FIELD_NUMBER: _ClassVar[int]
    PLANNEDTIME_FIELD_NUMBER: _ClassVar[int]
    UID: int
    PlannedTime: float
    def __init__(self, UID: _Optional[int] = ..., PlannedTime: _Optional[float] = ...) -> None: ...

class PredictResponse(_message.Message):
    __slots__ = ("ActualTime", "Status")
    ACTUALTIME_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ActualTime: float
    Status: str
    def __init__(self, ActualTime: _Optional[float] = ..., Status: _Optional[str] = ...) -> None: ...

class UserWithTime(_message.Message):
    __slots__ = ("UID", "Time")
    UID_FIELD_NUMBER: _ClassVar[int]
    TIME_FIELD_NUMBER: _ClassVar[int]
    UID: int
    Time: float
    def __init__(self, UID: _Optional[int] = ..., Time: _Optional[float] = ...) -> None: ...

class PredictListRequest(_message.Message):
    __slots__ = ("PlannedUserTime",)
    PLANNEDUSERTIME_FIELD_NUMBER: _ClassVar[int]
    PlannedUserTime: _containers.RepeatedCompositeFieldContainer[UserWithTime]
    def __init__(self, PlannedUserTime: _Optional[_Iterable[_Union[UserWithTime, _Mapping]]] = ...) -> None: ...

class PredictListResponse(_message.Message):
    __slots__ = ("PredictedUserTime", "UnpredictedUIDs")
    PREDICTEDUSERTIME_FIELD_NUMBER: _ClassVar[int]
    UNPREDICTEDUIDS_FIELD_NUMBER: _ClassVar[int]
    PredictedUserTime: _containers.RepeatedCompositeFieldContainer[UserWithTime]
    UnpredictedUIDs: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, PredictedUserTime: _Optional[_Iterable[_Union[UserWithTime, _Mapping]]] = ..., UnpredictedUIDs: _Optional[_Iterable[int]] = ...) -> None: ...
