# -*- coding= utf-8 -*-

__time__ = '2024/09/01'
__author__ = '虎小黑'


import asyncio
from typing import TypeVar
from typing import Type, cast
from logger.logger import Logger
from actor.rpc_meta import RpcMeta
from message.message import GCActor
from utils.monkey_type import ActorType
from utils.singleton import Singleton
from actor.actor_base import ActorBase
from utils.monkey_time import MonkeyTime


logger = Logger().get_logger("Monkey")


class ActorManager(Singleton):

    def __init__(self) -> None:
        super().__init__()
        self.__actors: dict[str, ActorBase] = {}
        self.__weight = 0

    @property
    def weight(self) -> int:
        return self.__weight

    def get_actor(self, actor_type: Type[ActorType], actor_id: str) -> ActorType | None:
        unique_id = f'{actor_type.actor_type()}:{actor_id}'
        return cast(ActorType, self.__actors.get(unique_id, None))

    def get_or_new(self, actor_type: Type[ActorType], actor_id: str) -> ActorType | None:
        actor = self.get_actor(actor_type, actor_id)
        if actor:
            return actor

        impl_type = RpcMeta.get_actor_impl_type(actor_type)
        if impl_type is None:
            raise Exception(f'actor type {actor_type.__qualname__} not found')
        actor = impl_type(actor_id)
        self.__actors[f'{actor_type.actor_type()}:{actor_id}'] = actor
        return actor

    async def calc_weight_loop(self):
        while True:
            await asyncio.sleep(10)
            weight = 0
            for actor in self.__actors.values():
                weight += actor.actor_weight
            self.__weight = weight

    async def __gc_actor(self, actor: ActorBase) -> bool:
        if actor.context is None:
            return False
        if actor.context.last_msg_time + actor.gc_time <= MonkeyTime.timestamp_sec():
            logger.info(
                f'ActorManager __gc_actor {actor.actor_type()}:{actor.actor_id}')
            await actor.context.push_message(GCActor(actor_id=actor.actor_id))
            return True
        return False

    async def gc_loop(self):
        while True:
            await asyncio.sleep(60)
            unique_ids: list[str] = []
            now = MonkeyTime.timestamp_sec()
            for actor in self.__actors.values():
                if actor.context is None:
                    continue
                try:
                    if await self.__gc_actor(actor):
                        unique_ids.append(
                            f"{actor.actor_type()}:{actor.actor_id}")
                except Exception as e:
                    logger.exception(
                        f'ActorManager gc_loop gc {actor.actor_type()}:{actor.actor_id} error:{e}')
                finally:
                    if MonkeyTime.timestamp_sec() - now > 100:
                        break
            for unique_id in unique_ids:
                self.__actors.pop(unique_id, None)
