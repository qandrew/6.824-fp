package client_common

import (
	"crypto/rand"
	"log"
	"math/big"
	// "net/rpc"
	"net"
	"github.com/cenkalti/rpc2"
	"op"
	"fmt"
)

type OTClient struct {
	Rpc_client *rpc2.Client
	uid        int64
	Version    int // client side sent
	versionS   int // client side received
}
type Args struct{ A, B int }
type Reply int

func NewOTClient() *OTClient {
	cl := &OTClient{}
	conn, _ := net.Dial("tcp", "127.0.0.1:5000")

	cl.Rpc_client = rpc2.NewClient(conn)
	cl.uid = nrand()
	cl.Version = 1

	cl.Rpc_client.Handle("ReceiveOp", func(client *rpc2.Client, args int, ack *bool) error {
		fmt.Println("client processed", args)	
		// if args.Uid == cl.uid {
		// 	return cl.receiveOp(args)
		// }
		// *ack = true
		return nil
   	})

   	go cl.Rpc_client.Run() // runs the client

   	var rep bool
   	cl.Rpc_client.Call("Init", cl.uid, &rep)
   	fmt.Println("init received:", rep)
	return cl
}

// to be deprecated

// type OTClient struct {
// 	rpc_client *rpc.Client
// 	uid        int64
// 	version    int // client side sent
// 	versionS   int // client side received
// }

// func NewOTClient() *OTClient {
// 	cl := &OTClient{}
// 	rpc_client, err := rpc.Dial("tcp", "localhost:42586")
// 	cl.rpc_client = rpc_client
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	cl.uid = nrand()
// 	cl.version = 1
// 	var ack bool
// 	cl.rpc_client.Call("OTServer.Init", cl.uid, &ack)

// 	return cl
// }

func (cl *OTClient) GetSnapshot() string {
	snap := op.Snapshot{}
	snapIn := op.Snapshot{}
	// cl.Rpc_client.Call("OTServer.GetSnapshot", &snapIn, &snap)
	cl.Rpc_client.Call("GetSnapshot", &snapIn, &snap)
	return snap.Value
}

func (cl *OTClient) receiveOp(args *op.Op) error {
	fmt.Println(args)
	return nil
}

func (cl *OTClient) Insert(ch rune, pos int) {
	args := op.Op{"ins", pos, cl.Version, cl.versionS, cl.uid, string(ch)} // version?
	reply := cl.SendOp(&args)
	cl.Version = reply.Version // temporary
}

func (cl *OTClient) Delete(pos int) {
	if pos != 0 { // can't delete first
		args := op.Op{"del", pos, cl.Version, cl.versionS, cl.uid, ""} // version?
		reply := cl.SendOp(&args)
		cl.Version = reply.Version
	}
}

func (cl *OTClient) SendOp(args *op.Op) op.Op {
	var reply op.Op
	// err := cl.rpc_client.Call("OTServer.ApplyOp", args, &reply)
	err := cl.Rpc_client.Call("ApplyOp", args, &reply)
	if err != nil {
		log.Fatal(err)
	}
	return reply
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 40) // was 62
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}
