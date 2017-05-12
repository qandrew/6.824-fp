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

const SLEEP = 1000 // longer for better OT debug, shorter for accurate performance

type OTClient struct {
	rpc_client *rpc.Client
	uid        int64
	logs       []op.Op // logs of all operations
	currState  string  // current state of text
	version    int     // version=x means that the client has processed all of the
	// server's messages up to x. If a message is send, it will be with
	// version x+1.

	// versionS   int // client side received
	Debug   bool
	logFile *bufio.Writer

	outgoingQueue []op.Op // A queue of operations that have been locally
	// applied but messages have not been sent
	waitingForPop bool // Used to wait for messages in the buffer to pop
	startBufferIndex int

	insertCb func(int, rune)
	deleteCb func(int)

	chanPull chan bool
	chanSend chan bool
	chanPop chan bool
}

func NewOTClient() *OTClient {
	cl := &OTClient{}

	reader := bufio.NewReader(os.Stdin)
	cl.Println("Enter IP addr (empty for localhost): ")
	text, _ := reader.ReadString('\n')
	if len(text) == 1 {
		text = "localhost:42586"
		cl.Println("addr\t", text)
	} else {
		text = text[:len(text)-1] + ":42586"
		cl.Println("addr\t", text)
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
	cl.waitingForPop = false
	cl.outgoingQueue = make([]op.Op, 0)

	cl.insertCb = func(x int, ch rune) { /* noop */ }
	cl.deleteCb = func(x int) { /* noop */ }

	cl.chanPull = make(chan bool, 3)   // not sure how wide
	cl.chanSend = make(chan bool, 10) // not sure how wide
	cl.chanPop = make(chan bool, 1)

	var ack bool
	cl.rpc_client.Call("OTServer.Init", cl.uid, &ack)

	go cl.Pull()     // receiving operations
	go cl.SendShit() // sending operations

	return cl
}

func (cl *OTClient) QueuePush(op op.Op) {
	cl.outgoingQueue = append(cl.outgoingQueue, op)
}

func (cl *OTClient) QueuePop() op.Op {
	firstOp := cl.outgoingQueue[0]
	cl.outgoingQueue = cl.outgoingQueue[1:]
	return firstOp
}

func (cl *OTClient) QueueFirstItem() op.Op {
	return cl.outgoingQueue[0]
}

func (cl *OTClient) getNextLogVer() int {
	if len(cl.logs) == 0{
		return cl.version 
	} else {
		return cl.logs[len(cl.logs)-1].Version+1
	}
}

func (cl *OTClient) Pull() {
	// endless goroutine
	var empty op.Op
	empty.Uid = cl.uid
	empty.OpType = "empty"
	for {
		select { // choose how long to wait
			case <-cl.chanPull:
			case <-time.After(SLEEP * time.Millisecond):
		}

		empty.Version = cl.version // update version
		var reply op.OpReply
		reply.Logs = make([]op.Op, 1)
		// if cl.Debug {cl.Println("client to call", empty, reply)}
		err := cl.rpc_client.Call("OTServer.Broadcast", empty, &reply)
		if err != nil {
			log.Fatal(err)
		}
		if reply.Logs[0].OpType != "empty" {
			if cl.Debug { cl.Println("client behind; received", reply) }
			//respond to all of the logs
			cl.startBufferIndex = 0 // reset startBufferIndex
			for i := 0; i < len(reply.Logs); i++ {
				cl.receiveSingleLog(reply.Logs[i])
			}
			cl.startBufferIndex = 0
			if cl.waitingForPop{
				cl.chanPop <- true // request to pop buffer if there isn't anything in buffer
			}
			if cl.Debug { cl.Println("done receive, ver now", cl.version) }
		}
	}
}

func (cl *OTClient) Println(args ...interface{}) {
	if cl.logFile != nil {
		fmt.Fprintln(cl.logFile, args...)
		cl.logFile.Flush()
	} else {
		fmt.Println(args...)
	}
}

func (cl *OTClient) SetLogFile(filename string) {
	f, _ := os.Create(filename)
	cl.logFile = bufio.NewWriter(f)
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
	if cl.logs[i-1].Version == i { // we 1 index
		return cl.logs[i-1]
		// } else cl.logs[i-1].Version == i {

	} else {
		for _, log := range cl.logs {
			if log.Version == i {
				return log
			}
		}
		return cl.logs[i-1]
	}
}

func (cl *OTClient) addCurrState(args op.Op) {
	// no OT needed
	if args.OpType == "ins" {
		// cl.currState += args.Payload
		if args.Position == 0 {
			cl.currState = args.Payload + cl.currState // append at beginning
		} else if args.Position == len(cl.currState) {
			cl.currState = cl.currState + args.Payload
		} else {
			cl.currState = cl.currState[:args.Position] + args.Payload + cl.currState[args.Position:]
		}
	} else if args.OpType == "del" { 
		if args.Position == len(cl.currState) && len(cl.currState) != 0 {
			cl.currState = cl.currState[:args.Position-1]
		} else if args.Position == 0 {
			cl.currState = cl.currState[1:]
		} else {
			cl.currState = cl.currState[:args.Position-1] + cl.currState[args.Position:]
		}
	}
	// args.Version = cl.version       // safety check?
	cl.logs = append(cl.logs, args) // add to logs
	// cl.version++                    // increment version only when we have appended it to logs
	if cl.Debug {
		cl.Println("addCurrState:\n=====\n"+cl.currState, "\n=====\nver", cl.version, "logs", cl.logs)
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
	if cl.Debug {
		cl.Println("shapshot got", snap, "now ver", cl.version, "currState", cl.currState)
	}
	return snap.Value
}

func (cl *OTClient) Insert(ch rune, pos int) {
	args := op.Op{OpType: "ins", Position: pos, Version: cl.getNextLogVer(), VersionS: cl.version,
		Uid: cl.uid, Payload: string(ch)}
	cl.addCurrState(args)
	cl.QueuePush(args)
	cl.chanSend <- true
}

func (cl *OTClient) Delete(pos int) {
	if pos != 0 { // can't delete first
		args := op.Op{OpType: "del", Position: pos, Version: cl.getNextLogVer(), VersionS: cl.version,
			Uid: cl.uid, Payload: ""}
		cl.addCurrState(args)
		cl.QueuePush(args)
		cl.chanSend <- true
	}
}

func (cl *OTClient) RandOp() {
	// deprecated
	// Needs to be revamped or something now that SendOp is gone
	// let the client do a random operation
	var pos int
	if len(cl.currState) == 0 {
		pos = 0
	} else {
		pos = r.Intn(len(cl.currState))
	}
	val := strconv.Itoa(r.Intn(9))
	args := op.Op{OpType: "ins", Position: pos, Version: cl.version, Uid: cl.uid, Payload: val}
	cl.Println("calling", args)
	// cl.SendShit(&args)
}

func (cl *OTClient) SendShit() {
	// infinite loop to send stuff
	for {
		select{
			case <-cl.chanSend: // receive operation to send
		}
		args := cl.QueueFirstItem()
		args.Version = cl.version // update version number before we send
		// args := cl.QueuePop()
		if cl.Debug {
			cl.Println("beginning send", args)
		}
		/*
		if args != cl.logs[len(cl.logs)-1] {
			// do pre-RPC call OT
			for i := args.Version; i < cl.logs[len(cl.logs)-1].Version; i++ {
				args, _ = op.Xform(args, cl.getLogVersion(i))
			}
			args.Version = cl.version // update version
		}
		*/

		var upToDate bool
		err := cl.rpc_client.Call("OTServer.ApplyOp", &args, &upToDate)
		if err != nil {
			log.Fatal(err)
		}
		if cl.Debug {
			cl.Println("after send", args, upToDate)
		}
		if !upToDate {
			cl.chanPull <- true // send pull request
			cl.waitingForPop = true
			select {
				case <- cl.chanPop:
					cl.QueuePop()
					cl.waitingForPop = false
			}
		} else {
			cl.version++ // increment version since we are up to date with the server
			cl.QueuePop() // pop the queue
		}

	}
}

func (cl *OTClient) receiveSingleLog(args op.Op) {
	// once one log is received, xform everything in the buffer
	// furthermore, xform the current state with the transformed log
	// if cl.Debug {//cl.Println("received", args)}
	if args.Uid == cl.uid {
		// don't insert something that the client already has
		cl.startBufferIndex = 1
		cl.version++
		return
	}

	if args.OpType == "empty" || args.OpType == "good" {
		// don't do anything
	} else if args.OpType == "ins" || args.OpType == "del" || args.OpType == "noOp" {
		/*
		if args.VersionS == cl.version && len(cl.outgoingQueue) == 0 && args.VersionS == cl.logs[len(cl.logs)-1].Version+1 {
			cl.addCurrState(args) // no need to do OT, simply update and add logs
			cl.version++
			if args.OpType == "ins" {
				r := []rune(args.Payload) // convert string to rune
				cl.Println("Simply insert ", args.Payload, " at position ", args.Position)
				cl.insertCb(args.Position, r[0]) // when this works
			} else {
				cl.Println("Simply delete at position ", args.Position)
				cl.deleteCb(args.Position)
			}
		} else { */
		temp := args
		// t1 := args // temporary setting
		/*
		for i := args.Version; i < cl.logs[len(cl.logs)-1].Version; i++ {
			// do operational transforms
			t1 = cl.getLogVersion(i)
			if temp.Uid != t1.Uid{ // only do OT if on a different version
				temp, _ = op.Xform(temp, t1) // get log version could be optimized
			}
		}
		*/

		// We need to xform everything in the buffer
		for i := cl.startBufferIndex; i < len(cl.outgoingQueue); i++ {
			// I think this is right but m a y b e n o t
			if temp.Uid != cl.outgoingQueue[i].Uid {
				// This check might not be necessary
				cl.outgoingQueue[i], temp = op.Xform(cl.outgoingQueue[i], temp)
			}
		}

		if cl.Debug {
			cl.Println("receive xform to add", temp, "cl ver", cl.version, "buffer", cl.outgoingQueue)
		}
		if len(cl.logs) > 0 {
			temp.Version = cl.logs[len(cl.logs)-1].Version+1 // overwrite the version
		}
		cl.addCurrState(temp)
		cl.version++
		if args.OpType == "ins" {
			r := []rune(temp.Payload)
			cl.Println("insert ", temp.Payload, " at pos ", temp.Position, " after transform")
			cl.insertCb(temp.Position, r[0]) // not sure if it works
		} else {
			cl.Println("delete at position ", args.Position)
			cl.deleteCb(temp.Position)
		}
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
		if cl.Debug {
			//cl.Println("receive xform: now", cl.currState, "ver", cl.version, "logs", cl.logs)
		}
	}
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 40) // was 62
	bigx, _ := randC.Int(randC.Reader, max)
	x := bigx.Int64()
	return x
}
