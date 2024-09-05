# -*- coding= utf-8 -*-

__time__ = '2024/08/11'
__author__ = '虎小黑'

import json
import yaml
import random
from abc import ABC, abstractmethod
from typing import Type, TypeVar


class MonkeyConfig(ABC):

    @property
    @abstractmethod
    def name(self) -> str:
        pass

    @property
    @abstractmethod
    def log_prefix(self) -> str:
        pass

    @property
    @abstractmethod
    def log_rotation(self) -> str:
        pass

    @property
    @abstractmethod
    def log_format(self) -> str:
        pass

    @property
    @abstractmethod
    def log_level(self) -> str:
        pass

    @property
    @abstractmethod
    def services(self) -> dict[str, str]:
        pass

    @property
    @abstractmethod
    def tcp_address(self) -> int:
        pass

    @property
    @abstractmethod
    def tcp_port(self) -> int:
        pass

    @property
    @abstractmethod
    def tcp_ttl(self) -> int:
        pass

    @property
    @abstractmethod
    def tcp_buffer_size(self) -> int:
        pass

    @property
    @abstractmethod
    def tcp_buffer_max_size(self) -> int:
        pass

    @property
    @abstractmethod
    def tcp_session_timeout(self) -> int:
        pass

    @property
    @abstractmethod
    def tcp_window_size(self) -> int:
        pass

    @property
    @abstractmethod
    def rpc_timeout(self) -> int:
        pass

    @property
    @abstractmethod
    def socket_gc_interval(self) -> int:
        pass

    @abstractmethod
    def parse(self, file_path: str) -> None:
        pass

    @property
    @abstractmethod
    def magic_code(self) -> str:
        pass

    @property
    @abstractmethod
    def consul_namespace(self) -> str:
        pass

    @property
    @abstractmethod
    def consul_address(self) -> str:
        pass

    @property
    @abstractmethod
    def consul_try_times(self) -> int:
        pass

    @property
    @abstractmethod
    def consul_token(self) -> str:
        pass

    @property
    @abstractmethod
    def consul_tag(self) -> list[str]:
        pass

    @property
    @abstractmethod
    def consul_support_ns(self) -> bool:
        pass


class DefaultMonkeyConfig(MonkeyConfig):

    def __init__(self) -> None:
        super().__init__()
        self.__name = 'monkey'
        self.__log_prefix = 'monkey'
        self.__log_rotation = '00:00:00'
        self.__log_format = r'{time:YYYY-MM-DD HH:mm:ss.SSSS} | {extra[tag]} | {level} | {message}'
        self.__log_level = 'DEBUG'
        self.__services = {}
        self.__tcp_address = '127.0.0.1'
        self.__tcp_port = 8080
        self.__tcp_ttl = 60
        self.__tcp_buffer_size = 1024
        self.__tcp_buffer_max_size = 1024 * 4
        self.__tcp_session_timeout = 60
        self.__tcp_window_size = 1024 * 4
        self.__rpc_timeout = 5
        self.__socket_gc_interval = 30
        self.__magic_code = 'Monkey'
        self.__consul_namespace = 'MonkeyDev'
        self.__consul_address = ['http://127.0.0.1:8500']
        self.__consul_try_times = 3
        self.__consul_token = ''
        self.__consul_tag = []
        self.__consul_support_ns = False

    @property
    def name(self) -> str:
        return self.__name

    @property
    def log_prefix(self) -> str:
        return self.__log_prefix

    @property
    def log_rotation(self) -> str:
        return self.__log_rotation

    @property
    def log_format(self) -> str:
        return self.__log_format

    @property
    def log_level(self) -> str:
        return self.__log_level

    @property
    def services(self) -> dict[str, str]:
        return self.__services

    @property
    def tcp_address(self) -> str:
        return self.__tcp_address

    @property
    def tcp_port(self) -> int:
        return self.__tcp_port

    @property
    def tcp_ttl(self) -> int:
        return self.__tcp_ttl

    @property
    def tcp_buffer_size(self) -> int:
        return self.__tcp_buffer_size

    @property
    def tcp_buffer_max_size(self) -> int:
        return self.__tcp_buffer_max_size

    @property
    def tcp_session_timeout(self) -> int:
        return self.__tcp_session_timeout

    @property
    def tcp_window_size(self) -> int:
        return self.__tcp_window_size

    @property
    def rpc_timeout(self) -> int:
        return self.__rpc_timeout

    @property
    def socket_gc_interval(self) -> int:
        return self.__socket_gc_interval

    @property
    def magic_code(self) -> str:
        return self.__magic_code

    @property
    def consul_namespace(self) -> str:
        return self.__consul_namespace

    @property
    def consul_address(self) -> str:
        return random.sample(self.__consul_address, 1)[0]

    @property
    def consul_try_times(self) -> int:
        return self.__consul_try_times

    @property
    def consul_token(self) -> str:
        return self.__consul_token

    @property
    def consul_tag(self) -> list[str]:
        return self.__consul_tag

    @property
    def consul_support_ns(self) -> bool:
        return self.__consul_support_ns

    @classmethod
    def _load_config(cls, file_path: str) -> dict:
        return cls._load_config_as_json(file_path)

    @classmethod
    def _load_config_as_json(cls, file_path: str) -> dict:
        if file_path.endswith('.json'):
            with open(file_path, 'r') as file:
                return json.load(file)
        elif file_path.endswith('.yaml') or file_path.endswith('.yml'):
            with open(file_path, 'r') as file:
                return yaml.safe_load(file)
        else:
            raise ValueError('Unsupported file format')

    def parse(self, file_path: str) -> None:
        server_config = self._load_config(file_path)
        if 'name' in server_config:
            self.__name = server_config['name']
        if 'logPrefix' in server_config:
            self.__log_prefix = server_config['logPrefix']
        if 'logRotation' in server_config:
            self.__log_rotation = server_config['logRotation']
        if 'logFormat' in server_config:
            self.__log_format = server_config['logFormat']
        if 'logLevel' in server_config:
            self.__log_level = server_config['logLevel']
        if 'services' in server_config:
            self.__services = server_config['services']
        if 'tcpAddress' in server_config:
            self.__tcp_address = server_config['tcpAddress']
        if 'tcpPort' in server_config:
            self.__tcp_port = server_config['tcpPort']
        if 'tcpTTL' in server_config:
            self.__tcp_ttl = server_config['tcpTTL']
        if 'tcpBufferSize' in server_config:
            self.__tcp_buffer_size = server_config['tcpBufferSize']
        if 'tcpBufferMaxSize' in server_config:
            self.__tcp_buffer_max_size = server_config['tcpBufferMaxSize']
        if 'tcpSessionTimeout' in server_config:
            self.__tcp_session_timeout = server_config['tcpSessionTimeout']
        if 'tcpWindowSize' in server_config:
            self.__tcp_window_size = server_config['tcpWindowSize']
        if 'rpcTimeout' in server_config:
            self.__rpc_timeout = server_config['rpcTimeout']
        if 'socketGCInterval' in server_config:
            self.__socket_gc_interval = server_config['socketGCInterval']
        if 'magicCode' in server_config:
            self.__magic_code = server_config['magicCode']
        if 'consulNamespace' in server_config:
            self.__consul_namespace = server_config['consulNamespace']
        if 'consulAddress' in server_config:
            self.__consul_address = server_config['consulAddress']
        if 'consulTryTimes' in server_config:
            self.__consul_try_times = server_config['consulTryTimes']
        if 'consulToken' in server_config:
            self.__consul_token = server_config['consulToken']
        if 'consulTag' in server_config:
            self.__consul_tag = server_config['consulTag']
        if 'consulSupportNS' in server_config:
            self.__consul_support_ns = server_config['consulSupportNS']


ConfigType = TypeVar('ConfigType', bound=MonkeyConfig)
_config: MonkeyConfig | None = None


def get_config() -> MonkeyConfig:
    global _config
    if _config is None:
        _config = DefaultMonkeyConfig()
    return _config


def set_config_impl(Config: Type[ConfigType]) -> None:
    global _config
    _config = Config()
