# -*- coding= utf-8 -*-

__time__ = '2024/09/02'
__author__ = '虎小黑'


class RedisScript(object):

    __find_actor_position_lua = '''
local actor_type = KEYS[1]
local actor_id = KEYS[2]
local server_id = ARGV[1]
local expire_time = ARGV[2]
local actor_key = actor_type .. ':' .. actor_id
local cnt = redis.call('setnx', actor_key, server_id)
if cnt == 1 then
    redis.call('expire', actor_key, expire_time)
    return server_id
end
return redis.call('get', actor_key)
'''

    __actor_keep_alive_lua = '''
local actor_type = KEYS[1]
local actor_id = KEYS[2]
local server_id = ARGV[1]
local expire_time = ARGV[2]
local actor_key = actor_type .. ':' .. actor_id
local old_server_id = redis.call('get', actor_key)
if old_server_id == server_id then
    redis.call('expire', actor_key, expire_time)
    return 'success'
end
return 'fail'
'''

    @classmethod
    def find_actor_position_lua(cls) -> str:
        return cls.__find_actor_position_lua

    @classmethod
    def actor_keep_alive_lua(cls) -> str:
        return cls.__actor_keep_alive_lua
