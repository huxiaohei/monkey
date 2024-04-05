package main

import (
	"monkey/logger"
	"monkey/pd"
	"monkey/pd/storage"
)

var (
	mlog, _ = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

func main() {

	mlog.Info("pd started")

	memoryStorage := storage.NewMemoryStorage()

	pd.Start(memoryStorage, memoryStorage, memoryStorage, "0.0.0.0:8000", "0.0.1")

	mlog.Info("pd stopped")
}
