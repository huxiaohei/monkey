# -*- coding= utf-8 -*-

__time__ = '2024/08/11'
__author__ = '虎小黑'


from utils import monkey_config
from logger.logger import Logger


logger = Logger().get_logger('Monkey')


class Buffer(object):

    def __init__(self) -> None:
        self.__buffer = bytearray(monkey_config.get_config().tcp_buffer_size)
        self.__read = 0
        self.__write = 0

    @staticmethod
    def from_bytes(data: bytes) -> 'Buffer':
        buffer = Buffer()
        buffer.append(data)
        return buffer

    def readable_len(self) -> int:
        return self.__write - self.__read

    def writable_len(self) -> int:
        return len(self.__buffer) - self.__write

    def has_read(self, length: int) -> None:
        self.__read += length

    def slice(self, length: int) -> bytearray:
        if length <= 0:
            return self.__buffer[self.__read:self.__write]
        return self.__buffer[self.__read:self.__read + length]

    def shrink(self) -> None:
        length = self.readable_len()
        if length > 0:
            self.__buffer[:length] = self.__buffer[self.__read:self.__write]
        self.__write = length
        self.__read = 0

    def read(self, length: int = -1) -> bytearray:
        if length > self.readable_len() or length <= 0:
            length = self.readable_len()
        data = self.slice(length)
        self.has_read(length)
        return data

    def append(self, data: bytes) -> None:
        first_space = min(self.writable_len(), len(data))
        second_space = len(data) - first_space
        self.__buffer[self.__write:self.__write +
                      first_space] = data[:first_space]
        self.__write += first_space
        if second_space > 0:
            if len(self.__buffer) + second_space > monkey_config.get_config().tcp_buffer_max_size:
                logger.error(
                    f'Monkey network buffer is full bytes:{self.__buffer} read:{self.__read} write:{self.__write} extend:{data[first_space:]}')
                raise BufferError('buffer is full')
            self.__buffer.extend(data[first_space:])
            self.__write += second_space
