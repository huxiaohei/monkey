# -*- coding= utf-8 -*-

__time__ = '2024/08/15'
__author__ = '虎小黑'


import asyncio
from network import session_id
from network.codec import Codec
from utils import monkey_config
from logger.logger import Logger
from typing import Coroutine, Type
from utils.singleton import Singleton
from network.codec_manager import CodecManager
from network.tcp_session import TcpSocketSession


logger = Logger().get_logger('Monkey')


class TcpServer(Singleton):

    def __init__(self) -> None:
        super().__init__()
        try:
            import uvloop
            uvloop.install()
        except Exception as e:
            logger.error('uvloop install failed, error:%s', e)
        self.__loop = asyncio.get_event_loop()

    @classmethod
    async def __handle_new_session(cls, codec: Codec, reader: asyncio.StreamReader, writer: asyncio.StreamWriter) -> None:
        session = TcpSocketSession(
            session_id.new_session_id(), codec, reader, writer)
        await session.recv()

    async def listen(self, host: str, port: int, codec_type: Type[Codec]) -> None:
        codec = CodecManager().get_codec(codec_type)
        if codec is None:
            logger.error(
                f'TcpServer listen error Codec:{codec_type.code_id()} not found')
            return

        async def callback(reader: asyncio.StreamReader, writer: asyncio.StreamWriter) -> None:
            await self.__handle_new_session(codec, reader, writer)

        try:
            logger.info(f'TcpServer listen on {host}:{port}')
            await asyncio.start_server(callback, host=host, port=port, limit=monkey_config.get_config().tcp_window_size)
        except Exception as e:
            logger.error(
                f'TcpServer listen on {host}:{port} failed, error:{e}')

    def create_task(self, co: Coroutine) -> None:
        self.__loop.create_task(co)

    def run(self) -> None:
        self.__loop.run_forever()
