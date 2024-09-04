# -*- coding= utf-8 -*-

__time__ = '2024/09/04'
__author__ = '虎小黑'

import unittest
from utils import monkey_config
from network.buffer import Buffer


class TestBuffer(unittest.TestCase):

    def test_buffer(self):
        buf = Buffer()
        assert buf.readable_len() == 0
        assert buf.writable_len() == monkey_config.get_config().tcp_buffer_size

        buf.append(b'hello')
        assert buf.readable_len() == 5

        data = buf.read()
        assert data == b'hello'
        buf.shrink()
        assert buf.readable_len() == 0

        max_size = monkey_config.get_config().tcp_buffer_max_size
        for size in range(max_size):
            buf.append(b'a')
            assert buf.readable_len() == size + 1
        assert buf.writable_len() == 0

        with self.assertRaises(BufferError):
            buf.append(b'b')
