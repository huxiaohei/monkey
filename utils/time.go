package utils

import "time"

func GetNow() time.Time {
	return time.Now()
}

func GetNowSec() int64 {
	return time.Now().Unix()
}

func GetNowMs() int64 {
	return time.Now().UnixMilli()
}

func SleepSec(sec int64) {
	if sec < 0 {
		sec = 0
	}
	time.Sleep(time.Duration(sec) * time.Second)
}

func SleepMs(ms int64) {
	if ms < 0 {
		ms = 0
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func SleepUs(us int64) {
	if us < 0 {
		us = 0
	}
	time.Sleep(time.Duration(us) * time.Microsecond)
}
