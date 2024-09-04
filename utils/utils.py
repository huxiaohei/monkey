# -*- coding= utf-8 -*-

__time__ = '2024/08/12'
__author__ = '虎小黑'

import pickle
from typing import Any, cast
from logger.logger import Logger
from zstd import ZSTD_compress, ZSTD_uncompress

logger = Logger().get_logger('Monkey')

THRESHOLD = 256
COMPRESSED = b"1"
UNCOMPRESSED = b"0"


def to_dict(obj: object) -> dict:
    try:
        if isinstance(obj, dict):
            return {k: to_dict(v) for k, v in obj.items()}
        elif getattr(obj, '__slots__', None):
            return {
                k: to_dict(getattr(obj, k)) for k in getattr(obj, '__slots__') if isinstance(k, str) and not k.startswith('_')
            }
        elif getattr(obj, '__dict__', None):
            return {k: to_dict(v) for k, v in getattr(obj, '__dict__', {}).items() if isinstance(k, str) and not k.startswith('_')}
        else:
            return cast(dict, obj)
    except Exception as e:
        logger.exception('to_dict error:', e)
        return {}


def pickle_dump(o: Any) -> bytes:
    array = pickle.dumps(o, protocol=pickle.HIGHEST_PROTOCOL)
    compressed = UNCOMPRESSED + array
    if len(compressed) > THRESHOLD:
        compressed = COMPRESSED + ZSTD_compress(array)
    return compressed


def pickle_load(data: bytes) -> Any:
    if data[0:1] == COMPRESSED:
        return pickle.loads(ZSTD_uncompress(data[1:]))
    return pickle.loads(data[1:])
