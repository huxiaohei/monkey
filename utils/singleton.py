# -*- coding= utf-8 -*-

__time__ = '2024/08/06'
__author__ = '虎小黑'


from typing import Any


class SingletonMeta(type):

    _instance = {}

    def __call__(cls, *args: Any, **kwds: Any):
        if cls not in cls._instance:
            instance = super(SingletonMeta, cls).__call__(*args, **kwds)
            cls._instance[cls] = instance
        return cls._instance[cls]


class Singleton(metaclass=SingletonMeta):
    pass
