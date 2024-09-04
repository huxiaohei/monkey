# -*- coding= utf-8 -*-

__time__ = '2024/09/02'
__author__ = '虎小黑'

import unittest
from redis.asyncio import Redis
from logger.logger import Logger
from utils.redis_script import RedisScript

logger = Logger().get_logger('Monkey')


class TestRedisPlacement(unittest.IsolatedAsyncioTestCase):

    async def test_find_actor_position_lua(self):
        client = Redis.from_url('redis://localhost:6379/0')

        pong: bool = await client.ping()
        assert pong

        await client.delete('player:10001')

        async_script = client.register_script(
            RedisScript.find_actor_position_lua())
        res = await async_script(keys=['player', '10001'],
                                 args=['10001', 120])
        if isinstance(res, bytes):
            res = res.decode('utf-8')
            assert res == '10001'
        else:
            assert False

        res = await async_script(keys=['player', '10001'],
                                 args=['10002', 120])
        if isinstance(res, bytes):
            res = res.decode('utf-8')
            assert res == '10001'
        else:
            assert False

        await client.close()

    async def test_actor_keep_alive_lua(self):
        client = Redis.from_url('redis://localhost:6379/0')

        pong: bool = await client.ping()
        assert pong

        await client.delete('player:10002')

        async_keep_alive_script = client.register_script(
            RedisScript.actor_keep_alive_lua())
        res = await async_keep_alive_script(keys=['player', '10002'],
                                            args=['10002', 120])
        if isinstance(res, bytes):
            res = res.decode('utf-8')
            assert res == 'fail'
        else:
            assert False

        async_find_position_script = client.register_script(
            RedisScript.find_actor_position_lua())
        res = await async_find_position_script(keys=['player', '10002'],
                                               args=['10002', 120])
        if isinstance(res, bytes):
            res = res.decode('utf-8')
            assert res == '10002'
        else:
            assert False

        res = await async_keep_alive_script(keys=['player', '10002'],
                                            args=['10002', 120])
        if isinstance(res, bytes):
            res = res.decode('utf-8')
            assert res == 'success'
        else:
            assert False

        res = await async_keep_alive_script(keys=['player', '10002'],
                                            args=['10003', 120])
        if isinstance(res, bytes):
            res = res.decode('utf-8')
            assert res == 'fail'
        else:
            assert False

        await client.close()
