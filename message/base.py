# -*- coding= utf-8 -*-

__time__ = '2024/08/11'
__author__ = '虎小黑'


import inspect
from utils import utils
from logger.logger import Logger
from dataclasses import dataclass


logger = Logger().get_logger('Monkey')


__json_models: dict[bytes, type['JsonMessage']] = {}


def register_model(cls: type) -> None:
    global __json_models
    __json_models[cls.__qualname__.encode()] = cls


def find_model(name: bytes) -> type['JsonMessage'] | None:
    global __json_models
    return __json_models.get(name, None)


class JsonMeta(type):

    def __new__(cls, class_name, class_parents, class_attr):
        cls = type.__new__(cls, class_name, class_parents, class_attr)
        register_model(cls)
        return cls


@dataclass(slots=True)
class JsonMessage(metaclass=JsonMeta):
    msg_id: int = 0
    name: str = ''

    @classmethod
    def from_dict(cls, kwargs: dict) -> 'JsonMessage':
        try:
            return cls(**kwargs)
        except Exception as e:
            logger.exception(f"JsonMessage from_dict error:{e}")
            parameters = inspect.signature(cls).parameters
            return cls(**{k: kwargs[k] for k in parameters if k in kwargs})

    def to_dict(self) -> dict:
        return utils.to_dict(self)
