// client.go
package main

import (
 // "fmt"
 // "github.com/cenkalti/rpc2"
 // "net"
 "client_common"
)

// type Args struct{ A, B int }
// type Reply int

// type OT2Client struct {
// 	rpc_client *rpc2.Client
// 	uid        int64
// 	version    int // client side sent
// 	versionS   int // client side received
// }

// func NewOT2Client() *OT2Client {
// 	cl := &OT2Client{}
// 	conn, _ := net.Dial("tcp", "127.0.0.1:5000")

// 	cl.rpc_client = rpc2.NewClient(conn)
// 	cl.rpc_client.Handle("mult", func(client *rpc2.Client, args *Args, reply *Reply) error {
// 		*reply = Reply(args.A * args.B)
// 		return nil
//    	})
//    	go cl.rpc_client.Run() // runs the client

// 	return cl
// }

func main(){
	// conn, _ := net.Dial("tcp", "127.0.0.1:5000")

	// clt := rpc2.NewClient(conn)
	// clt.Handle("mult", func(client *rpc2.Client, args *Args, reply *Reply) error {
	// 	*reply = Reply(args.A * args.B)
	// 	return nil
 //   	})
 //   	go clt.Run()
	cl := client_common.NewOTClient()
	cl.Version = 1
	for {
		
	}

    // var rep bool
    // cl.Rpc_client.Call("Init", 
    // 	3, 
    // 	&rep)
    // fmt.Println("add result:", rep)
   }