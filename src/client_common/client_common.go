package client_common

import (
	randC "crypto/rand"
	"fmt"
	"log"
	"math/big"
	r "math/rand"
	"net/rpc"
	"op"
	"time"
	//	"errors"
	"strconv"

	// for reading in
	"bufio"
	"os"
)

const SLEEP = 1000

type OTClient struct {
	rpc_client *rpc.Client
	uid        int64
	logs       []op.Op // logs of all operations
	currState  string // current state of text
	version    int // version=x means that the client has processed all of the
		       // server's messages up to x. If a message is send, it will be with
		       // version x+1.

	// versionS   int // client side received
	Debug 	   bool

	outgoingQueue     []op.Op // A queue of operations that have been locally
				  // applied but messages have not been sent

	insertCb func(int, rune)
	deleteCb func(int)
}

func NewOTClient() *OTClient {
	cl := &OTClient{}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter IP addr (empty for localhost): ")
	text, _ := reader.ReadString('\n')
	if len(text) == 1 {
		text = "localhost:42586"
		fmt.Println("addr\t", text)
	} else {
		text = text[:len(text)-1] + ":42586"
		fmt.Println("addr]t", text)
	}
	time.Sleep(SLEEP * time.Millisecond)
	rpc_client, err := rpc.Dial("tcp", text)
	cl.rpc_client = rpc_client
	if err != nil {
		log.Fatal(err)
	}

	cl.uid = nrand()
	// cl.versionS = 1 // let the starting state be (1,1)
	cl.version = 1 // let the starting state be (1,1)
	cl.Debug = false
	cl.outgoingQueue = make([]op.Op, 0)

	cl.insertCb = func(x int, ch rune) { /* noop */ }
	cl.deleteCb = func(x int) { /* noop */ }

	var ack bool
	cl.rpc_client.Call("OTServer.Init", cl.uid, &ack)

	go func() {
		var empty op.Op
		sleep := 1000
		empty.Uid = cl.uid
		empty.OpType = "empty"
		for {
			time.Sleep(time.Duration(sleep)*time.Millisecond) // some time/duration bug
			empty.Version = cl.version // update version
			// empty.VersionS = cl.versionS // update version
			var reply op.OpReply
			reply.Logs = make([]op.Op, 1)
			reply.Logs[0].Payload = "u"
			reply.Num = 3
			if cl.Debug {
				fmt.Println("client to call", empty, reply)
			}
			err := cl.rpc_client.Call("OTServer.Broadcast", empty, &reply)
			if err != nil {
				log.Fatal(err)
			}
			if reply.Logs[0].OpType != "empty" {
				sleep = 10 // instantly request more
				if cl.Debug {
					fmt.Println("client behind; received", reply)
				}
				// TODO: change this to respond to all of the logs
				cl.receiveSingleLog(reply.Logs[0]) // do some OT
			} else {
				sleep = 1000 // go back to periodical
			}
		}
	}()

	return cl
}

func (cl *OTClient) RegisterInsertCb(f func(int, rune)) {
	cl.insertCb = f
}

func (cl *OTClient) RegisterDeleteCb(f func(int)) {
	cl.deleteCb = f
}

func (cl *OTClient) getLogVersion(i int) op.Op {
	// return the log with version index i
	// in case we condense log in future
	if cl.logs[i-1].Version == i {
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
	if cl.Debug {
		fmt.Println("addCurrState: now", cl.currState, "ver", cl.version)
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
		// cl.versionS = snap.VersionS
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
	args := op.Op{OpType: "ins", Position: pos, Version: cl.addVersion(), VersionS: 0,
			Uid: cl.uid, Payload: string(ch)}
	cl.SendOp(&args)
}

func (cl *OTClient) Delete(pos int) {
	if pos != 0 { // can't delete first
		args := op.Op{OpType: "del", Position: pos, Version: cl.addVersion(), VersionS: 0, Payload: ""}
		cl.SendOp(&args)
	}
}

func (cl *OTClient) RandOp() {
	// let the client do a random operation
	var pos int
	if len(cl.currState) == 0 {
		pos = 0
	} else {
		pos = r.Intn(len(cl.currState))
	}
	val := strconv.Itoa(r.Intn(9))
	args := op.Op{OpType: "ins", Position: pos, Version: cl.addVersion(), Uid: cl.uid, Payload: val}
	fmt.Println("calling", args)
	cl.SendOp(&args)

}

func (cl *OTClient) SendOp(args *op.Op) op.Op {
	// TODO: make SendOp be able to edit the version number
	// TODO: when the buffer is implemented, SendOp should probably just add the op to the buffer.
	var reply op.OpReply

	cl.outgoingQueue = append(cl.outgoingQueue, *args)
	reply.Logs = make([]op.Op, 1) // make at least one, let server append
	cl.logs = append(cl.logs, *args) // add to logs

	err := cl.rpc_client.Call("OTServer.ApplyOp", args, &reply)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: What we probably want to do is call receiveSingleLog on each of the elements in the
	// log individually, and only then do we remove an op from the buffer
	for i := 0; i < len(reply.Logs); i++ {
		cl.receiveSingleLog(reply.Logs[i])
	}
	// cl.receiveSingleLog(reply.Logs[0])
	// pop from the queue
	cl.outgoingQueue = cl.outgoingQueue[1:]
	return reply.Logs[0]
}

func (cl *OTClient) receiveSingleLog(args op.Op) {
	// once one log is received, xform everything in the buffer
	// furthermore, xform the current state with the transformed log
	if args.OpType == "empty" || args.OpType == "noOp" {
		// don't do anything
	} else if args.OpType == "ins" || args.OpType == "del" {
		// We need to xform everything in the buffer
		temp := args
		for i := 0; i < len(cl.outgoingQueue); i++ {
			// I think this is right but m a y b e n o t
			cl.outgoingQueue[i], temp = op.Xform(cl.outgoingQueue[i], temp)
		}

		cl.addCurrState(temp)
		cl.version++
		/*
		Here's a shitty schematic of the above operations

		 buf[0] /\
		       /  \
	      buf[1]  /\  / new buf[0]
		     /  \/
	    buf[2]  /\  / new buf[1]
		   /  \/
		   \  / new buf[2] and so on
      temp (newest) \/

		We need to apply the last value of temp locally
		*/
		if cl.Debug{
			fmt.Println("xform: now", cl.currState, "ver", cl.version, "logs", cl.logs)
		}
	}
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 40) // was 62
	bigx, _ := randC.Int(randC.Reader, max)
	x := bigx.Int64()
	return x
}
