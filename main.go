package main

import (
	"fmt"
	"github.com/beleege/gosrt/server"
	logger "log"
	"time"

	"github.com/beleege/gosrt/config"
	"github.com/beleege/gosrt/util/log"
)

const (
	version = "0.1"
)

func printBanner() {
	logger.Println(fmt.Sprintf(`
===============================================
 _______  _______  _______  ______    _______ 
|       ||       ||       ||    _ |  |       |
|    ___||   _   ||  _____||   | ||  |_     _|
|   | __ |  | |  || |_____ |   |_||_   |   |  
|   ||  ||  |_|  ||_____  ||    __  |  |   |  
|   |_| ||       | _____| ||   |  | |  |   |  
|_______||_______||_______||___|  |_|  |___|
===============================================
version: %s`, version))
}

func loadResource() {
	config.InitConfig()
}

func preInit() {
	log.InitLog()
}

func serverInit() {
	server.SetupUDPServer()
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("srt server panic: ", r)
			time.Sleep(1 * time.Second)
		}
	}()

	printBanner()
	loadResource()
	preInit()
	serverInit()
}
