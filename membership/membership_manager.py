# -*- coding= utf-8 -*-

__time__ = '2024/08/15'
__author__ = '虎小黑'


import random
from logger.logger import Logger
from typing import Callable, Type
from utils.singleton import Singleton
from utils.monkey_type import ActorType
from membership.placement import Placement
from membership.membership import Membership
from membership.server_node import ServerNode


logger = Logger().get_logger('Monkey')


class MembershipManager(Singleton):

    def __init__(self) -> None:
        super().__init__()
        self.__servers: dict[str, ServerNode] = {}
        self.__add_callback: None | Callable[[ServerNode], None] = None
        self.__remove_callback: None | Callable[[ServerNode], None] = None
        self.__membership: None | Membership = None
        self.__placement: None | Placement = None

    def set_add_callback(self, callback: Callable[[ServerNode], None]) -> None:
        self.__add_callback = callback

    def set_remove_callback(self, callback: Callable[[ServerNode], None]) -> None:
        self.__remove_callback = callback

    def set_membership(self, membership: Membership) -> None:
        self.__membership = membership

    def set_placement(self, placement: Placement) -> None:
        self.__placement = placement

    async def register_server(self, namespace: str, name: str, address: str, port: int, tags: list[str], meta: dict[str, str]) -> bool:
        if self.__membership is None:
            return False
        return await self.__membership.register_server(namespace, name, address, port, tags, meta)

    async def unregister_server(self, server_id: str) -> bool:
        if self.__membership is None:
            return False
        return await self.__membership.unregister_server(server_id)

    async def find_position(self, actor_type: Type[ActorType], actor_id: str) -> ServerNode | None:
        if self.__placement is None:
            return None
        return await self.__placement.find_position(actor_type, actor_id)

    async def actor_keep_alive(self, actor_type: Type[ActorType], actor_id: str, sec: int) -> bool:
        if self.__placement is None:
            return False
        return await self.__placement.actor_keep_alive(actor_type, actor_id, sec)

    def choose_member(self, mate: str) -> ServerNode | None:
        servers: list[ServerNode] = []
        for server in self.__servers.values():
            if not server.is_available:
                continue
            if not server.is_support(mate):
                continue
            servers.append(server)
        if len(servers) == 0:
            return None
        return random.choice(servers)

    def get_member(self, server_id: str) -> None | ServerNode:
        return self.__servers.get(server_id, None)

    def get_members(self) -> list[ServerNode]:
        return list(self.__servers.values())

    def add_member(self, member: ServerNode) -> None:
        if member.server_id not in self.__servers:
            logger.info(f'MembershipManager add member {member.server_id}')
            self.__servers[member.server_id] = member
            if self.__add_callback is not None:
                self.__add_callback(member)

    def remove_member(self, server_id: str) -> None:
        if server_id in self.__servers:
            logger.info(f'MembershipManager remove member {server_id}')
            server_node = self.__servers.pop(server_id, None)
            if self.__remove_callback is not None and server_node is not None:
                self.__remove_callback(server_node)

    def refresh_members(self, members: list[ServerNode]) -> None:
        old_server_ids = set(self.__servers.keys())
        new_server_ids = set([m.server_id for m in members])
        delete_server_ids = old_server_ids - new_server_ids
        for server_id in delete_server_ids:
            self.remove_member(server_id)
        for m in members:
            self.add_member(m)
