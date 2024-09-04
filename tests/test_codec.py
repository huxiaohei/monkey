# -*- coding= utf-8 -*-

__time__ = '2024/09/04'
__author__ = '虎小黑'


import unittest
from message.message import RequestHeartBeat
from message.rpc_message import RpcMessage, RpcRequest
from network.buffer import Buffer
from network.codec_rpc import CodecRpc
from network.codec_echo import CodecEcho
from network.codec_manager import CodecManager
from utils import utils
from logger.logger import Logger


logger = Logger().get_logger('Monkey')


class TestCodec(unittest.TestCase):

    def test_codec_echo(self):
        codec = CodecManager().get_codec(CodecEcho)
        if codec is None:
            self.fail('CodecEcho not found')

        buf = Buffer.from_bytes(codec.encode('hello'))
        assert codec.decode(buf) == 'hello'

    def test_codec_rpc(self):
        codec = CodecManager().get_codec(CodecRpc)
        if codec is None:
            self.fail('CodecRpc not found')

        request = RpcRequest(
            msg_id=10001,
            name='RpcRequest',
            server_name='Monkey',
            method_name='test',
            actor_id='123',
            reentrant_id=1,
            request_id=1001,
            server_id=1001,
            _args=[1, 2, 3],
            _kwargs={'a': 1, 'b': 2})
        buf = codec.encode(request)

        msg = codec.decode(Buffer.from_bytes(buf))
        assert isinstance(msg, RpcMessage)
        assert isinstance(msg.meta, RpcRequest)
        assert request.to_dict() == msg.meta.to_dict()
