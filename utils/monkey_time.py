# -*- coding= utf-8 -*-

__time__ = '2024/08/15'
__author__ = '虎小黑'


import time


class MonkeyTime(object):

    @classmethod
    def timestamp_sec(cls) -> int:
        return int(time.time())

    @classmethod
    def timestamp_ms(cls) -> int:
        return int(time.time() * 1000)
