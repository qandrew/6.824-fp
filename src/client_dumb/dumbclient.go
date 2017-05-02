package main

import (
	"client_common"
	"fmt"
)

func main() {
	client := client_common.NewOTClient()
	client.Debug = true
	if client.Debug{
		fmt.Println("Debug")
	}
	for {
		
	}
}
