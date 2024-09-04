# -*- coding= utf-8 -*-

__time__ = '2024/08/31'
__author__ = '虎小黑'


import inspect
from types import ModuleType
from logger.logger import Logger
from actor.actor_base import ActorBase
from typing import Any, Callable, Type, TypeVar
from actor.actor_interface import ActorInterface


logger = Logger().get_logger('Monkey')

T = TypeVar('T')
Interface = TypeVar("Interface", bound=ActorInterface)


class RpcMeta(object):

    __actor_interface_set: set[Type] = set()
    __actor_interface_name_map: dict[str, Type] = {}
    __actor_impl_map: dict[Type, Type] = {}
    __actor_impl_name_map: dict[str, Type] = {}
    __actor_impl_method: dict[tuple[str, str], Callable] = {}

    @classmethod
    def register_actor_interface(cls, interface: Type[Interface]) -> None:
        cls.__actor_interface_set.add(interface)
        cls.__actor_interface_name_map[interface.__qualname__] = interface

    @classmethod
    def is_interface(cls, interface: Type) -> bool:
        return interface in cls.__actor_interface_set

    @classmethod
    def get_interface_type_by(cls, interface_name: str) -> Type | None:
        return cls.__actor_interface_name_map.get(interface_name, None)

    @classmethod
    def register_actor_impl(cls, interface: Type[Interface], impl: Type) -> None:
        cls.__actor_impl_map[interface] = impl
        cls.__actor_impl_name_map[impl.__qualname__] = impl

    @classmethod
    def get_actor_impl_type(cls, type: Type[T]) -> Type[T] | None:
        return cls.__actor_impl_map.get(type, None)

    @classmethod
    def get_actor_impl_type_by_name(cls, name: str) -> Type | None:
        return cls.__actor_impl_name_map.get(name, None)

    @classmethod
    def get_all_impl_types(cls) -> list[tuple[str, Type]]:
        rst: list[tuple[str, Type]] = []
        for name, impl in cls.__actor_impl_name_map.items():
            rst.append((name, impl))
        return rst

    @classmethod
    def get_actor_rpc_impl_method(cls, name: tuple[str, str]) -> Callable | None:
        if name in cls.__actor_impl_method:
            return cls.__actor_impl_method[name]
        impl_type = cls.get_actor_impl_type_by_name(name[0])
        if impl_type is None:
            return None
        fn = getattr(impl_type, name[1], None)
        if isinstance(fn, Callable):
            cls.__actor_impl_method[name] = fn
            return fn
        return None

    @classmethod
    def build_meta_info(cls, items: dict[str, Type | ModuleType]) -> None:
        interface_set: set[Type] = set()
        actor_type_set: set[Type] = set()
        ignore_set: set[Any] = set()

        def check_actor_meta_info(clz: Type):
            if clz == ActorInterface or clz == ActorBase:
                return
            if issubclass(clz, ActorInterface) and not issubclass(clz, ActorBase):
                interface_set.add(clz)
            if issubclass(clz, ActorBase):
                actor_type_set.add(clz)

        def check(val: Type | ModuleType):
            try:
                if isinstance(val, dict) or val in ignore_set:
                    return
                ignore_set.add(val)
            except Exception as e:
                logger.exception(
                    f'RpcMeta build_meta_info check error val:{val} error:{e}')

            if isinstance(val, ModuleType):
                members = inspect.getmembers(val, inspect.isclass)
                for _, v in members:
                    check_actor_meta_info(v)
                members = inspect.getmembers(val, inspect.ismodule)
                for _, v in members:
                    check(v)
            else:
                check_actor_meta_info(val)

        for _, val in sorted(list(items.items()), reverse=True):
            check(val)

        for interface_type in interface_set:
            for impl_type in actor_type_set:
                if not issubclass(impl_type, interface_type):
                    continue
                cls.__actor_interface_set.add(interface_type)
                cls.__actor_interface_name_map[interface_type.__qualname__] = interface_type
