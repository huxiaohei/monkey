# -*- coding= utf-8 -*-

__time__ = '2024/09/01'
__author__ = '虎小黑'


import time
import asyncio
from typing import Type
from utils.lru import DictLRU
from redis.asyncio import Redis
from utils import monkey_config
from logger.logger import Logger
from utils.singleton import Singleton
from pydantic import BaseModel, Field
from network.codec_rpc import CodecRpc
from utils.monkey_type import ActorType
from utils.monkey_time import MonkeyTime
from utils.redis_script import RedisScript
from membership.placement import Placement
from message.message import RequestHeartBeat
from membership.server_node import ServerNode
from network.tcp_session import TcpSocketSession
from membership.membership_manager import MembershipManager


logger = Logger().get_logger('Monkey')


class ActorPlacement(BaseModel):
    actor_id: str = Field(alias='actor_id')
    actor_type: str = Field(alias='actor_type')
    server_id: str = Field(alias='server_id')
    expire_time: int = Field(alias='expire_time')


class RedisPlacement(Singleton, Placement):

    def __init__(self, uri: str) -> None:
        super().__init__()
        self.__cache: DictLRU[str, ActorPlacement] = DictLRU(1024 * 2)
        self.__redis_client = Redis.from_url(uri)
        asyncio.create_task(self.__heart_beat_loop())

    @classmethod
    def placement_key(cls, actor_type: Type[ActorType], actor_id: str) -> str:
        return f'{actor_type.actor_type()}:{actor_id}'

    def on_add_server(self, node: ServerNode):
        logger.info(f'RedisPlacement on_add_server {node.server_id}')
        pass

    def on_remove_server(self, node: ServerNode):
        logger.info(f'RedisPlacement on_remove_server {node.server_id}')
        if node.session:
            node.session.close()
        for key in self.__cache.keys():
            placement = self.__cache.get(key)
            if placement is not None and placement.server_id == node.server_id:
                self.__cache.pop(key)

    def find_position_in_cache(self, actor_type: Type[ActorType], actor_id: str) -> ServerNode | None:
        actor_placement = self.__cache.get(
            self.placement_key(actor_type, actor_id))
        if actor_placement is None:
            return None
        if actor_placement.expire_time < MonkeyTime.timestamp_sec():
            self.__cache.pop(self.placement_key(actor_type, actor_id))
            return None
        return MembershipManager().get_member(actor_placement.server_id)

    async def find_position(self, actor_type: Type[ActorType], actor_id: str) -> ServerNode | None:
        node = self.find_position_in_cache(actor_type, actor_id)
        if node is not None and node.session is not None:
            return node
        node = MembershipManager().choose_member(actor_type.actor_type())
        if node is None:
            return None
        async_script = self.__redis_client.register_script(
            RedisScript.find_actor_position_lua())
        server_id = await async_script(keys=[actor_type.actor_type(), actor_id], args=[node.server_id, 120])
        if isinstance(server_id, bytes):
            server_id = server_id.decode('utf-8')
        node = MembershipManager().get_member(server_id)
        if node is not None:
            self.__cache.put(
                self.placement_key(actor_type, actor_id),
                ActorPlacement(actor_id=actor_id,
                               actor_type=actor_type.actor_type(),
                               server_id=server_id,
                               expire_time=MonkeyTime.timestamp_sec() + 120))
        return node

    def remove_position(self, actor_type: Type[ActorType], actor_id: str) -> None:
        """理论上不应该主动清理Redis中的数据,Actor下线时按道理,Redis中的数据会自动过期"""
        self.__cache.pop(self.placement_key(actor_type, actor_id))

    async def actor_keep_alive(self, actor_type: Type[ActorType], actor_id: str, sec: int) -> bool:
        actor_placement = self.__cache.get(
            self.placement_key(actor_type, actor_id))
        if actor_placement is None:
            return False
        async_script = self.__redis_client.register_script(
            RedisScript.actor_keep_alive_lua())
        result = await async_script(keys=[actor_type.actor_type(), actor_id], args=[actor_placement.server_id, sec])
        if isinstance(result, bytes):
            result = result.decode('utf-8')
        if result == 'success':
            actor_placement.expire_time = MonkeyTime.timestamp_sec() + sec
            return True
        self.remove_position(actor_type, actor_id)
        return False

    async def __try_send_heart_beat(self):
        head_beat = RequestHeartBeat(now_sec=MonkeyTime.timestamp_sec())
        for node in MembershipManager().get_members():
            if node.session is None or node.session.is_closed:
                asyncio.create_task(self.__try_connect(node))
                continue
            try:
                await node.session.send(head_beat)
            except Exception as e:
                logger.error(
                    f'RedisPlacement __try_send_heart_beat Send heart beat to {node.address}:{node.port} failed, error:{e}')

    async def __heart_beat_loop(self):
        while True:
            try:
                await asyncio.sleep(monkey_config.get_config().tcp_ttl // 3)
                await self.__try_send_heart_beat()
            except Exception as e:
                logger.error(
                    f'RedisPlacement __heart_beat_loop heart beat error:{e}')

    @classmethod
    async def __try_connect(cls, node: ServerNode):
        begin = time.time()
        try:
            session = await TcpSocketSession.connect(node.address, node.port, CodecRpc)
            if session is not None:
                node.session = session
                logger.info(
                    f'RedisPlacement __try_connect Connect to {node.address}:{node.port} success')
        except Exception as e:
            end = time.time()
            logger.error(
                f'RedisPlacement __try_connect Connect to {node.address}:{node.port} failed, cost {end - begin}s, error: {e}')

    async def check_health(self) -> bool:
        suc: bool = await self.__redis_client.ping()
        return suc
