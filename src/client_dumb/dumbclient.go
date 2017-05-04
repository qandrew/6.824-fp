package main

import (
	"client_common"
	"fmt"
	"time"
)

func main() {
	cl := client_common.NewOTClient()
	cl.Debug = true
	if cl.Debug{
		fmt.Println("Debug")
	}
	for {
		time.Sleep(2*client_common.SLEEP*time.Millisecond)
		// cl.RandOp()
	}
}
