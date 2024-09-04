# -*- coding= utf-8 -*-

__time__ = '2024/08/31'
__author__ = '虎小黑'


from abc import ABC, abstractmethod
from actor.actor_context import ActorContext
from actor.actor_interface import ActorInterface


class Actor(ActorInterface, ABC):

    @property
    @abstractmethod
    def actor_id(self) -> str:
        pass

    @property
    @abstractmethod
    def context(self) -> ActorContext:
        pass

    @classmethod
    def actor_type(cls) -> str:
        return cls.__qualname__

    @property
    @abstractmethod
    def gc_time(cls) -> int:
        pass

    @property
    @abstractmethod
    def actor_weight(cls) -> int:
        pass
