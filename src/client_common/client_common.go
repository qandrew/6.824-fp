package client_common

import (
	"crypto/rand"
	"log"
	"math/big"
	"net/rpc"
	"op"
	"time"
	"fmt"
)

const SLEEP = 1000

type OTClient struct {
	rpc_client *rpc.Client
	uid        int64
	currState  string // current state of text
	version    int // client side sent
	versionS   int // client side received
	// DEBUG 	
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

	go func(){
		var empty op.Op
		empty.Uid = cl.uid
		empty.OpType = "empty"
		for {
			time.Sleep(SLEEP*time.Millisecond)
			empty.Version = cl.version // update version
			empty.VersionS = cl.versionS // update version
			// fmt.Println("sending", empty)
			var reply op.Op
			err := cl.rpc_client.Call("OTServer.ApplyOp", empty, &reply)
			if err != nil {
				log.Fatal(err)
			} 
			if reply.OpType != "empty" {
				fmt.Println("client behind; received", reply)
				// do some OT
				// but for now
				cl.version = reply.Version
				cl.versionS = reply.VersionS
			}
		}
	}()

	return cl
}

func (cl *OTClient) GetSnapshot() string {
	snap := op.Snapshot{}
	snapIn := op.Snapshot{}
	cl.rpc_client.Call("OTServer.GetSnapshot", &snapIn, &snap)
	if cl.version < snap.VersionS { 
		// update client's version of itself and client's server version record
		cl.version = snap.VersionS
		cl.versionS = snap.VersionS
		cl.currState = snap.Value // this might not be safe?
	}
	return snap.Value
}

func (cl *OTClient) addVersion() int {
	cl.version += 1
	return cl.version
}

func (cl *OTClient) Insert(ch rune, pos int) {
	args := op.Op{"ins", pos, cl.addVersion(),cl.versionS,cl.uid, string(ch)} // version?
	reply := cl.SendOp(&args)
	if reply.Version > cl.version{
		// do OT
	}
	// cl.version = reply.Version // temporary
}

func (cl *OTClient) Delete(pos int) {
	if pos != 0 { // can't delete first
		args := op.Op{"del", pos, cl.addVersion(),cl.versionS,cl.uid, ""} // version?
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
