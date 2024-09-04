# -*- coding= utf-8 -*-

__time__ = '2024/09/01'
__author__ = '虎小黑'

from typing import Generic, Self, TypeVar


K = TypeVar('K')
V = TypeVar('V')


class KVNode(Generic[K, V]):

    def __init__(self, k: K | None, v: V | None) -> None:
        super().__init__()
        self.__pre: Self | None = None
        self.__next: Self | None = None
        self.__k = k
        self.__v = v

    @property
    def val(self) -> V | None:
        return self.__v

    @val.setter
    def val(self, val: V) -> None:
        self.__v = val

    @property
    def key(self) -> K | None:
        return self.__k

    @property
    def pre(self) -> Self | None:
        return self.__pre

    @pre.setter
    def pre(self, pre: Self | None) -> None:
        self.__pre = pre

    @property
    def next(self) -> Self | None:
        return self.__next

    @next.setter
    def next(self, next: Self | None) -> None:
        self.__next = next


class DictLRU(Generic[K, V]):

    def __init__(self, max_size: int) -> None:
        super().__init__()
        self.__max_size = 1 if max_size <= 0 else max_size
        self.__cache_dict: dict[K, KVNode[K, V]] = {}
        self.__head: KVNode[K, V] = KVNode(None, None)
        self.__tail = self.__head

    def keys(self) -> list[K]:
        return list(self.__cache_dict.keys())

    def get(self, key: K) -> V | None:
        node = self.__cache_dict.get(key, None)
        if node is None:
            return None
        self.__remove(node)
        self.__append(node)
        return node.val

    def put(self, k: K, v: V) -> None:
        node = self.__cache_dict.get(k, None)
        if node:
            node.val = v
        else:
            if len(self.__cache_dict) >= self.__max_size and self.__head.next:
                self.__remove(self.__head.next)
            node = KVNode(k, v)
            self.__cache_dict[k] = node
            self.__append(node)

    def pop(self, key: K) -> V | None:
        node = self.__cache_dict.get(key, None)
        if node is None:
            return None
        self.__remove(node)
        return node.val

    def __append(self, node: KVNode[K, V]) -> None:
        self.__tail.next = node
        node.pre = self.__tail
        self.__tail = node

    def __remove(self, node: KVNode[K, V]) -> None:
        if node.key not in self.__cache_dict:
            return
        pre = node.pre
        next = node.next
        if pre:
            pre.next = next
            if node.key:
                self.__cache_dict.pop(node.key)
