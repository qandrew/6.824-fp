package main

import (
	"log"
	"net/rpc"
	"op"
)

type OTClient struct {
	rpc_client *rpc.Client
	version		int 		// most up to date version

}

func NewOTClient() *OTClient {
	cl := &OTClient{}
	rpc_client, err := rpc.Dial("tcp", "localhost:42586")
	cl.rpc_client = rpc_client
	if err != nil {
		log.Fatal(err)
	}

	return cl
}

func (cl *OTClient) Insert(ch rune, pos int) {
	args := op.Op{"ins", pos, 0, string(ch)} // version?
	reply := cl.SendOp(&args)
	cl.version = reply.Version // temporary
}

func (cl *OTClient) Delete(pos int) {
	if pos != 0 { // can't delete first 
		args := op.Op{"del", pos, 0, ""} // version?
		reply := cl.SendOp(&args)
		cl.version = reply.Version
	}
}

func (cl *OTClient) SendOp(args *op.Op) op.Op {
	var reply op.Op
	err := cl.rpc_client.Call("OTServer.ApplyOp", args, &reply)
	if err != nil {
		log.Fatal(err)
	} 
	return reply
}

func main() {
	client := NewOTClient()
	start_ui(client)
}
