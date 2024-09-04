# -*- coding= utf-8 -*-

__time__ = '2024/08/11'
__author__ = '虎小黑'


from network.buffer import Buffer
from abc import ABC, abstractmethod


class Codec(ABC):

    def __init__(self) -> None:
        super().__init__()

    @classmethod
    def code_id(cls) -> str:
        return cls.__qualname__

    @abstractmethod
    def decode(self, buffer: Buffer) -> object | None:
        pass

    @abstractmethod
    def encode(self, data: object) -> bytes:
        pass
