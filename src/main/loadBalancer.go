package main

//
// start the LoadBalancer process
//
// go run loadbalancer.go
//

import "core"
import "time"
import "fmt"
import (
	"io/ioutil"
	"log"
	"net/http"
)

type LoadBalancerHandler struct {
	lb *core.LoadBalancer
}

func (lbh *LoadBalancerHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	msg, err := ioutil.ReadAll(req.Body)
	if err != nil {
		res.Write([]byte(err.Error()))
		return
	}
	if req.URL.RequestURI() == "/favicon.ico" {
        return
    }

	//Load Balancer
	res = lbh.lb.TransferRequest(res, req)
	
	//Output Format
	writeLen, err := res.Write(msg)
	if err != nil || writeLen != len(msg) {
		log.Println(err, "write len:", writeLen)
	}
}

func main() {
	ip, err := core.ExternalIP()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("LoadBalancer Running on "+ip.String()+":8080")

	lbh := LoadBalancerHandler{}
	lbh.lb = core.InitiationLB(ip.String())

	time.Sleep(100 * time.Millisecond)
	handleHttp(&lbh, ip.String())
}

func handleHttp(lbh *LoadBalancerHandler, ip string) {
	//http.HandleFunc("/", lbh.Transfer)
	err := http.ListenAndServe(ip+":8080", lbh)
	if err != nil {
		log.Fatal(err)
	}
}