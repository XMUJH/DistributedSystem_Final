package main

//
// start the LoadBlancer process
//
// go run loadblancer.go
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

type LoadBlancerHandler struct {
	lb *LoadBlancer
}

func (lbh *LoadBlancerHandler) Transfer(res http.ResponseWriter, req *http.Request) {
	msg, err := ioutil.ReadAll(req.Body)
	if err != nil {
		res.Write([]byte(err.Error()))
		return
	}
	if req.URL.RequestURI() == "/favicon.ico" {
        return
    }

	//Load Blancer
	//TODO
	res = lbh.lb.TransferRequest(res, req)

	//Output Format
	writeLen, err := res.Write(msg)
	if err != nil || writeLen != len(msg) {
		log.Println(err, "write len:", writeLen)
	}
}

func main() {
	lbh := LoadBlancerHandler{}
	lbh.lb := core.Initiation()

	time.Sleep(100 * time.Millisecond)
	handleHttp(&lbh)
}

func handleHttp(lbh *LoadBlancerHandler) {
	http.HandleFunc("/", lbh.Transfer)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}