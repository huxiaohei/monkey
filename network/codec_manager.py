# -*- coding= utf-8 -*-

__time__ = '2024/08/11'
__author__ = '虎小黑'


from typing import Type
from network.codec import Codec
from logger.logger import Logger
from utils.singleton import Singleton
from network.codec_rpc import CodecRpc
from network.codec_echo import CodecEcho


logger = Logger().get_logger('Monkey')


class CodecManager(Singleton):

    def __init__(self):
        self.__codec_map: dict[str, Codec] = {}
        self.register_codec(CodecEcho())
        self.register_codec(CodecRpc())

    def register_codec(self, codec: Codec) -> None:
        if codec.code_id() in self.__codec_map:
            logger.error(f'Codec {codec.code_id()} already exists')
            return
        self.__codec_map[codec.code_id()] = codec

    def get_codec(self, codec_type: Type[Codec]) -> Codec | None:
        return self.__codec_map.get(codec_type.code_id(), None)
