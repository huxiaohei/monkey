# -*- coding= utf-8 -*-

__time__ = '2024/08/12'
__author__ = '虎小黑'


from dataclasses import dataclass
from message.base import JsonMessage


@dataclass(slots=True)
class RequestHeartBeat(JsonMessage):
    now_sec: int = 0


@dataclass(slots=True)
class ResponseHeartBeat(JsonMessage):
    now_sec: int = 0


@dataclass(slots=True)
class GCActor(JsonMessage):
    actor_id: str = ''
