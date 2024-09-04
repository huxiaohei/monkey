# -*- coding= utf-8 -*-

__time__ = '2024/08/11'
__author__ = '虎小黑'


class SequenceId(object):

    SHIFT = 1000000

    def __init__(self) -> None:
        self.__seed = 0
        self.__id = 0

    @property
    def seed(self) -> int:
        return self.__seed

    def set_seed(self, seed: int) -> None:
        if seed < self.__seed:
            raise ValueError(
                'seed must be greater than or equal to the previous seed')
        self.__seed = seed
        self.__id = self.__seed * self.SHIFT

    def new_sequence_id(self) -> int:
        self.__id += 1
        return self.__id
