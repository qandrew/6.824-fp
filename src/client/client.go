package main

import (
	"bufio"
	"log"
	"net/rpc"
	"op"
	"os"
)

type OTClient struct {
	rpc_client *rpc.Client
}

func NewOTClient() *OTClient {
	cl := &OTClient{}
	cl.rpc_client, err = rpc.Dial("tcp", "localhost:42586")
	if err != nil {
		log.Fatal(err)
	}
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
	err = rpc_client.Call("Listener.ApplyOp", op, &reply)
	if err != nil {
		log.Fatal(err)
	}
	return reply
}

func main() {
	client := NewOTClient()
	start_ui(client)
}
