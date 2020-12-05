package core

import (
	"log"
	"fmt"
)
import "net"
import "net/rpc"
import "net/http"
import "net/http/httputil"
import "net/url"


type LoadBalancer struct {
	allServers []ServerInfo
}

//
// create a LB
//
func InitiationLB() *LoadBalancer {
	lb := LoadBalancer{}

	lb.server()
	return &lb
}

//
// start a thread that listens for RPCs
//
func (lb *LoadBalancer) server() {
	rpc.Register(lb)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")

	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// Server Registration
func (lb *LoadBalancer) RegisterServer(args *RegisterServerArgs, reply *RegisterServerReply) error {
	//TODO
	fmt.Println("Server Registered")
	return nil
}

// Server Report Load
func (lb *LoadBalancer) ReportLoad(args *ReportLoadArgs, reply *ReportLoadReply) error {
	//TODO
	return nil
}

// Transfer Request
func (lb *LoadBalancer) TransferRequest(res http.ResponseWriter, req *http.Request) http.ResponseWriter {
	//TODO 
	//Implement LB Algorithm
	url, _ := url.Parse("http://localhost:8081")
	proxy := httputil.NewSingleHostReverseProxy(url)

	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	proxy.ServeHTTP(res, req)
	return res
}