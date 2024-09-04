# -*- coding= utf-8 -*-

__time__ = '2024/08/16'
__author__ = '虎小黑'


import aiohttp
import asyncio
import weakref
from http import HTTPStatus
from utils import monkey_config
from logger.logger import Logger
from utils.monkey_time import MonkeyTime
from membership.membership import Membership
from membership.server_node import ServerNode
from network.socket_session import SocketSession
from pydantic import BaseModel, PrivateAttr, Field
from membership.membership_manager import MembershipManager


logger = Logger().get_logger('Monkey')


class Check(BaseModel):
    node: str = Field(..., alias='Node', description='节点id')
    check_id: str = Field(..., alias='CheckID', description='检查id')
    name: str = Field(..., alias='Name', description='检查名称')
    status: str = Field(..., alias='Status', description='检查状态')
    notes: str = Field('', alias='Notes', description='备注')
    output: str = Field('', alias='Output', description='输出')
    service_id: str = Field(..., alias='ServiceID', description='服务id')
    service_name: str = Field(..., alias='ServiceName', description='服务名称')
    service_tags: list[str] = Field(
        [], alias='ServiceTags', description='服务标签')
    type: str = Field('', alias='Type', description='检查类型')
    interval: str = Field('', alias='Interval', description='检查间隔')
    timeout: str = Field('', alias='Timeout', description='超时时间')
    exposed_port: int = Field(0, alias='ExposedPort', description='暴露端口')
    definition: dict[str, str] = Field(
        {}, alias='Definition', description='定义')
    create_index: int = Field(0, alias='CreateIndex', description='创建索引')
    modify_index: int = Field(0, alias='ModifyIndex', description='修改索引')


class Node(BaseModel):
    node_id: str = Field(..., alias='ID', description='节点id')
    node_name: str = Field(..., alias='Node', description='节点名称')
    address: str = Field(..., alias='Address', description='节点地址')
    datacenter: str = Field('', alias='Datacenter', description='数据中心')
    tagged_addresses: dict[str, str] = Field(
        {}, alias='TaggedAddresses', description='标记地址')
    meta: dict[str, str] = Field(
        {}, alias='Meta', description='元数据')
    create_index: int = Field(0, alias='CreateIndex', description='创建索引')
    modify_index: int = Field(0, alias='ModifyIndex', description='修改索引')


class Server(BaseModel):
    server_id: str = Field(..., alias='ID', description='服务唯一id')
    name: str = Field(..., alias='Service', description='服务名称')
    tags: list[str] = Field([], alias='Tags', description='服务标签,用于服务分组')
    address: str = Field(..., alias='Address', description='服务地址')
    port: int = Field(..., alias='Port', description='服务端口')
    meta: dict[str, str] = Field(
        {}, alias='Meta', description='服务元数据,用于存储服务上的业务信息,比如服务上提供哪些actor')
    weights: dict[str, int] = Field(
        {'Passing': 10, 'Warning': 0}, alias='Weights', description='服务权重')
    create_index: int = Field(0, alias='CreateIndex', description='创建索引')
    modify_index: int = Field(0, alias='ModifyIndex', description='修改索引')


class ConsulServerNode(ServerNode):

    node: Node = Field(..., alias='Node', description='节点信息')
    server: Server = Field(..., alias='Service', description='服务信息')
    checks: list[Check] = Field(..., alias='Checks', description='检查信息')

    _session: None | weakref.ReferenceType[SocketSession] = PrivateAttr(None)
    _session_id: int = PrivateAttr(0)

    @property
    def server_id(self) -> str:
        return self.server.server_id

    @property
    def session_id(self) -> int:
        return self._session_id

    @property
    def session(self) -> None | SocketSession:
        return self._session() if self._session else None

    @session.setter
    def session(self, session: SocketSession) -> None:
        self._session = weakref.ref(session)
        self._session_id = session.session_id

    @property
    def address(self) -> str:
        return self.server.address

    @property
    def port(self) -> int:
        return self.server.port

    @property
    def weight(self) -> int:
        return 0

    @property
    def is_available(self) -> bool:
        if self.session is None:
            return False
        for node_check in self.checks:
            if node_check.check_id == self.server.server_id and \
                    node_check.status == 'passing':
                return True
        return False

    def is_support(self, mate: str) -> bool:
        return mate in self.server.meta


class ConsulPlacement(Membership):

    def __init__(self) -> None:
        super().__init__()
        self.__consul_index = -1

    async def register_server(self, namespace: str, name: str, address: str, port: int, tags: list[str], meta: dict[str, str]) -> bool:
        for _ in range(monkey_config.get_config().consul_try_times):
            try:
                logger.info(
                    f'ConsulPlacement register_server {namespace}-{name}-{address}-{port}')
                register_url = (
                    f'{monkey_config.get_config().consul_address}'
                    f'/v1/agent/service/register'
                )
                async with aiohttp.ClientSession() as session:
                    payload = {
                        'ID': f'{namespace}-{name}-{address}-{port}-{MonkeyTime.timestamp_sec()}',
                        'Name': name,
                        'Address': address,
                        'Port': port,
                        'Tags': tags,
                        'Meta': meta,
                        'EnableTagOverride': True,
                        'Check': {
                            'DeregisterCriticalServiceAfter': '15s',
                            'TCP': f'{address}:{port}',
                            'Interval': '3s',
                            'Timeout': '1s'
                        },
                        'Weights': {
                            'Passing': 10,
                            'Warning': 0
                        }
                    }
                    headers = {
                        'X-Consul-Namespace': namespace,
                        'Content-Type': 'application/json'
                    }
                    if not monkey_config.get_config().consul_support_ns:
                        del headers['X-Consul-Namespace']
                    async with session.put(register_url, json=payload, headers=headers) as response:
                        if response.status != HTTPStatus.OK:
                            logger.error(
                                f'ConsulPlacement register_server error {namespace}-{name}-{address}-{port} register_url:{register_url} response:{response}')
                            await asyncio.sleep(3)
                            continue
                        logger.info(
                            f'ConsulPlacement register_server success {namespace}-{name}-{address}-{port} register_url:{register_url}')
            except Exception as e:
                logger.exception(
                    f'ConsulPlacement register_server error {namespace}-{name}-{address}-{port} error:{e}')
                return False
            return True
        return False

    async def unregister_server(self, server_id: str) -> bool:
        for _ in range(monkey_config.get_config().consul_try_times):
            try:
                logger.info(
                    f'ConsulPlacement unregister_server server server_id:{server_id}')
                unregister_url = (
                    f'{monkey_config.get_config().consul_address}'
                    f'/v1/agent/service/deregister/{server_id}'
                )
                async with aiohttp.ClientSession() as session:
                    async with session.put(unregister_url) as response:
                        data = await response.json()
                        if response.status != HTTPStatus.OK:
                            logger.error(
                                f'ConsulPlacement unregister_server server error unregister_url:{unregister_url} response:{data}')
                            await asyncio.sleep(3)
                            continue
                        logger.info(
                            f'ConsulPlacement unregister_server server success unregister_url:{unregister_url} response:{data}')
            except Exception as e:
                logger.exception(
                    f'ConsulPlacement unregister_server server error server_id:{server_id} error:{e}')
                return False
            return True
        return False

    async def get_servers(self, namespace: str, server_tags: list[str] = []) -> list[ServerNode]:
        for _ in range(monkey_config.get_config().consul_try_times):
            try:
                logger.info(
                    f'ConsulPlacement get_servers namespace:{namespace}')
                services_url = (
                    f'{monkey_config.get_config().consul_address}'
                    f'/v1/catalog/services'
                )
                server_names = []
                async with aiohttp.ClientSession() as session:
                    headers = {
                        'X-Consul-Namespace': namespace,
                        'Content-Type': 'application/json'
                    }
                    if not monkey_config.get_config().consul_support_ns:
                        del headers['X-Consul-Namespace']
                    async with session.get(services_url, headers=headers) as response:
                        server_names_data: dict[str, list[str]] = await response.json()
                        if response.status != HTTPStatus.OK:
                            logger.error(
                                f'ConsulPlacement get_servers error namespace:{namespace} services_url:{services_url} response:{server_names_data}')
                            continue
                        self.__consul_index = response.headers.get(
                            'X-Consul-Index')
                        logger.info(
                            f'ConsulPlacement get_servers success namespace:{namespace} services_url:{services_url} response:{server_names_data} consul_index:{self.__consul_index}')
                        for server_name, tags in server_names_data.items():
                            if not server_tags or any(tag in tags for tag in server_tags):
                                server_names.append(server_name)

                    servers: list[ServerNode] = []
                    for server_name in server_names:
                        server_url = (
                            f'{monkey_config.get_config().consul_address}'
                            f'/v1/health/service/Monkey'
                        )
                        async with session.get(server_url, headers=headers) as response:
                            server_data: list = await response.json()
                            if response.status != HTTPStatus.OK:
                                logger.error(
                                    f'ConsulPlacement get_servers error namespace:{namespace} server_name:{server_name} response:{server_data}')
                                continue
                            servers.extend(
                                [ConsulServerNode(**json_data) for json_data in server_data])
                    return servers
            except Exception as e:
                logger.exception(
                    f'ConsulPlacement get_servers error error:{e}')
        return []

    async def check_health(self, namespace: str, server_tags: list[str] = []) -> None:

        server_nodes = await self.get_servers(namespace, server_tags)
        MembershipManager().refresh_members(server_nodes)

        while True:
            health_node_url = (
                f'{monkey_config.get_config().consul_address}'
                f'/v1/catalog/services?wait=30s&index={self.__consul_index}'
            )
            headers = {
                'X-Consul-Namespace': namespace,
                'Content-Type': 'application/json'
            }
            if not monkey_config.get_config().consul_support_ns:
                del headers['X-Consul-Namespace']

            async with aiohttp.ClientSession() as session:
                async with session.get(health_node_url, headers=headers) as response:
                    data = await response.json()
                    if response.status != HTTPStatus.OK:
                        logger.error(
                            f'ConsulPlacement run get health node error health_node_url:{health_node_url} response:{data}')
                        await asyncio.sleep(3)
                    else:
                        logger.info(
                            f'ConsulPlacement run get health node success health_node_url:{health_node_url} response:{data}')
                        consul_index = response.headers.get('X-Consul-Index')
                        if self.__consul_index != consul_index:
                            server_nodes = await self.get_servers(namespace, server_tags)
                            MembershipManager().refresh_members(server_nodes)
