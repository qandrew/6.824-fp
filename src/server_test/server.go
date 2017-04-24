// https://gist.github.com/jordanorelli/2629049

package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"op"
)

type Listener int

func (l *Listener) ApplyOp(op *op.Op, ack *bool) error {
	fmt.Println(op)
	return nil
}

func main() {
	addy, err := net.ResolveTCPAddr("tcp", "0.0.0.0:42586")
	if err != nil {
		log.Fatal(err)
	}

	inbound, err := net.ListenTCP("tcp", addy)
	if err != nil {
		log.Fatal(err)
	}

	listener := new(Listener)
	rpc.Register(listener)
	rpc.Accept(inbound)
}
