from google.protobuf import empty_pb2 as _empty_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Tag(_message.Message):
    __slots__ = ("Name", "Probability", "Id")
    NAME_FIELD_NUMBER: _ClassVar[int]
    PROBABILITY_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    Name: str
    Probability: float
    Id: int
    def __init__(self, Name: _Optional[str] = ..., Probability: _Optional[float] = ..., Id: _Optional[int] = ...) -> None: ...

class PredictInfo(_message.Message):
    __slots__ = ("ID", "UID", "TagsIDs", "PlannedTime")
    ID_FIELD_NUMBER: _ClassVar[int]
    UID_FIELD_NUMBER: _ClassVar[int]
    TAGSIDS_FIELD_NUMBER: _ClassVar[int]
    PLANNEDTIME_FIELD_NUMBER: _ClassVar[int]
    ID: int
    UID: int
    TagsIDs: _containers.RepeatedScalarFieldContainer[int]
    PlannedTime: float
    def __init__(self, ID: _Optional[int] = ..., UID: _Optional[int] = ..., TagsIDs: _Optional[_Iterable[int]] = ..., PlannedTime: _Optional[float] = ...) -> None: ...

class PredictedInfo(_message.Message):
    __slots__ = ("ID", "UID", "PredictedTime")
    ID_FIELD_NUMBER: _ClassVar[int]
    UID_FIELD_NUMBER: _ClassVar[int]
    PREDICTEDTIME_FIELD_NUMBER: _ClassVar[int]
    ID: int
    UID: int
    PredictedTime: float
    def __init__(self, ID: _Optional[int] = ..., UID: _Optional[int] = ..., PredictedTime: _Optional[float] = ...) -> None: ...

class PredictRequest(_message.Message):
    __slots__ = ("Info",)
    INFO_FIELD_NUMBER: _ClassVar[int]
    Info: PredictInfo
    def __init__(self, Info: _Optional[_Union[PredictInfo, _Mapping]] = ...) -> None: ...

class PredictResponse(_message.Message):
    __slots__ = ("ActualTime",)
    ACTUALTIME_FIELD_NUMBER: _ClassVar[int]
    ActualTime: float
    def __init__(self, ActualTime: _Optional[float] = ...) -> None: ...

class PredictListRequest(_message.Message):
    __slots__ = ("Infos",)
    INFOS_FIELD_NUMBER: _ClassVar[int]
    Infos: _containers.RepeatedCompositeFieldContainer[PredictInfo]
    def __init__(self, Infos: _Optional[_Iterable[_Union[PredictInfo, _Mapping]]] = ...) -> None: ...

class PredictListResponse(_message.Message):
    __slots__ = ("PredictedUserTime", "UnpredictedUIDs")
    PREDICTEDUSERTIME_FIELD_NUMBER: _ClassVar[int]
    UNPREDICTEDUIDS_FIELD_NUMBER: _ClassVar[int]
    PredictedUserTime: _containers.RepeatedCompositeFieldContainer[PredictedInfo]
    UnpredictedUIDs: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, PredictedUserTime: _Optional[_Iterable[_Union[PredictedInfo, _Mapping]]] = ..., UnpredictedUIDs: _Optional[_Iterable[int]] = ...) -> None: ...

class PredictTagRequest(_message.Message):
    __slots__ = ("Title", "Description")
    TITLE_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    Title: str
    Description: str
    def __init__(self, Title: _Optional[str] = ..., Description: _Optional[str] = ...) -> None: ...

class PredictTagResponse(_message.Message):
    __slots__ = ("Tags",)
    TAGS_FIELD_NUMBER: _ClassVar[int]
    Tags: _containers.RepeatedCompositeFieldContainer[Tag]
    def __init__(self, Tags: _Optional[_Iterable[_Union[Tag, _Mapping]]] = ...) -> None: ...

class TagList(_message.Message):
    __slots__ = ("Tags",)
    TAGS_FIELD_NUMBER: _ClassVar[int]
    Tags: _containers.RepeatedCompositeFieldContainer[Tag]
    def __init__(self, Tags: _Optional[_Iterable[_Union[Tag, _Mapping]]] = ...) -> None: ...
