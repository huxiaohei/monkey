# -*- coding= utf-8 -*-

__time__ = '2024/08/15'
__author__ = '虎小黑'


import asyncio
from typing import Type
from network import session_id
from network.codec import Codec
from utils import monkey_config
from logger.logger import Logger
from network.buffer import Buffer
from utils.monkey_time import MonkeyTime
from message.rpc_message import RpcMessage
from network.event_handler import EventHandler
from network.codec_manager import CodecManager
from network.socket_session import SocketSession


logger = Logger().get_logger('Monkey')


class TcpSocketSession(SocketSession):

    def __init__(self, session_id: int, codec: Codec, reader: asyncio.StreamReader, writer: asyncio.StreamWriter) -> None:
        super().__init__()
        self.__session_id = session_id
        self.__create_time = MonkeyTime.timestamp_sec()
        self.__last_update_time = MonkeyTime.timestamp_sec()
        self.__codec = codec
        self.__is_client = False
        self.__reader = reader
        self.__writer = writer
        peername = writer.get_extra_info('peername')
        if peername is None:
            self.__address = 'unknown'
        else:
            self.__address = f'{peername[0]}:{peername[1]}'
        self.__buffer = Buffer()
        self.__user_data: None | object = None
        self.__stop = False
        EventHandler.process_income_socket(self)

    def __del__(self):
        self.close()

    @property
    def session_id(self) -> int:
        return self.__session_id

    @property
    def create_time(self) -> int:
        return self.__create_time

    @property
    def heart_beat(self) -> int:
        return self.__last_update_time

    @property
    def is_dead(self) -> bool:
        return self.__stop or MonkeyTime.timestamp_sec() - self.__last_update_time >= monkey_config.get_config().tcp_session_timeout

    @property
    def is_closed(self) -> bool:
        return self.__stop or self.__writer.transport.is_closing() or self.__reader.at_eof()

    @property
    def is_client(self) -> bool:
        return self.__is_client

    @property
    def remote_address(self) -> str:
        return self.__address

    @property
    def codec(self) -> Codec:
        return self.__codec

    @property
    def user_data(self) -> None | object:
        return self.__user_data

    def set_user_data(self, user_data: None | object) -> None:
        self.__user_data = user_data

    def close(self) -> None:
        if self.__stop:
            return
        self.__stop = True
        try:
            self.__writer.close()
            self.__reader.feed_eof()
        except Exception as e:
            logger.error(
                f'TcpSocketSession Close session_id:{self.__session_id} error:{e}')
        finally:
            logger.info(
                f'TcpSocketSession Close session_id:{self.__session_id}')

    @staticmethod
    def get_real_type(o: object) -> Type:
        if isinstance(o, RpcMessage):
            return o.meta.__class__
        return o.__class__

    async def _recv_data(self) -> None | object:
        while not self.is_closed:
            msg = self.codec.decode(self.__buffer)
            if msg is not None:
                return msg
            self.__buffer.shrink()
            data = await self.__reader.read(1024)
            if not data or len(data) == 0:
                logger.error(
                    f'TcpSocketSession recv data empty session_id:{self.__session_id}')
                return None
            self.__buffer.append(data)
        return None

    async def recv(self) -> None | object:
        try:
            while not self.is_closed:
                msg = await self._recv_data()
                if msg is None:
                    logger.info(
                        f'TcpSocketSession recv session_id:{self.__session_id} recv None')
                    break
                await EventHandler.process_socket_message(self, self.get_real_type(msg), msg)
        except Exception as e:
            logger.exception(
                f'TcpSocketSession recv session_id:{self.__session_id} error:{e}')
            return None
        finally:
            EventHandler.process_close_socket(self.__session_id)

    async def send(self, msg: object) -> None:
        try:
            data = self.codec.encode(msg)
            self.__writer.write(data)
            await self.__writer.drain()
        except Exception as e:
            logger.error(
                f'TcpSocketSession send session_id:{self.__session_id} error:{e}')
            return None

    @classmethod
    async def connect(cls, host: str, port: int, codec_type: Type[Codec]) -> None | SocketSession:
        try:
            reader, writer = await asyncio.open_connection(host=host, port=port, limit=monkey_config.get_config().tcp_window_size)
            codec = CodecManager().get_codec(codec_type)
            if codec is None:
                logger.error(
                    f'TcpSocketSession connect codec not found:{codec_type.code_id()}')
                return None
            session = TcpSocketSession(
                session_id.new_session_id(), codec, reader, writer)
            session.__is_client = True
            asyncio.create_task(session.recv())
            return session
        except Exception as e:
            logger.error(
                f'TcpSocketSession connect host:{host} port:{port} error:{e}')
            return None
