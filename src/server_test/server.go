// https://gist.github.com/jordanorelli/2629049

package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"op"
	"strings"
	"sync"

	// for reading in
	"bufio"
	"os"
)

type OTServer struct {
	mu        sync.Mutex // Lock to protect shared access to this peer's state
	logs      []op.Op
	currState string        //string of the most updated version
	version   int           //last updated version

	clients   map[int64]int //*rpc.Client keep track of version num
}

func (sv *OTServer) getLogVersion (i int) op.Op{
	// return the log with version index i
	if sv.logs[i-1].Version == i{
		return sv.logs[i-1]
	}
	fmt.Println("fuck")
	return sv.logs[i-1]
}

func (sv *OTServer) getLastLogVersion() int {
	return sv.logs[len(sv.logs)-1].Version
}

func (sv *OTServer) Init(clientID int64, resp *bool) error {
	if _, ok :=sv.clients[clientID]; !ok {
		sv.clients[clientID] = 1
		fmt.Println("registered client", clientID)
		*resp = true
	}
	*resp = false
	return nil
}

func (sv *OTServer) ApplyOp(args *op.Op, resp *op.OpReply) error {
	// operation called by the client
	// fmt.Println(args)
	var err error
	if args.OpType == "empty" {
		err = sv.Broadcast(args,resp)
	} else {
		err = sv.ApplyTransformation(args, resp) // do the actual OT here
	}

	return err
}

func (sv *OTServer) GetSnapshot(req *op.Snapshot, resp *op.Snapshot) error {
	// resp.VersionS = sv.version // respond with version also
	resp.Value = sv.currState
	sv.clients[req.Uid] = sv.version // assume that client can force update
	fmt.Println("Snapshot to", req.Uid, sv.clients)
	return nil
}

func (sv *OTServer) Broadcast(args *op.Op, resp *op.OpReply) error {
	// if sv.clients[args.Uid] > args.Version {
	// 	return errors.New("Broadcast: client missing operation, out of order")
	// }
	resp.Logs = make([]op.Op,1)
	// fmt.Println("received", args, sv.version, len(resp.Logs), resp)
	if sv.version > args.Version { // server ahead of client
		fmt.Println("broadcast to", args.Uid, resp)
		// args.Version guaranteed to be >= 1
		resp.Logs[0] = sv.logs[args.Version-1] // return the next log

		sv.clients[args.Uid] += 1 // assume that client will be able to resolve
		// tho technically we might need an ack from client
	} else {
		// fmt.Println("broadcast up to date", args.Uid, sv.version, sv.clients[args.Uid] )
		resp.Logs[0].OpType = "empty" // client is at same state as server
	}
	// fmt.Println("processed", resp)

	return nil
}

func (sv *OTServer) useOperationAndUpdate(args op.Op){
	// apply to curr state, increment version, and append to log
	sv.currState = op.ApplyOperation(args,sv.currState)
	args.VersionS = sv.version // safety check
	sv.version++ // SINCE WE APPLIED FUNCTION, we can update server version
	sv.logs = append(sv.logs, args)
}

func (sv *OTServer) ApplyTransformation(args *op.Op, resp *op.OpReply) error {
	fmt.Println("incoming op ", args)
	if args.OpType != "ins" && args.OpType != "del" {
		log.Fatal(errors.New("xform: wrong operation input"))
	}
	resp.Logs = make([]op.Op,1) // make a new entry in resp logs

	if args.Version == sv.version {
		// in this case, we don't need to do any transforms
		// Make this shit into some function
		sv.useOperationAndUpdate(*args)
		sv.clients[args.Uid] =  args.Version + 1

		resp.Logs[0].OpType = "good"
		// resp.Logs[0].VersionS = sv.version
	} else if args.Version < sv.version {
		// diverging situation
		// ex if cl at (1,0) and args at (0,1)
		// we want to apply args' such that cl will end up at (1,1)
		tempArg := *args
		transformIndex := args.Version

		for ; transformIndex <= sv.getLastLogVersion(); transformIndex++ {
			tempArg, _ = op.Xform(tempArg, sv.getLogVersion(transformIndex))
		}

		sv.useOperationAndUpdate(tempArg)
/*
	if args.VersionS == sv.version && args.Version == sv.clients[args.Uid] {
		// in this case, we don't need to do any transforms
		sv.useOperationAndUpdate(*args)
		sv.clients[args.Uid] =  args.Version + 1
		resp.Logs[0].OpType = "good"
		resp.Logs[0].VersionS = sv.version
	} else {
		// process the correct transformation
		// append to log, apply operation etc

		if args.VersionS == sv.version -1 {
			// if client just one behind, the transformation is simple

			a1, _ := op.Xform(*args, sv.getLogVersion(args.VersionS)) // do the transformation
			sv.useOperationAndUpdate(a1)
			sv.clients[args.Uid] =  sv.version // assuming operation sent back
			resp.Logs[0] = sv.getLogVersion(args.VersionS)
			// resp.Logs[0].VersionS = args.VersionS // since operation already done on server side, using temporary code, send back to client correct operation

		} else {
			fmt.Println("ApplyTransformation haven't dealt with this situation")
		}
		*/
		sv.clients[args.Uid] = args.Version + 1 // update server version kept on args
		sv.logs = append(sv.logs, tempArg) // append the modified logs
	} else {
		fmt.Println("ERROR: client version higher than the server version")
	}

	fmt.Println("ApplyTransformation now: ", strings.Replace(sv.currState, "\n", "\\n", -1))
	// fmt.Println("Clients version", sv.clients, sv.version)
	fmt.Println("logs", sv.logs)
	fmt.Println("replying", resp)

	return nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter IP addr (empty for localhost): ")
	text, _ := reader.ReadString('\n')
	if len(text) == 1 {
		text = "localhost:42586"
		fmt.Println("addr\t", text)
	} else {
		text = text[:len(text)-1] + ":42586"
		fmt.Println("addr\t", text)
	}
	addy, err := net.ResolveTCPAddr("tcp", text)
	if err != nil {
		log.Fatal(err)
	}

	inbound, err := net.ListenTCP("tcp", addy)
	if err != nil {
		log.Fatal(err)
	}

	sv := new(OTServer)
	sv.clients = make(map[int64]int) //*rpc.Client)
	sv.version = 1

	rpc.Register(sv)
	rpc.Accept(inbound)
}
