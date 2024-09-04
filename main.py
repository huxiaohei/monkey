# -*- coding= utf-8 -*-

__time__ = '2024/09/04'
__author__ = '虎小黑'

from typing import Type, cast
from network.tcp_server import TcpServer
from network.codec_echo import CodecEcho
from network.event_handler import EventHandler
from network.socket_session import SocketSession


async def handle_str_msg(session: SocketSession, _: Type, msg: object):
    msg = cast(str, msg)
    await session.send(msg)


def main():

    EventHandler.register_hander(str, handle_str_msg)

    server = TcpServer()
    server.create_task(server.listen(
        host='127.0.0.1', port=8888, codec_type=CodecEcho))
    server.run()


if __name__ == '__main__':
    main()
