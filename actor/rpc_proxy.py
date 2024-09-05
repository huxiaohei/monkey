# -*- coding= utf-8 -*-

__time__ = '2024/09/05'
__author__ = '虎小黑'


import asyncio
import weakref
from actor.actor_base import ActorBase
from utils import utils
from typing import Any, Type, cast
from actor.rpc_future import RpcFuture
from utils.monkey_type import ActorType
from actor.actor_context import ActorContext
from actor.rpc_request_id import RpcRequestId
from membership.server_node import ServerNode
from actor.actor_interface import ActorInterfaceType
from membership.membership_manager import MembershipManager
from message.rpc_message import RpcErrorCode, RpcException, RpcMessage, RpcRequest


async def __rpc_call(unique_id: int) -> Any:
    future = asyncio.get_event_loop().create_future()
    RpcFuture().add_future(unique_id, future)
    await asyncio.wait_for(future, timeout=5)
    return future.result()


class __RpcMethodObject(object):

    def __init__(self,
                 server_name: str,
                 actor_id: str,
                 method_name: str,
                 reentrant_id: int,
                 server_node: None | ServerNode,
                 check_position: bool = True) -> None:
        super().__init__()
        self.__server_name = server_name
        self.__actor_id = actor_id
        self.__method_name = method_name
        self.__reentrant_id = reentrant_id
        self.__server_node = server_node
        self.__check_position = check_position

    async def __send_request(self, *args, **kwargs) -> int:
        if self.__server_node:
            position = self.__server_node
        else:
            position = await MembershipManager().find_position(self.__server_name, self.__actor_id)

        if position is None:
            raise RpcException.position_not_found(
                self.__server_name, self.__actor_id)

        session = position.session
        if session is None:
            raise RpcException.target_server_not_valid(
                self.__server_name, self.__actor_id, position.server_id)

        req = RpcRequest(
            server_name=self.__server_name,
            method_name=self.__method_name,
            actor_id=self.__actor_id,
            reentrant_id=self.__reentrant_id,
            request_id=RpcRequestId.get_request_id(),
            server_id=position.server_id)
        if not self.__check_position:
            req.server_id = ''
        raw_rags = utils.pickle_dump((args, kwargs))
        await session.send(RpcMessage.from_msg(req, raw_rags))
        return req.request_id

    async def __call__(self, *args, **kwargs) -> Any:
        for _ in range(3):
            try:
                request_id = await self.__send_request(*args, **kwargs)
                return await __rpc_call(request_id)
            except RpcException as e:
                if e.code == RpcErrorCode.RpcErrorPositionChanged:
                    MembershipManager().remove_position_from_cache(
                        self.__server_name, self.__actor_id)
                    continue
                raise e


class __RpcProxyObject(object):

    def __init__(self,
                 server_name: str,
                 actor_id: str,
                 context: None | ActorContext,
                 server_node: None | ServerNode,
                 check_position: bool = True) -> None:
        super().__init__()
        self.__server_name = server_name
        self.__actor_id = actor_id
        self.__context: weakref.ReferenceType[ActorContext] | None = None
        if context is not None:
            self.__context = weakref.ref(context)
        self.__server_node = server_node
        self.__check_position = check_position

    def __getattr__(self, name: str) -> Any:
        ctx: ActorContext | None = None
        if self.__context is not None:
            ctx = self.__context()
        if ctx is not None:
            reentrant_id = ctx.reentrant_id
        else:
            reentrant_id = RpcRequestId.get_request_id()
        method = __RpcMethodObject(self.__server_name,
                                   self.__actor_id,
                                   name,
                                   reentrant_id,
                                   self.__server_node,
                                   self.__check_position)
        return method


def get_rpc_proxy(actor_type: Type[ActorInterfaceType],
                  actor_id: str,
                  context: None | ActorContext = None,
                  server_node: None | ServerNode = None,
                  check_position: bool = True) -> ActorInterfaceType:
    o = __RpcProxyObject(actor_type.__qualname__, actor_id, context,
                         server_node, check_position)
    return cast(ActorInterfaceType, o)
