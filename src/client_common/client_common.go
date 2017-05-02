package client_common

import (
	"crypto/rand"
	"log"
	"math/big"
	"net/rpc"
	"op"
	"time"
	"fmt"
	"errors"
	// "strconv"
)

const SLEEP = 1000

type OTClient struct {
	rpc_client *rpc.Client
	uid        int64
	logs       []op.Op // logs of all operations
	currState  string // current state of text
	version    int // client side sent

	versionS   int // client side received
	Debug 	   bool
}

func NewOTClient() *OTClient {
	cl := &OTClient{}
	rpc_client, err := rpc.Dial("tcp", "localhost:42586")
	cl.rpc_client = rpc_client
	if err != nil {
		log.Fatal(err)
	}

	cl.uid = nrand()
	cl.versionS = 1 // let the starting state be (1,1)
	cl.version = 1 // let the starting state be (1,1)
	cl.Debug = false
	var ack bool
	cl.rpc_client.Call("OTServer.Init", cl.uid, &ack)

	go func(){
		var empty op.Op
		// sleep := 1000
		empty.Uid = cl.uid
		empty.OpType = "empty"
		for {
			time.Sleep(SLEEP*time.Millisecond)
			// time.Sleep(sleep*time.Millisecond) // some time/duration bug
			empty.Version = cl.version // update version
			empty.VersionS = cl.versionS // update version
			// fmt.Println("sending", empty)
			var reply op.Op
			err := cl.rpc_client.Call("OTServer.ApplyOp", empty, &reply)
			if err != nil {
				log.Fatal(err)
			} 
			if reply.OpType != "empty" {
				if cl.Debug{
					fmt.Println("client behind; received", reply)
				}
				cl.xform(reply) // do some OT
				// sleep = 0 // instantly request more 
				// but for now
				// cl.version = reply.Version
				// cl.versionS = reply.VersionS
			} else {
				// sleep = 1000 // go back to periodical
			}
		}
	}()

	return cl
}

func (cl *OTClient) xform(args op.Op) {
	if args.OpType == "ins" || args.OpType == "del" {
		// cl.version = args.Version // CHANGE THIS LATER
		// cl.logs = append(cl.logs, *args) // CHANGE THIS LATER

		if args.OpType == "ins" {
			// cl.currState += args.Payload
			if args.Position == 0 {
				cl.currState = args.Payload + cl.currState // append at beginning
			} else {
				cl.currState = cl.currState[:args.Position] + args.Payload + cl.currState[args.Position:]
			}
		} else {
			if args.Position == len(cl.currState) && len(cl.currState) != 0 {
				cl.currState = cl.currState[:args.Position-1]
			} else {
				cl.currState = cl.currState[:args.Position-1] + cl.currState[args.Position:]
			}
		}
		cl.version =  args.VersionS + 1
		cl.versionS = args.VersionS + 1 // SINCE WE APPLIED FUNCTION, we can update server version
	} else {
		log.Fatal(errors.New("xform: wrong operation input"))
	}
	if cl.Debug{
		fmt.Println("xform: now", cl.currState, "ver", cl.version, "serv", cl.versionS)
	}
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
	// when applying a new log, update the version
	v := cl.version
	cl.version += 1
	return v
}

func (cl *OTClient) Insert(ch rune, pos int) {
	args := op.Op{"ins", pos, cl.addVersion(),cl.versionS,cl.uid, string(ch)}
	cl.SendOp(&args)
}

func (cl *OTClient) Delete(pos int) {
	if pos != 0 { // can't delete first
		args := op.Op{"del", pos, cl.addVersion(),cl.versionS,cl.uid, ""} 
		cl.SendOp(&args)
	}
}

func (cl *OTClient) RandOp() {
	// let the client do a random operation
	// pos := rand.Int(rand.Reader,len(currState))
	// val := strconv.Itoa(rand.Int(rand.Reader,9))
	// args := op.Op{"ins", pos, cl.addVersion(),cl.versionS,cl.uid, val} 
	// cl.SendOp(&args)

}

func (cl *OTClient) SendOp(args *op.Op) op.Op {
	var reply op.Op
	cl.logs = append(cl.logs, *args) // add to logs
	err := cl.rpc_client.Call("OTServer.ApplyOp", args, &reply)
	if err != nil {
		log.Fatal(err)
	}
	if reply.OpType == "empty"{
		// don't do anything
	} else if reply.OpType == "good" {
		cl.versionS = reply.VersionS // update server version
	} else if reply.OpType == "ins" || reply.OpType == "del"{
		cl.xform(reply) // do some OT
	}
	return reply
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 40) // was 62
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}
