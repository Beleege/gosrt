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

var ShutDownSignal = make(chan error)

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
	go server.SetupUDPServer()
	// here need programmatically setup media server
	go server.SetupHLSServer()
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("srt server panic: ", r)
			time.Sleep(1 * time.Second)
		}
		log.Infof("srt server stop")
	}()

	printBanner()
	loadResource()
	preInit()
	serverInit()

	select {
	case err := <-ShutDownSignal:
		if err != nil {
			panic(err)
		}
	}
}
