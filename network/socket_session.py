# -*- coding= utf-8 -*-

__time__ = '2024/08/12'
__author__ = '虎小黑'


import time
import asyncio
from abc import abstractmethod
from network.codec import Codec
from utils import monkey_config
from logger.logger import Logger
from utils.singleton import Singleton


logger = Logger().get_logger('Monkey')


class SocketSession:

    @property
    @abstractmethod
    def session_id(self) -> int:
        pass

    @property
    @abstractmethod
    def create_time(self) -> int:
        pass

    @abstractmethod
    def heart_beat(self, time_now: float) -> None:
        pass

    @abstractmethod
    def is_dead(self, current_time) -> bool:
        pass

    @property
    @abstractmethod
    def is_closed(self) -> bool:
        pass

    @property
    @abstractmethod
    def is_client(self) -> bool:
        pass

    @property
    @abstractmethod
    def remote_address(self) -> str:
        pass

    @property
    @abstractmethod
    def codec(self) -> Codec:
        pass

    @property
    @abstractmethod
    def user_data(self) -> None | object:
        pass

    @abstractmethod
    def set_user_data(self, data: object):
        pass

    @abstractmethod
    def close(self) -> None:
        pass

    @abstractmethod
    async def send(self, msg: object) -> None:
        pass


class SocketSessionManager(Singleton):

    def __init__(self):
        super().__init__()
        self._sessions: dict[int, SocketSession] = {}

    def add_session(self, session: SocketSession) -> bool:
        session_id = session.session_id
        if session_id in self._sessions:
            logger.warning(
                f"SocketSessionManager add session {session_id} failed, already exists")
            return False
        self._sessions[session_id] = session
        logger.info(
            f"SocketSessionManager add session SessionId:{session_id} RemoteAddress:{session.remote_address}")
        return True

    def remove_session(self, session_id: int):
        if session_id not in self._sessions:
            logger.warning(
                f"SocketSessionManager remove session {session_id} failed, not exists")
            return
        session = self._sessions.pop(session_id)
        session.close()
        logger.info(
            f"SocketSessionManager remove session SessionId:{session_id} RemoteAddress:{session.remote_address}")

    def get_session(self, session_id: int) -> SocketSession | None:
        """业务上持有session的地方,建议使用弱引用,否则会拉长session的生命周期"""
        return self._sessions.get(session_id, None)

    async def _gc_loop(self):
        deads: list[SocketSession] = []
        while True:
            deads.clear()
            current_time = time.time()
            for item in self._sessions.values():
                if item.is_dead(current_time):
                    deads.append(item)
            for item in deads:
                self.remove_session(item.session_id)
                logger.warning(f"session {item.session_id} is dead")
            await asyncio.sleep(monkey_config.get_config().socket_gc_interval)
