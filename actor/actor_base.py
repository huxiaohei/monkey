# -*- coding= utf-8 -*-

__time__ = '2024/08/31'
__author__ = '虎小黑'


import weakref
from queue import Queue
from actor.actor import Actor
from logger.logger import Logger
from message.base import JsonMessage
from typing import Any, Callable, cast
from message.rpc_message import RpcMessage
from actor.actor_context import ActorContext
from actor.actor_timer import ActorTimer, ActorTimerManager
from network.socket_session import SocketSession, SocketSessionManager
from message.gateway_message import NotifyActorSessionAborted, NotifyNewActorMessage, NotifyNewActorSession


logger = Logger().get_logger('Monkey')


class ActorBase(Actor):

    def __init__(self, actor_id: str) -> None:
        super().__init__()
        self.__actor_id: str = actor_id
        self.__context: ActorContext | None = None
        self.__socket_session: weakref.ReferenceType[SocketSession] | None = None
        self.__msg_cache: Queue[JsonMessage] = Queue(10)
        self.__timer_manager: ActorTimerManager | None = ActorTimerManager(
            weakref.ref(self))

    def init(self, context: ActorContext, socker_session: weakref.ReferenceType[SocketSession] | None) -> None:
        self.__context = context
        self.__socket_session = socker_session

    @property
    def actor_id(self) -> str:
        return self.__actor_id

    @property
    def context(self) -> ActorContext | None:
        return self.__context

    @property
    def gc_time(self) -> int:
        return 1800

    @property
    def actor_weight(cls) -> int:
        return 1

    async def on_active(self) -> None:
        pass

    async def on_deactive(self) -> None:
        pass

    async def dispatch_custom_message(self, msg: JsonMessage) -> None:
        pass

    async def on_new_session(self, msg: NotifyNewActorSession) -> None:
        self.__socket_session = None
        session = SocketSessionManager().get_session(msg.session_id)
        if not session:
            return
        self.__socket_session = weakref.ref(session)

    async def on_session_aborted(self, msg: NotifyActorSessionAborted) -> None:
        if not self.__socket_session:
            return
        session = self.__socket_session()
        if not session:
            return
        if session.session_id != msg.session_id:
            return
        self.__socket_session = None

    async def active(self) -> None:
        try:
            await self.on_active()
        except Exception as e:
            logger.exception(
                f'{type(self).__qualname__} active error actor_id:{self.actor_id} error:{e}')

    async def deactive(self) -> None:
        try:
            if self.__timer_manager:
                self.__timer_manager.unregister_all_timer()
                del self.__timer_manager
                self.__timer_manager = None
        except Exception as e:
            logger.exception(
                f'{type(self).__qualname__} deactive error actor_id:{self.actor_id} error:{e}')

        try:
            await self.on_deactive()
        except Exception as e:
            logger.exception(
                f'{type(self).__qualname__} deactive error actor_id:{self.actor_id} error:{e}')

    async def send_message(self, msg: JsonMessage) -> None:
        if self.__msg_cache.full():
            self.__msg_cache.get()
        self.__msg_cache.put(msg)
        if not self.__socket_session:
            return
        session = self.__socket_session()
        if not session:
            return
        await session.send(msg)

    async def dispatch_message(self, msg: JsonMessage | RpcMessage | ActorTimer) -> None:
        try:
            if isinstance(msg, ActorTimer):
                msg.tick()
            else:
                if isinstance(msg, RpcMessage):
                    rpc_msg = cast(RpcMessage, msg)
                    if isinstance(rpc_msg.meta, NotifyNewActorSession):
                        await self.on_new_session(rpc_msg.meta)
                    elif isinstance(rpc_msg.meta, NotifyActorSessionAborted):
                        await self.on_session_aborted(rpc_msg.meta)
                    elif isinstance(rpc_msg.meta, NotifyNewActorMessage):
                        await self.dispatch_custom_message(rpc_msg.meta)
                    else:
                        logger.error(
                            f'{type(self).__qualname__} dispatch_message error actor_id:{self.actor_id} msg:{msg}')
                else:
                    await self.dispatch_custom_message(msg)
                if self.__context:
                    self.__context.update_last_msg_time()
        except Exception as e:
            logger.exception(
                f'{type(self).__qualname__} dispatch_message error actor_id:{self.actor_id} error:{e}')

    def register_timer(
            self,
            delay: int,
            interval: int,
            repetition: int,
            fn: Callable[..., None],
            target: object,
            *args: Any,
            **kwargs: Any) -> int:
        if not self.__timer_manager:
            return -1
        timer = self.__timer_manager.register_timer(
            delay, interval, repetition, fn, target, *args, **kwargs)
        return timer.timer_id

    def unregister_timer(self, timer_id: int) -> None:
        if not self.__timer_manager:
            return
        self.__timer_manager.unregister_timer(timer_id)
