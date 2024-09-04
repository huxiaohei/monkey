# -*- coding= utf-8 -*-

__time__ = '2024/08/12'
__author__ = '虎小黑'

import json
from logger.logger import Logger

logger = Logger().get_logger('Monkey')

json_dumps = json.dumps
json_loads = json.loads

try:
    import orjson

    json_dumps = orjson.dumps
    json_loads = orjson.loads
except ImportError as e:
    logger.error(f'Import orjson failed: {e}')
