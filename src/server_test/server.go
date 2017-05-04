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
	resp.VersionS = sv.version // respond with version also
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
	fmt.Println("received", args, sv.version, len(resp.Logs), resp)
	if sv.version > args.Version { // server ahead of client
		fmt.Println("broadcast to", args.Uid, resp)
		// args.Version guaranteed to be >= 1
		resp.Logs[0] = sv.logs[args.Version-1] // return the next log

		sv.clients[args.Uid] += 1 // assume that client will be able to resolve
		// tho technically we might need an ack from client
	} else {
		fmt.Println("broadcast up to date", args.Uid, sv.version, sv.clients[args.Uid] )
		resp.Logs[0].OpType = "empty" // client is at same state as server
	}
	fmt.Println("processed", resp)

	return nil
}

func (sv *OTServer) ApplyTransformation(args *op.Op, resp *op.OpReply) error {
	fmt.Println("incoming op ", args)
	if args.OpType != "ins" && args.OpType != "del" {
		log.Fatal(errors.New("xform: wrong operation input"))
	}

	resp.Logs = make([]op.Op,1) // make a new entry in resp logs

	if args.VersionS == sv.version && args.Version == sv.clients[args.Uid] {
		// in this case, we don't need to do any transforms
		if args.OpType == "ins" {
			// sv.currState += args.Payload
			if args.Position == 0 {
				sv.currState = args.Payload + sv.currState // append at beginning
			} else {
				sv.currState = sv.currState[:args.Position] + args.Payload + sv.currState[args.Position:]
			}
		} else {
			if args.Position == len(sv.currState) && len(sv.currState) != 0 {
				sv.currState = sv.currState[:args.Position-1]
			} else {
				sv.currState = sv.currState[:args.Position-1] + sv.currState[args.Position:]
			}
		}
		sv.clients[args.Uid] =  args.Version + 1
		sv.version = args.VersionS + 1 // SINCE WE APPLIED FUNCTION, we can update server version
		sv.logs = append(sv.logs, *args)

		resp.Logs[0].OpType = "good"
		resp.Logs[0].VersionS = sv.version
	} else if sv.clients[args.Uid] < args.Version && sv.version > args.VersionS {
		// diverging situation
		// ex if cl at (1,0) and args at (0,1)
		// we want to apply args' such that cl will end up at (1,1)
		logTemp := sv.getLogVersion(args.Version)
		if logTemp.Position < args.Position{
			// modify where we actually want to insert
			// since a previous insert will mess up position
			if logTemp.OpType == "ins" { args.Position += 1
			} else if logTemp.OpType == "del" {args.Position -= 1}
		}
		if args.OpType == "ins" {
			// sv.currState += args.Payload
			if args.Position == 0 {
				sv.currState = args.Payload + sv.currState // append at beginning
			} else {
				sv.currState = sv.currState[:args.Position] + args.Payload + sv.currState[args.Position:]
			}
		} else {
			if args.Position == len(sv.currState) && len(sv.currState) != 0 {
				sv.currState = sv.currState[:args.Position-1]
			} else {
				sv.currState = sv.currState[:args.Position-1] + sv.currState[args.Position:]
			}
		}
		sv.clients[args.Uid] = args.Version + 1 // update server version kept on args
		sv.logs = append(sv.logs, *args) // append the modified logs

	}
	fmt.Println("ApplyTransformation now: ", strings.Replace(sv.currState, "\n", "\\n", -1))
	fmt.Println("Clients version", sv.clients, sv.version)
	fmt.Println("logs", sv.logs)
	fmt.Println("replying", resp)


	return nil
}

func main() {
	addy, err := net.ResolveTCPAddr("tcp", "localhost:42586")
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
