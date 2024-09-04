# -*- coding= utf-8 -*-

__time__ = '2024/08/31'
__author__ = '虎小黑'


from dataclasses import dataclass
from message.base import JsonMessage


@dataclass(slots=True)
class RequestAccountLogin(JsonMessage):
    role_id: int = 0
    session_id: int = 0
    msg_seq_id: int = 0


@dataclass(slots=True)
class ResponseAccountLogin(JsonMessage):
    role_id: int = 0
    actor_id: str = ''
    actor_type: str = ''


@dataclass(slots=True)
class NotifyNewActorSession(JsonMessage):
    role_id: int = 0
    actor_id: str = ''
    actor_type: str = ''
    session_id: int = 0


@dataclass(slots=True)
class NotifyActorSessionAborted(JsonMessage):
    actor_id: str = ''
    actor_type: str = ''
    session_id: int = 0


@dataclass(slots=True)
class RequestCloseSession(JsonMessage):
    actor_id: str = ''
    actor_type: str = ''
    session_id: int = 0


@dataclass(slots=True)
class NotifyNewActorMessage(JsonMessage):
    actor_id: str = ''
    actor_type: str = ''
    session_id: int = 0
