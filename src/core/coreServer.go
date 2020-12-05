package core

import (
	"log"
	"fmt"
)
import "net/rpc"


type Server struct {
}

//
// create a Server
//
func InitiationServer() *Server {
	s := Server{}
	s.Registration()

	//TODO report load process

	return &s
}

//
// call for RPCs
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	c, err := rpc.DialHTTP("tcp", ":1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

// Server Registration
func (s *Server) Registration() {
	args := RegisterServerArgs{}
	reply := RegisterServerReply{}

	call("LoadBalancer.RegisterServer", &args, &reply)
}

// Report Load
func (s *Server) Report() {
	args := ReportLoadArgs{}
	reply := ReportLoadReply{}

	call("LoadBalancer.ReportLoad", &args, &reply)
}