# -*- coding= utf-8 -*-

__time__ = '2024/09/01'
__author__ = '虎小黑'


from utils.sequence_id import SequenceId


class RpcRequestId(object):

    __request_id = SequenceId()
    __reentrant_id = SequenceId()

    @classmethod
    def get_request_id(cls) -> int:
        return cls.__request_id.new_sequence_id()

    @classmethod
    def get_reentrant_id(cls) -> int:
        return cls.__reentrant_id.new_sequence_id()
