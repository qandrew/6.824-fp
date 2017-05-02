// server.go
package main

import (
 "net"
 "github.com/cenkalti/rpc2"
 // "fmt"
 // "client_common"
)

type Args struct{ A, B int }
type Reply int



func main(){
 srv := rpc2.NewServer()
 srv.Handle("add", func(client *rpc2.Client, args *Args, reply *Reply) error{
    // Reversed call (server to client)
    // var rep Reply
    // client.Call("mult", Args{2, 3}, &rep)
    // fmt.Println("mult result:", rep)

    *reply = Reply(args.A + args.B)
    return nil
 })

 lis, _ := net.Listen("tcp", "127.0.0.1:5000")
 srv.Accept(lis)
}