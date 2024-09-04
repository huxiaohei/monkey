# -*- coding= utf-8 -*-

__time__ = '2024/08/31'
__author__ = '虎小黑'


import weakref
from asyncio import Queue
from message.base import JsonMessage
from actor.actor_timer import ActorTimer
from utils.monkey_time import MonkeyTime
from network.socket_session import SocketSession
from message.rpc_message import RpcMessage, RpcRequest


MsgType = tuple[
    weakref.ReferenceType[SocketSession], RpcRequest] | RpcMessage | JsonMessage | ActorTimer


class ActorContext(object):

    def __init__(self):
        super().__init__()
        self.__queue: Queue[MsgType] = Queue()
        self.__loop_id: int = 0
        self.__reentrant_id: int = 0
        self.__last_msg_time: int = MonkeyTime.timestamp_sec()

    @property
    def loop_id(self) -> int:
        return self.__loop_id

    @loop_id.setter
    def loop_id(self, loop_id: int) -> None:
        self.__loop_id = loop_id

    @property
    def reentrant_id(self) -> int:
        return self.__reentrant_id

    @reentrant_id.setter
    def reentrant_id(self, reentrant_id: int) -> None:
        self.__reentrant_id = reentrant_id

    @property
    def last_msg_time(self) -> int:
        return self.__last_msg_time

    def update_last_msg_time(self) -> None:
        self.__last_msg_time = MonkeyTime.timestamp_sec()

    async def pop_message(self) -> MsgType:
        return await self.__queue.get()

    async def push_message(self, message: MsgType) -> None:
        try:
            self.__queue.put_nowait(message)
        except Exception as _:
            await self.__queue.put(message)
