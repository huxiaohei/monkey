# -*- coding= utf-8 -*-

__time__ = '2024/08/11'
__author__ = '虎小黑'


from typing import cast
from network.codec import Codec
from network.buffer import Buffer


class CodecEcho(Codec):

    def __init__(self) -> None:
        super().__init__()

    def decode(self, buffer: Buffer) -> object | None:
        if buffer.readable_len() > 0:
            arr = buffer.read()
            return arr.decode(encoding='utf-8')
        return None

    def encode(self, data: object) -> bytes:
        msg = cast(str, data)
        return msg.encode(encoding='utf-8')
