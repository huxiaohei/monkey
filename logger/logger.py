# -*- coding= utf-8 -*-

__time__ = '2024/08/06'
__author__ = '虎小黑'


import psutil
from utils.singleton import Singleton
from loguru import logger as loguru_logger


class Logger(Singleton):

    def __init__(self) -> None:
        super().__init__()
        current_process = psutil.Process().name()
        loguru_logger.remove()
        loguru_logger.add(
            '%s_{time:YYYY_MM_DD}.log' % current_process,
            rotation='00:00:00',
            format="{time:YYYY-MM-DD HH:mm:ss.SSSS} | {extra[tag]} | {level} | {message}",
            level='DEBUG')

    def init_logger(self, prefix: str, rotation: str, format: str, level: str):
        loguru_logger.remove()
        loguru_logger.add(
            '%s_{time:YYYY_MM_DD}.log' % prefix,
            rotation=rotation,
            format=format,
            level=level)

    def get_logger(self, tag: str):
        return loguru_logger.bind(tag=tag)
