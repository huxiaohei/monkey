# -*- coding= utf-8 -*-

__time__ = '2024/09/01'
__author__ = '虎小黑'


from abc import ABC, abstractmethod


class Membership(ABC):

    @abstractmethod
    async def register_server(self, namespace: str, name: str, address: str, port: int, tags: list[str], meta: dict[str, str]) -> bool:
        pass

    @abstractmethod
    async def unregister_server(self, server_id: str) -> bool:
        pass

    @abstractmethod
    async def check_health(self, namespace: str, server_tags: list[str] = []) -> None:
        pass
