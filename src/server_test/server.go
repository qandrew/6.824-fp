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

func (sv *OTServer) Init(clientID int64, resp *bool) error {
	if _, ok :=sv.clients[clientID]; !ok {
		sv.clients[clientID] = 0
		fmt.Println("registered client", clientID)
		*resp = true
	}
	*resp = false
	return nil
}

func (sv *OTServer) ApplyOp(args *op.Op, resp *op.Op) error {
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
	return nil
}

func (sv *OTServer) Broadcast(args *op.Op, resp *op.Op) error {
	// if sv.clients[args.Uid] > args.Version {
	// 	return errors.New("Broadcast: client missing operation, out of order")
	// } 
	if sv.version > args.Version { // server ahead of client
		*resp = sv.logs[args.Version] // return the next log
		fmt.Println("broadcast to", args.Uid, resp)
	} else {
		fmt.Println("broadcast up to date", args.Uid, sv.version, sv.clients[args.Uid] )
		resp.OpType = "empty" // client is at same state as server
	}

	return nil
}

func (sv *OTServer) ApplyTransformation(args *op.Op, resp *op.Op) error {
	fmt.Println("incoming op ", args)
	if args.OpType == "ins" || args.OpType == "del" {
		// if sv.clients[args.Uid] != args.Version-1{
		// 	return errors.New("ApplyTransformation: client operation out of order")
		// } 
		sv.clients[args.Uid] = args.Version // update client version
		sv.version = args.Version // CHANGE THIS LATER
		sv.logs = append(sv.logs, *args) // CHANGE THIS LATER

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
	} else {
		return errors.New("ApplyTransformation: wrong operation input")
	}
	fmt.Println("ApplyTransformation now: ", strings.Replace(sv.currState, "\n", "\\n", -1))
	fmt.Println("Clients version", sv.clients, sv.version)
	fmt.Println("logs", sv.logs)


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

	rpc.Register(sv)
	rpc.Accept(inbound)
}
