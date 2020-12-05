package core

import (
	"log"
)
import "net"
import "net/rpc"
import "net/http"
import "net/http/httputil"
import "net/url"


type LoadBlancer struct {
	allServers []Server
}

//
// create a LB
//
func Initiation() *LoadBlancer {
	lb := LoadBlancer{}

	lb.server()
	return &lb
}

//
// start a thread that listens for RPCs
//
func (lb *LoadBlancer) server() {
	rpc.Register(lb)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")

	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// Server Registration
func (lb *LoadBlancer) RegisterServer(args *RegisterServerArgs, reply *RegisterServerReply) error {
	return nil
}

// Transfer Request
func (lb *LoadBlancer) TransferRequest(res http.ResponseWriter, req *http.Request) http.ResponseWriter {
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