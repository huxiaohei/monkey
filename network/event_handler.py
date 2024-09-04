# -*- coding= utf-8 -*-

__time__ = '2024/08/15'
__author__ = '虎小黑'


from logger.logger import Logger
from typing import Awaitable, Callable, Type
from network.socket_session import SocketSession
from network.socket_session import SocketSessionManager


logger = Logger().get_logger('Monkey')


class EventHandler(object):

    __close_socket_handler: None | Callable[[SocketSession], None] = None
    __new_socket_handler: None | Callable[[SocketSession], None] = None
    __message_handler: dict[Type, Callable[[
        SocketSession, Type, object], Awaitable]] = {}

    @classmethod
    def set_close_socket_handler(cls, handler: Callable[[SocketSession], None]) -> None:
        cls.__close_socket_handler = handler

    @classmethod
    def set_new_socket_handler(cls, handler: Callable[[SocketSession], None]) -> None:
        cls.__new_socket_handler = handler

    @classmethod
    def register_hander(cls, clz: Type, handler: Callable[[SocketSession, Type, object], Awaitable]) -> None:
        if clz in cls.__message_handler:
            logger.error(
                f"EventHandler register handler error, clz:{clz} handler already exist")
            return
        cls.__message_handler[clz] = handler
        logger.debug(
            f"EventHandler register handler clz:{clz.__qualname__} handler:{handler.__qualname__}")

    @classmethod
    def process_income_socket(cls, session: SocketSession) -> None:
        suc = SocketSessionManager().add_session(session)
        if suc:
            try:
                if cls.__new_socket_handler is not None:
                    cls.__new_socket_handler(session)
            except Exception as e:
                logger.exception(
                    f"EventHandler new socket handler call error session_id:{session.session_id} error:{e}")
        logger.debug(
            f"EventHandler add session_id:{session.session_id} address:{session.remote_address} result:{suc}")

    @classmethod
    def process_close_socket(cls, session_id: int) -> None:
        session = SocketSessionManager().get_session(session_id)
        if session is not None:
            try:
                if cls.__close_socket_handler is not None:
                    cls.__close_socket_handler(session)
            except Exception as e:
                logger.exception(
                    f"EventHandler close socket handler call error session_id:{session.session_id} error:{e}")
            finally:
                SocketSessionManager().remove_session(session_id)
        else:
            logger.debug(
                f"EventHandler close session_id:{session_id} failed, session not found")

    @classmethod
    async def process_socket_message(cls, session: SocketSession, clz: Type,  msg: object) -> None:
        try:
            if cls.__message_handler.get(clz) is not None:
                await cls.__message_handler[clz](session, clz, msg)
            else:
                logger.error(
                    f"EventHandler process message error, clz:{clz} handler not found")
        except Exception as e:
            logger.exception(
                f"EventHandler process message error, session_id:{session.session_id} clz:{clz} error:{e}")
