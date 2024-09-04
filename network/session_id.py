# -*- coding= utf-8 -*-

__time__ = '2024/08/12'
__author__ = '虎小黑'


from utils.sequence_id import SequenceId


_sequence = SequenceId()


def new_session_id() -> int:
    return _sequence.new_sequence_id()
