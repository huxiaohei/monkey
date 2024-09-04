# -*- coding= utf-8 -*-

__time__ = '2024/09/01'
__author__ = '虎小黑'


from typing import TypeVar
from actor.actor_base import ActorBase


ActorType = TypeVar('ActorType', bound=ActorBase)
