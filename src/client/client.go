package main

import (
	"client_common"
	"os"
)

func main() {
	client := client_common.NewOTClient()
	client.SetLogFile("client-" + os.Args[1] + ".log")
	client.Debug = true
	start_ui(client)
}
