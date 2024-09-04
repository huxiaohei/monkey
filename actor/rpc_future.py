# -*- coding= utf-8 -*-

__time__ = '2024/09/01'
__author__ = '虎小黑'


import weakref
from asyncio.futures import Future
from utils.singleton import Singleton


class RpcFuture(Singleton):

    def __init__(self) -> None:
        super().__init__()
        self.__futures: dict[int, weakref.ReferenceType[Future]] = {}

    def add_future(self, request_id: int, future: Future) -> None:
        self.__futures[request_id] = weakref.ref(future)

    def get_future(self, request_id: int) -> Future | None:
        weak_future = self.__futures.get(request_id, None)
        if weak_future is None:
            return None
        return weak_future()
