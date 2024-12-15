from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class RecalculateAndSaveTaskRequest(_message.Message):
    __slots__ = ("ID", "UID", "Title", "Description", "PlannedTime", "ActualTime")
    ID_FIELD_NUMBER: _ClassVar[int]
    UID_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    PLANNEDTIME_FIELD_NUMBER: _ClassVar[int]
    ACTUALTIME_FIELD_NUMBER: _ClassVar[int]
    ID: int
    UID: int
    Title: str
    Description: str
    PlannedTime: float
    ActualTime: float
    def __init__(self, ID: _Optional[int] = ..., UID: _Optional[int] = ..., Title: _Optional[str] = ..., Description: _Optional[str] = ..., PlannedTime: _Optional[float] = ..., ActualTime: _Optional[float] = ...) -> None: ...

class RecalculateAndSaveTaskResponse(_message.Message):
    __slots__ = ("Status",)
    STATUS_FIELD_NUMBER: _ClassVar[int]
    Status: str
    def __init__(self, Status: _Optional[str] = ...) -> None: ...

class RecalculateRequest(_message.Message):
    __slots__ = ("UID",)
    UID_FIELD_NUMBER: _ClassVar[int]
    UID: int
    def __init__(self, UID: _Optional[int] = ...) -> None: ...

class RecalculateResponse(_message.Message):
    __slots__ = ("Status",)
    STATUS_FIELD_NUMBER: _ClassVar[int]
    Status: str
    def __init__(self, Status: _Optional[str] = ...) -> None: ...

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
