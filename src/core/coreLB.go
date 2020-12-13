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
import "crypto/rand"
import "time"
import "math/big"
import "encoding/csv"
import "os"

var mapLock sync.Mutex
var rrLock sync.Mutex
var benchMarkLock sync.Mutex
var startLock sync.Mutex

type LoadBalancer struct {
	allServers map[string]float64
	index map[string]int
	serverCnt int
	originalList []string
	lastServer int

	proxyMap map[string]*httputil.ReverseProxy

	maxDmin []float64
	loadMonitor map[string][]float64
	requestCnt map[string]int

	isStart bool
}

//
// create a LB
//
func InitiationLB(ip string) *LoadBalancer {
	lb := LoadBalancer{}
	lb.allServers = make(map[string]float64)
	lb.index = make(map[string]int)
	lb.proxyMap = make(map[string]*httputil.ReverseProxy)
	lb.serverCnt = 0
	lb.originalList = []string{}
	lb.lastServer = -1;
	lb.maxDmin = []float64{}
	lb.loadMonitor = make(map[string][]float64)
	lb.requestCnt = make(map[string]int)
	lb.isStart = false

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
	lb.originalList = append(lb.originalList, args.Info.Address)

	url, _ := url.Parse("http://"+args.Info.Address)
	proxy := httputil.NewSingleHostReverseProxy(url)
	lb.proxyMap[args.Info.Address] = proxy
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
func (lb *LoadBalancer) TransferRequest(res http.ResponseWriter, req *http.Request) {
	//Start Benchmarks
	if(lb.isStart == false) {
		startLock.Lock()
		if(lb.isStart == false) {
			lb.isStart = true
			go lb.benchmarks()
		}
		startLock.Unlock()
	}

	//Reverse Proxy
	if(len(lb.allServers)>0) {
		//LB Algorithm
		//dist := lb.minLoad()

		//rrLock.Lock()
		//dist := lb.roundRobin()
		//rrLock.Unlock()

		dist := lb.randomSelect()

		url, _ := url.Parse("http://"+dist)

		req.URL.Host = url.Host
		req.URL.Scheme = url.Scheme
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Host = url.Host

		lb.proxyMap[dist].ServeHTTP(res, req)
		fmt.Println("Request Transferred to server "+strconv.Itoa(lb.index[dist]))

		benchMarkLock.Lock()
		lb.requestCnt[dist]++;
		benchMarkLock.Unlock()
	} else {
		fmt.Println("Request Transfer Failed Because No Active Server Exists")
	}
}

// Min Load
func (lb *LoadBalancer) minLoad() string {
	var listServer = []ServerInfo{}
	mapLock.Lock()
	for k, v := range lb.allServers {
		listServer = append(listServer, ServerInfo {k, v})
	}
	mapLock.Unlock()

	sort.Slice(listServer, func(i, j int) bool {
		return listServer[i].Load < listServer[j].Load 
	})

	//fmt.Println(listServer)

	return listServer[0].Address
}

//Round Robin
func (lb *LoadBalancer) roundRobin() string {
	lb.lastServer++
	if(lb.lastServer>=len(lb.originalList)) {
		lb.lastServer = 0
	}

	return lb.originalList[lb.lastServer]
}

//Random Selection
func (lb *LoadBalancer) randomSelect() string {
	
	result, _ := rand.Int(rand.Reader, big.NewInt(int64(lb.serverCnt)))
	index, _ := strconv.Atoi(result.String())

	return lb.originalList[index]
}

//benchmarks
func (lb *LoadBalancer) benchmarks() {
	for true {
		time.Sleep(time.Second * 1)

		mapLock.Lock()
		for k, v := range lb.allServers {
			lb.loadMonitor[k]=append(lb.loadMonitor[k], v)
		}
		mapLock.Unlock()

		var cntList = []int{}
		flag := false
		benchMarkLock.Lock()
		for k, v := range lb.requestCnt {
			cntList = append(cntList, v)
			if(v!=0) {
				flag=true
			}
			lb.requestCnt[k] = 0
		}
		benchMarkLock.Unlock()

		sort.Slice(cntList, func(i, j int) bool {
			return cntList[i] < cntList[j]
		})

		if(len(cntList)!=0) {
			lb.maxDmin = append(lb.maxDmin, float64(cntList[len(cntList)-1])/float64(cntList[0]))
		}

		if(flag==false) {
			// for k, v := range lb.loadMonitor {
			// 	fmt.Println(k)
			// 	fmt.Println(v)
			// }
			for i := 0; i<len(lb.originalList);i++ {
				fmt.Println(lb.originalList[i])
				fmt.Println(lb.loadMonitor[lb.originalList[i]])
			}

			//Write Resutl
			file, err := os.OpenFile("serverLoad.csv", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
			if err != nil {
				fmt.Println("open file is failed, err: ", err)
		     }
			defer file.Close()
			
			file.WriteString("\xEF\xBB\xBF")
			w := csv.NewWriter(file)
			writeList := []string{}

			writeList=append(writeList, "time")
			for i := 0; i<len(lb.originalList);i++ {
				writeList = append(writeList, lb.originalList[i])
			}
			w.Write(writeList)
			w.Flush()
			writeList = []string{}
			
			timeIndex := 1
			for i:= 0; i<len(lb.loadMonitor[lb.originalList[0]]); i++{
				writeList=append(writeList, strconv.Itoa(timeIndex))
				timeIndex++
				for j := 0; j<len(lb.originalList);j++ {
					writeList=append(writeList, strconv.FormatFloat(lb.loadMonitor[lb.originalList[j]][i], 'E', -1, 64))
				}
				w.Write(writeList)
				w.Flush()
				writeList = []string{}
			}

			//fmt.Println("max Divide min")
			//fmt.Println(lb.maxDmin)
			lb.isStart = false
			lb.maxDmin = []float64{}
			lb.loadMonitor = make(map[string][]float64)
			lb.requestCnt = make(map[string]int)
			return 
		}

	}
}