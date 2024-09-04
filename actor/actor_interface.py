# -*- coding= utf-8 -*-

__time__ = '2024/08/31'
__author__ = '虎小黑'


from abc import ABC
from typing import TypeVar


class ActorInterface(ABC):
    pass


ActorInterfaceType = TypeVar('ActorInterfaceType', bound=ActorInterface)
