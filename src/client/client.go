package main

import (
	"log"
	"net/rpc"
	"op"
)

type OTClient struct {
	rpc_client *rpc.Client
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
	op := op.Op{"ins", pos, 0, string(ch)} // version?
	cl.SendOp(&op)
}

func (cl *OTClient) Delete(pos int) {
	op := op.Op{"del", pos, 0, ""} // version?
	cl.SendOp(&op)
}

func (cl *OTClient) SendOp(op *op.Op) bool {
	var reply bool
	err := cl.rpc_client.Call("Listener.ApplyOp", op, &reply)
	if err != nil {
		log.Fatal(err)
	}
	return reply
}

func main() {
	client := NewOTClient()
	start_ui(client)
}
