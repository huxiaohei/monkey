# -*- coding= utf-8 -*-

__time__ = '2024/08/31'
__author__ = '虎小黑'

import asyncio
import weakref
from actor.actor import Actor
from typing import Callable, Any
from logger.logger import Logger
from utils.sequence_id import SequenceId


logger = Logger().get_logger('Monkey')


class ActorTimer(object):

    def __init__(
            self,
            timer_id: int,
            weaker_manager: weakref.ReferenceType['ActorTimerManager'],
            weak_actor: weakref.ReferenceType[Actor],
            delay: int,
            interval: int,
            repetition: int,
            fn: Callable[..., None],
            target: object,
            *args: Any,
            **kwargs: Any

    ) -> None:
        super().__init__()
        self.__timer_id = timer_id
        self.__weaker_manager = weaker_manager
        self.__weak_actor = weak_actor
        self.__delay = delay if delay > 0 else 0
        self.__interval = interval if interval > 0 else 0
        self.__repetition = repetition
        self.__fn = fn
        self.__target = target
        self.__args = args
        self.__kwargs = kwargs
        self.__is_cancel = False
        self.__tick_cnt = 0

    @property
    def timer_id(self) -> int:
        return self.__timer_id

    @property
    def interval(self) -> int:
        return self.__interval

    @property
    def repetition(self) -> int:
        return self.__repetition

    @property
    def is_cancel(self) -> bool:
        return self.__is_cancel

    @property
    def tick_cnt(self) -> int:
        return self.__tick_cnt

    def cancel(self) -> None:
        self.__is_cancel = True

    def tick(self) -> None:
        if self.__is_cancel:
            return
        try:
            self.__tick_cnt += 1
            self.__fn(self.__target, *self.__args, **self.__kwargs)
        except Exception as e:
            logger.exception(
                f'ActorTimer tick error, timer_id:{self.__timer_id}, error:{e}')
        finally:
            manager = self.__weaker_manager()
            if not manager:
                return
            manager.__register_timer(self)

    def next_tick_time(self) -> int:
        if self.__repetition > 0 and self.__tick_cnt >= self.__repetition:
            self.__is_cancel = True
            return 0
        if self.__tick_cnt > 0:
            return self.__interval
        return self.__delay

    def run(self):
        actor = self.__weak_actor()
        if actor and actor.context:
            asyncio.create_task(actor.context.push_message(self))
        else:
            manager = self.__weaker_manager()
            if manager:
                manager.unregister_timer(self.__timer_id)


class ActorTimerManager(object):

    def __init__(self, weak_actor: weakref.ReferenceType[Actor]) -> None:
        self.__seq_id = SequenceId()
        self.__weak_actor = weak_actor
        self.__timers: dict[int, ActorTimer] = {}

    @classmethod
    async def __run_timer(cls, sleep: int, timer: ActorTimer) -> None:
        if sleep > 0:
            await asyncio.sleep(sleep)
        timer.run()

    def __register_timer(self, timer: ActorTimer) -> None:
        asyncio.create_task(self.__run_timer(timer.next_tick_time(), timer))

    def register_timer(
            self,
            delay: int,
            interval: int,
            repetition: int,
            fn: Callable[..., None],
            target: object,
            *args: Any,
            **kwargs: Any) -> ActorTimer:
        timer = ActorTimer(
            self.__seq_id.new_sequence_id(),
            weakref.ref(self),
            self.__weak_actor,
            delay,
            interval,
            repetition,
            fn,
            target,
            *args,
            **kwargs)
        self.__timers[timer.timer_id] = timer
        self.__register_timer(timer)
        return timer

    def unregister_timer(self, timer_id: int) -> None:
        timer = self.__timers.pop(timer_id, None)
        if timer:
            timer.cancel()

    def unregister_all_timer(self) -> None:
        for timer_id in self.__timers.keys():
            self.unregister_timer(timer_id)
