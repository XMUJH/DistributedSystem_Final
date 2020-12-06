package core

import (
	"log"
	"fmt"
	"sort"
)
import "net"
import "net/rpc"
import "net/http"
import "net/http/httputil"
import "net/url"
import "strconv"
import "sync"

var mapLock sync.Mutex

type LoadBalancer struct {
	allServers map[string]float64
	index map[string]int
	serverCnt int
}

//
// create a LB
//
func InitiationLB(ip string) *LoadBalancer {
	lb := LoadBalancer{}
	lb.allServers = make(map[string]float64)
	lb.index = make(map[string]int)
	lb.serverCnt = 0

	lb.server(ip)
	return &lb
}

//
// start a thread that listens for RPCs
//
func (lb *LoadBalancer) server(ip string) {
	rpc.Register(lb)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ip+":1234")

	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// Server Registration
func (lb *LoadBalancer) RegisterServer(args *RegisterServerArgs, reply *RegisterServerReply) error {

	mapLock.Lock()
	lb.allServers[args.Info.Address] = args.Info.Load
	lb.serverCnt += 1
	lb.index[args.Info.Address] = lb.serverCnt
	mapLock.Unlock()

	fmt.Println("Server Registered")
	fmt.Println(lb.index)

	return nil
}

// Server Report Load
func (lb *LoadBalancer) ReportLoad(args *ReportLoadArgs, reply *ReportLoadReply) error {
	
	mapLock.Lock()
	_, ok := lb.allServers[args.Info.Address]
	if(ok) {
		lb.allServers[args.Info.Address] = args.Info.Load;
		mapLock.Unlock()
		//fmt.Println("Load Reported")
	} else {
		mapLock.Unlock()
		args2 := RegisterServerArgs{}
		reply2 := RegisterServerReply{}
		args2.Info = args.Info
		lb.RegisterServer(&args2, &reply2)
	}

	return nil
}

// Transfer Request
func (lb *LoadBalancer) TransferRequest(res http.ResponseWriter, req *http.Request) http.ResponseWriter {

	//LB Algorithm
	var listServer = []ServerInfo{}
	mapLock.Lock()
	for k, v := range lb.allServers {
		listServer = append(listServer, ServerInfo {k, v})
	}
	mapLock.Unlock()

	sort.Slice(listServer, func(i, j int) bool {
		return listServer[i].Load < listServer[j].Load 
	})

	fmt.Println(listServer)

	//Reverse Proxy
	if(len(listServer)>0) {
		url, _ := url.Parse("http://"+listServer[0].Address)
		proxy := httputil.NewSingleHostReverseProxy(url)

		req.URL.Host = url.Host
		req.URL.Scheme = url.Scheme
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Host = url.Host

		proxy.ServeHTTP(res, req)
		fmt.Println("Request Transferred to server "+strconv.Itoa(lb.index[listServer[0].Address]))
	} else {
		fmt.Println("Request Transfer Failed Because No Active Server Exists")
	}
	return res
}