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
	//"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type LoadBalancerHandler struct {
	lb *core.LoadBalancer
}

func (lbh *LoadBalancerHandler) Transfer(res http.ResponseWriter, req *http.Request) {
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
	fmt.Println("Request Transferred by LB")

	//Output Format
	writeLen, err := res.Write(msg)
	if err != nil || writeLen != len(msg) {
		log.Println(err, "write len:", writeLen)
	}
}

func main() {
	lbh := LoadBalancerHandler{}
	lbh.lb = core.InitiationLB()

	time.Sleep(100 * time.Millisecond)
	handleHttp(&lbh)
}

func handleHttp(lbh *LoadBalancerHandler) {
	http.HandleFunc("/", lbh.Transfer)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}