# -*- coding= utf-8 -*-

__time__ = '2024/09/01'
__author__ = '虎小黑'


class GatewayMessageDispatch(object):

    def __init__(self) -> None:
        self.__message_handlers = {}

    def add_message_handler(self, message_type: int, handler) -> None:
        self.__message_handlers[message_type] = handler

    def get_message_handler(self, message_type: int):
        return self.__message_handlers.get(message_type, None)
