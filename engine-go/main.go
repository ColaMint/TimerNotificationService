package main

import (
	"flag"
	"fmt"
	"github.com/cihub/seelog"
	"os"
)

var (
	logConfigPath    = flag.String("logconfig", "seelog.xml", "seelog config file path")
	engineConfigPath = flag.String("engineconfig", "config.yml", "engine config file path")
)

func main() {

	flag.Parse()

	defer seelog.Flush()

	InitLogger(*logConfigPath)

	LoadConfig(*engineConfigPath)

	InitRedisPool()

	engine := NewEngine()
	engine.AddTask(new(EmailTaskHandler), 3, 1)
	engine.Start()
}

func InitLogger(configPath string) {
	logger, err := seelog.LoggerFromConfigAsFile(configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	seelog.ReplaceLogger(logger)
}
