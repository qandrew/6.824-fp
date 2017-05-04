package client_common

import (
	randC "crypto/rand"
	r "math/rand"
	"log"
	"math/big"
	"net/rpc"
	"op"
	"time"
	"fmt"
	"errors"
	"strconv"
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

	outgoing   chan op.Op
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
	cl.outgoing = make(chan op.Op, 100)
	var ack bool
	cl.rpc_client.Call("OTServer.Init", cl.uid, &ack)

	go func(){
		var empty op.Op
		sleep := 1000
		empty.Uid = cl.uid
		empty.OpType = "empty"
		for {
			time.Sleep(time.Duration(sleep)*time.Millisecond) // some time/duration bug
			empty.Version = cl.version // update version
			empty.VersionS = cl.versionS // update version
			var reply op.Op
			err := cl.rpc_client.Call("OTServer.ApplyOp", empty, &reply)
			if err != nil {
				log.Fatal(err)
			}
			if reply.OpType != "empty" {
				sleep = 10 // instantly request more
				if cl.Debug{
					fmt.Println("client behind; received", reply)
				}
				cl.receive(reply) // do some OT
			} else {
				sleep = 1000 // go back to periodical
			}
		}
	}()

	return cl
}

func (cl *OTClient) getLogVersion (i int) op.Op{
	// return the log with version index i
	// in case we condense log in future
	if cl.logs[i-1].Version == i{
		return cl.logs[i-1]
	}
	fmt.Println("fuck")
	return cl.logs[i-1]
}

func (cl *OTClient) addCurrState(args op.Op) {
	// no OT needed
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
	if cl.Debug{
		fmt.Println("addCurrState: now", cl.currState, "ver", cl.version, "serv", cl.versionS)
	}
}

func (cl *OTClient) GetSnapshot() string {
	snapIn := op.Snapshot{}
	snapIn.Uid = cl.uid
	snap := op.Snapshot{}
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
	var pos int
	if len(cl.currState) == 0{ pos = 0
	} else { pos = r.Intn(len(cl.currState))}
	val := strconv.Itoa(r.Intn(9))
	args := op.Op{"ins", pos, cl.addVersion(),cl.versionS,cl.uid, val}
	fmt.Println("calling", args)
	cl.SendOp(&args)

}

func (cl *OTClient) SendOp(args *op.Op) op.Op {
	var reply op.Op
	cl.logs = append(cl.logs, *args) // add to logs
	cl.addCurrState(*args) // do some OT
	err := cl.rpc_client.Call("OTServer.ApplyOp", args, &reply)
	if err != nil {
		log.Fatal(err)
	}
	cl.receive(reply)
	return reply
}

func (cl *OTClient) receive(args op.Op) {
	if args.OpType == "empty"{
		// don't do anything
/*	} else if reply.OpType == "good" {
		cl.versionS = reply.VersionS // update server version
*/
	} else if args.OpType == "ins" || args.OpType == "del"{
/*		if args.OpType != "ins" && args.OpType != "del" {
			log.Fatal(errors.New("xform: wrong operation input"))
		}
*/
		if args.VersionS == cl.versionS && args.Version == cl.version {
			// in this case, we don't need to do any transforms
			cl.addCurrState(args)
			cl.versionS = args.VersionS + 1 // SINCE WE APPLIED FUNCTION, we can update server version
			cl.logs = append(cl.logs, args)
			if cl.Debug{
				fmt.Println("xform normal: now", cl.currState, "ver", cl.version, "serv", cl.versionS)
			}
		} else if cl.version > args.Version && cl.versionS < args.VersionS {
			// diverging situation
			// ex if cl at (1,0) and args at (0,1)
			// we want to apply args' such that cl will end up at (1,1)
			logTemp := cl.getLogVersion(args.Version)
			if logTemp.Position < args.Position{
				// modify where we actually want to insert
				// since a previous insert will mess up position
				if logTemp.OpType == "ins" { args.Position += 1
				} else if logTemp.OpType == "del" {args.Position -= 1}
			}
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
			cl.versionS = args.VersionS // update server version kept on args
			cl.logs = append(cl.logs, args) // append the modified logs
			if cl.Debug{
				fmt.Println("xform diverge: now", cl.currState, "ver", cl.version, "serv", cl.versionS)
			}
		}
		if cl.Debug{
			fmt.Println("xform: now", cl.currState, "ver", cl.version, "serv", cl.versionS, "logs", cl.logs)
		}
	}
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 40) // was 62
	bigx, _ := randC.Int(randC.Reader, max)
	x := bigx.Int64()
	return x
}