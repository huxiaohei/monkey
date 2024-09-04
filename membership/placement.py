# -*- coding= utf-8 -*-

__time__ = '2024/09/02'
__author__ = '虎小黑'


from typing import Type
from abc import abstractmethod
from utils.monkey_type import ActorType
from membership.server_node import ServerNode
from membership.membership_manager import MembershipManager


class Placement():

    def __init__(self) -> None:
        super().__init__()
        MembershipManager().set_add_callback(self.on_add_server)
        MembershipManager().set_remove_callback(self.on_remove_server)

    @abstractmethod
    def on_add_server(self, node: ServerNode):
        pass

    @abstractmethod
    def on_remove_server(self, node: ServerNode):
        pass

    @abstractmethod
    def find_position_in_cache(self, actor_type: Type[ActorType], actor_id: str) -> ServerNode | None:
        pass

    @abstractmethod
    async def find_position(self, actor_type: Type[ActorType], actor_id: str) -> ServerNode | None:
        pass

    @abstractmethod
    async def actor_keep_alive(self, actor_type: Type[ActorType], actor_id: str, sec: int) -> bool:
        pass
