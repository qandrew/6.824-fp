package client_common

import (
	"crypto/rand"
	"log"
	"math/big"
	"net/rpc"
	"op"
)

type OTClient struct {
	rpc_client *rpc.Client
	uid        int64
	version    int // client side sent
	versionS   int // client side received
}

func NewOTClient() *OTClient {
	cl := &OTClient{}
	rpc_client, err := rpc.Dial("tcp", "localhost:42586")
	cl.rpc_client = rpc_client
	if err != nil {
		log.Fatal(err)
	}

	cl.uid = nrand()
	// cl.
	var ack bool
	cl.rpc_client.Call("OTServer.Init", cl.uid, &ack)

	return cl
}

func (cl *OTClient) GetSnapshot() string {
	snap := op.Snapshot{}
	snapIn := op.Snapshot{}
	cl.rpc_client.Call("OTServer.GetSnapshot", &snapIn, &snap)
	return snap.Value
}

func (cl *OTClient) Insert(ch rune, pos int) {
	args := op.Op{"ins", pos, 0, 0, string(ch)} // version?
	reply := cl.SendOp(&args)
	cl.version = reply.Version // temporary
}

func (cl *OTClient) Delete(pos int) {
	if pos != 0 { // can't delete first
		args := op.Op{"del", pos, 0, 0, ""} // version?
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

func nrand() int64 {
	max := big.NewInt(int64(1) << 40) // was 62
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}
