package main

import (
	"client_common"
	"os"
)

func main() {
	client := client_common.NewOTClient()
	clientId := "0"
	if len(os.Args) > 1 {
		clientId = os.Args[1]
	}
	client.SetLogFile("client-" + clientId + ".log")
	client.Debug = true
	start_ui(client)
}
