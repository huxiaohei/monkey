# -*- coding= utf-8 -*-

__time__ = '2024/08/15'
__author__ = '虎小黑'


from pydantic import BaseModel
from abc import ABC, abstractmethod
from network.socket_session import SocketSession


class ServerNode(ABC, BaseModel):

    @property
    @abstractmethod
    def server_id(self) -> str:
        pass

    @property
    @abstractmethod
    def session_id(self) -> int:
        pass

    @property
    @abstractmethod
    def session(self) -> None | SocketSession:
        pass

    @session.setter
    @abstractmethod
    def session(self, session: SocketSession) -> None:
        pass

    @property
    @abstractmethod
    def address(self) -> str:
        pass

    @property
    @abstractmethod
    def port(self) -> int:
        pass

    @property
    @abstractmethod
    def weight(self) -> int:
        pass

    @property
    @abstractmethod
    def is_available(self) -> bool:
        pass

    @abstractmethod
    def is_support(self, mate: str) -> bool:
        pass
