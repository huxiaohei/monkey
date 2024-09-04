# -*- coding= utf-8 -*-

__time__ = '2024/08/12'
__author__ = '虎小黑'


from typing import cast
from message import base
from utils import mokey_json
from network.codec import Codec
from utils import monkey_config
from logger.logger import Logger
from network.buffer import Buffer
from message.message import JsonMessage
from message.rpc_message import RpcMessage


logger = Logger().get_logger('Monkey')


class CodecRpc(Codec):

    MAGIC_LEN = len(monkey_config.get_config().magic_code)
    MAGIC_CODE = monkey_config.get_config().magic_code.encode()
    HEARD_LENGTH = MAGIC_LEN + 8
    __class_name_cache: dict[type[JsonMessage], bytes] = {}

    def __init__(self) -> None:
        super().__init__()

    @classmethod
    def _encode_meta(cls, o: JsonMessage) -> bytes:
        # 1字节长度
        # N字节MessageName
        # M字节json
        _class = o.__class__
        name_bytes = cls.__class_name_cache.get(_class, None)
        if not name_bytes:
            if _class not in cls.__class_name_cache:
                name: str = _class.__qualname__
                cls.__class_name_cache[_class] = name.encode()
            name_bytes = cls.__class_name_cache[_class]
        json_data: bytes = cast(bytes, mokey_json.json_dumps(o.to_dict()))
        return b"".join((len(name_bytes).to_bytes(1, 'big'), name_bytes, json_data))

    def encode(self, msg: object) -> bytes:
        if not isinstance(msg, RpcMessage):
            msg = RpcMessage.from_msg(cast(JsonMessage, msg))
        msg = cast(RpcMessage, msg)
        meta_data = self._encode_meta(msg.meta)
        body_data = msg.body if msg.body else b''
        return b''.join(
            (
                self.MAGIC_CODE,
                len(meta_data).to_bytes(4, 'little'),
                len(body_data).to_bytes(4, 'little'),
                meta_data,
                body_data
            )
        )

    @classmethod
    def _decode_meta(cls, array: bytearray) -> JsonMessage | None:
        name_length = array[0]
        name = bytes(array[1: name_length + 1])
        model = base.find_model(name)
        if model is not None:
            json = mokey_json.json_loads(array[name_length + 1:])
            return model.from_dict(json)
        return None

    def decode(self, buffer: Buffer) -> RpcMessage | None:
        if buffer.readable_len() < self.HEARD_LENGTH:
            return None
        header = buffer.slice(self.HEARD_LENGTH)
        if header[:self.MAGIC_LEN] != self.MAGIC_CODE:
            logger.error(
                f"CodecRpc decode magic code error: {header[:self.MAGIC_LEN]}")
            raise ValueError(
                f"CodecRpc decode magic code error: {header[:self.MAGIC_LEN]}")
        meta_len = int.from_bytes(
            header[self.MAGIC_LEN: self.MAGIC_LEN + 4], 'little')
        body_len = int.from_bytes(header[self.MAGIC_LEN + 4:], 'little')
        if buffer.readable_len() < self.HEARD_LENGTH + meta_len + body_len:
            return None
        buffer.has_read(self.HEARD_LENGTH)
        meta_data = buffer.read(meta_len)
        body_data = buffer.read(body_len)
        meta = self._decode_meta(meta_data)
        if meta is None:
            logger.error(f"CodecRpc decode meta error: {meta_data}")
            raise ValueError(f"CodecRpc decode meta error: {meta_data}")
        return RpcMessage.from_msg(meta, body_data if body_data else b'')
