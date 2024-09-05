# -*- coding= utf-8 -*-

__time__ = '2024/08/12'
__author__ = '虎小黑'


from typing import Self
from enum import IntEnum
from dataclasses import dataclass
from message.base import JsonMessage


class RpcErrorCode(IntEnum):
    OK = 0
    UnknownError = -1
    TimeoutError = -1001
    PositionNotFound = -1002
    MehodNotFound = -1003
    RpcErrorPositionChanged = -2001


class RpcException(Exception):

    def __init__(self, code: RpcErrorCode, message: str) -> None:
        super().__init__(message)
        self.__code = code

    @property
    def code(self) -> RpcErrorCode:
        return self.__code

    @classmethod
    def position_not_found(cls, actor_type: str, actor_unique_id: str) -> Self:
        return cls(RpcErrorCode.PositionNotFound, f"{actor_type}:{actor_unique_id} position not found")

    @classmethod
    def target_server_not_valid(cls, actor_type: str, actor_unique_id: str, server_id: str) -> Self:
        return cls(RpcErrorCode.PositionNotFound, f"{actor_type}:{actor_unique_id} target server:{server_id} not valid")

    @classmethod
    def method_not_found(cls, actor_type: str, actor_unique_id: str, method_name: str) -> Self:
        return cls(RpcErrorCode.MehodNotFound, f"{actor_type}:{actor_unique_id} {method_name} not found")


@dataclass(slots=True)
class RpcMessage:
    meta: JsonMessage
    body: bytes = b''

    @classmethod
    def from_msg(cls, meta: JsonMessage, body: bytes = b'') -> Self:
        return cls(meta, body)


@dataclass(slots=True)
class RpcRequest(JsonMessage):
    server_name: str = ''
    method_name: str = ''
    actor_id: str = ''
    reentrant_id: int = 0
    request_id: int = 0
    server_id: str = ''
    _args: list | None = None
    _kwargs: dict | None = None

    @property
    def args(self) -> list | None:
        return self._args

    @property
    def kwargs(self) -> dict | None:
        return self._kwargs


@dataclass(slots=True)
class RpcResponse(JsonMessage):
    request_id: int = 0
    error_code: RpcErrorCode = RpcErrorCode.OK
    error_str: str = ''
