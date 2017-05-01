// https://gist.github.com/jordanorelli/2629049

package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"op"
)

type OTServer struct {
	logs 		[]op.Op
	currState	string	//string of the most updated version
	version		int		//last updated version
}

func (sv *OTServer) ApplyOp(op *op.Op, ack *bool) error {

	fmt.Println(op)
	fmt.Println("currState")
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

	rpc.Register(sv)
	rpc.Accept(inbound)
}
