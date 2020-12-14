package core

import (
	"log"
	"fmt"
)
import "net/rpc"
import "github.com/shirou/gopsutil/cpu"
import "time"


type Server struct {
	Info ServerInfo 
}

func GetCpuPercent() float64 {
	percent, _:= cpu.Percent(time.Second, false)
	return percent[0]
}

//
// create a Server
//
func InitiationServer(address string) *Server {
	s := Server{}
	s.Info.Address = address
	s.Info.Load = GetCpuPercent()

	s.Registration()

	go s.keepRefresh()

	return &s
}

// Server Registration
func (s *Server) Registration() {
	args := RegisterServerArgs{}
	reply := RegisterServerReply{}

	args.Info = s.Info

	call("LoadBalancer.RegisterServer", &args, &reply)
}

// Report Load
func (s *Server) Report() {
	args := ReportLoadArgs{}
	reply := ReportLoadReply{}

	args.Info = s.Info

	call("LoadBalancer.ReportLoad", &args, &reply)
}

//keep refresh Load on the server
func (s * Server) keepRefresh() {
	for true {
		time.Sleep(time.Millisecond * 333)

		s.Info.Load = GetCpuPercent()
		s.Report()
	}
}

//
// call for RPCs
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	c, err := rpc.DialHTTP("tcp", "172.17.40.2:1234")
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