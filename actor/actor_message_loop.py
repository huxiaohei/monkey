# -*- coding= utf-8 -*-

__time__ = '2024/08/31'
__author__ = '虎小黑'

import asyncio
import weakref
import traceback
from typing import cast
from message.message import GCActor
from utils import utils
from logger.logger import Logger
from actor.rpc_meta import RpcMeta
from message.base import JsonMessage
from actor.actor_base import ActorBase
from actor.actor_timer import ActorTimer
from utils.sequence_id import SequenceId
from network.socket_session import SocketSession
from message.rpc_message import RpcRequest, RpcResponse
from message.rpc_message import RpcErrorCode, RpcException, RpcMessage


logger = Logger().get_logger("Monkey")


class ActorMessageLoop(object):

    __loop_id = SequenceId()

    @classmethod
    async def send_error_resp(cls, session: SocketSession, request_id: int, e: Exception):
        resp = RpcResponse(request_id=request_id)
        if isinstance(e, RpcResponse):
            resp.error_code = e.error_code
            resp.error_str = e.error_str
        else:
            resp.error_code = RpcErrorCode.UnknownError
            resp.error_str = traceback.format_exc()
        await session.send(resp)

    @classmethod
    async def dispatch_actor_rpc_request(cls, actor: ActorBase, session: SocketSession | None, request: RpcRequest) -> None:
        assert actor.context

        try:
            method = RpcMeta.get_actor_rpc_impl_method(
                (request.server_name, request.method_name))
            if method is None:
                raise RpcException.method_not_found(
                    f'{request.server_name}:{request.actor_id}', request.method_name)
            actor.context.update_last_msg_time()
            result = method.__call__(actor, *request.args, **request.kwargs)
            if asyncio.iscoroutine(result):
                result = await result
            resp = RpcResponse()
            resp.request_id = request.request_id
            if session:
                await session.send(RpcMessage.from_msg(resp, utils.pickle_dump(result)))
        except Exception as e:
            if session is not None:
                await cls.send_error_resp(session, request.request_id, e)

    @ classmethod
    async def dispatch_actor_message_in_loop(cls, actor: ActorBase):
        assert actor.context

        if actor.context.loop_id != 0:
            return
        loop_id = cls.__loop_id.new_sequence_id()
        actor.context.loop_id = loop_id
        loaded = False
        try:
            try:
                await actor.active()
                loaded = True
            except Exception as e:
                logger.exception(
                    f'ActorMessageLoop dispatch_actor_message_in_loop {actor.actor_id} active failed error:{e}')
                actor.context.loop_id = 0
                return
            while True:
                await asyncio.sleep(0)
                o = await actor.context.pop_message()
                if o is None or isinstance(o, GCActor):
                    logger.error(
                        f'ActorMessageLoop dispatch_actor_message_in_loop {actor.actor_id} pop_message return None or GCActor')
                    break
                if isinstance(o, tuple):
                    session, msg = cast(
                        tuple[weakref.ReferenceType[SocketSession], RpcRequest], o)
                    await cls.dispatch_actor_rpc_request(actor, session(), msg)
                else:
                    await actor.dispatch_message(o)
        except Exception as e:
            logger.exception(
                f'ActorMessageLoop dispatch_actor_message_in_loop {actor.actor_id} error:{e}')

        try:
            if loaded:
                await actor.deactive()
        except Exception as e:
            logger.exception(
                f'ActorMessageLoop dispatch_actor_message_in_loop {actor.actor_id} deactive failed error:{e}')

        if actor.context.loop_id == loop_id:
            actor.context.loop_id = 0
            actor.context.reentrant_id = -1

        logger.info(
            f'ActorMessageLoop dispatch_actor_message_in_loop {actor.actor_type()}:{actor.actor_id} exit')

    def run_message_loop(self, actor: ActorBase) -> None:
        assert actor.context

        if actor.context.loop_id != 0:
            return
        asyncio.create_task(self.dispatch_actor_message_in_loop(actor))

    async def dispatch_actor_message(
            self,
            actor: ActorBase,
            session: SocketSession,
            msg: RpcRequest | RpcMessage | JsonMessage | ActorTimer) -> None:
        assert actor.context

        if isinstance(msg, RpcRequest):
            if actor.context.reentrant_id == msg.reentrant_id:
                asyncio.create_task(
                    self.dispatch_actor_rpc_request(actor, session, msg))
            else:
                await actor.context.push_message((weakref.ref(session), msg))
        else:
            await actor.context.push_message(msg)
