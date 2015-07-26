package main

import (
	"fmt"
	"github.com/cihub/seelog"
	"os"
)

func main() {

	defer seelog.Flush()

	InitLogger()

	LoadConfig()

	InitRedisPool()

	seelog.Info("Start")

	FetchEmailTasksFromRedis()
}

func InitLogger() {
	logger, err := seelog.LoggerFromConfigAsFile("seelog.xml")
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	seelog.ReplaceLogger(logger)
}
