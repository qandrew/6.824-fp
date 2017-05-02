// https://gist.github.com/jordanorelli/2629049

package main

import (
	"errors"
	"fmt"
	// "log"
	"net"
	// "net/rpc"
	"op"
	"strings"
	"sync"

 	"github.com/cenkalti/rpc2"
)

type OTServer struct {
	mu        sync.Mutex // Lock to protect shared access to this peer's state
	Rpc_client *rpc2.Client
	logs      []op.Op
	currState string        //string of the most updated version
	version   int           //last updated version
	clients   map[int64]int //*rpc.Client keep track of version num
}

func (sv *OTServer) Init(clientID int64, resp *bool) error {
	if sv.clients[clientID] == 0 {
		sv.clients[clientID] = 1
		fmt.Println("registered client", clientID)
		*resp = true
	}
	*resp = false
	return nil
}

func (sv *OTServer) ApplyOp(args *op.Op, resp *op.Op) error {
	// operation called by the client
	// fmt.Println(args)
	sv.logs = append(sv.logs, *args)

	err := sv.ApplyTransformation(args, resp) // do the actual OT here

	return err
}

func (sv *OTServer) GetSnapshot(req *op.Snapshot, resp *op.Snapshot) error {
	resp.Value = sv.currState
	fmt.Println("snapshot returned to")
	return nil
}

func (sv *OTServer) ApplyTransformation(args *op.Op, resp *op.Op) error {
	fmt.Println("incoming op ", args)
	if args.OpType == "ins" {
		// sv.currState += args.Payload
		if args.Position == 0 {
			sv.currState = args.Payload + sv.currState // append at beginning
		} else {
			sv.currState = sv.currState[:args.Position] + args.Payload + sv.currState[args.Position:]
		}
	} else if args.OpType == "del" {
		if args.Position == len(sv.currState) && len(sv.currState) != 0 {
			sv.currState = sv.currState[:args.Position-1]
		} else {
			sv.currState = sv.currState[:args.Position-1] + sv.currState[args.Position:]
		}
	} else {
		return errors.New("ApplyTransformation: wrong operation input")
	}
	fmt.Println("ApplyTransformation: now", strings.Replace(sv.currState, "\n", "\\n", -1))

	return nil
}

func main() {
	// addy, err := net.ResolveTCPAddr("tcp", "localhost:42586")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// inbound, err := net.ListenTCP("tcp", addy)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// sv := new(OTServer)
	// sv.clients = make(map[int64]int) //*rpc.Client)

	// rpc.Register(sv)
	// rpc.Accept(inbound)

	sv := new(OTServer)
	sv.clients = make(map[int64]int) //*rpc.Client)

	srv := rpc2.NewServer()
	srv.Handle("Init", func(client *rpc2.Client, clientID int64, resp *bool) error{

	  // Reversed call (server to client)
	  // var rep Reply
	  // client.Call("mult", Args{2, 3}, &rep)
	  fmt.Println("server received:", clientID)

	  *resp = true
	  return nil
	})

	srv.Handle("ApplyOp", func(client *rpc2.Client, args *op.Op, resp *op.Op) error{
		fmt.Println("operation received", args)
		return sv.ApplyOp(args,resp)
	})

	srv.Handle("GetSnapshot",func(client *rpc2.Client, req *op.Snapshot, resp *op.Snapshot) error{
		fmt.Println("snapshot received", req)
		return sv.GetSnapshot(req,resp)
	})

	lis, _ := net.Listen("tcp", "127.0.0.1:5000")
	srv.Accept(lis)
}
