package main

//
// start the Server process
//
// go run server.go
//

import "core"
import "time"
import "fmt"
import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type ServerHandler struct {
	s *core.Server
}

func (sh *ServerHandler) HandleRequest(res http.ResponseWriter, req *http.Request) {
	msg, err := ioutil.ReadAll(req.Body)
	if err != nil {
		res.Write([]byte(err.Error()))
		return
	}
	if req.URL.RequestURI() == "/favicon.ico" {
        return
    }

	//Server Process Request
	//TODO
	fmt.Println("Request Handled by Server")

	//Output Format
	noUse := "This msg is from server"
	type OutPut struct {
		NoUse string `json:"no_use"`
	}
	out := OutPut{
		NoUse: noUse,
	}

	data, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	res.Write(data)
	writeLen, err := res.Write(msg)
	if err != nil || writeLen != len(msg) {
		log.Println(err, "write len:", writeLen)
	}
}

func main() {
	sh := ServerHandler{}
	sh.s = core.InitiationServer()

	time.Sleep(100 * time.Millisecond)
	handleHttp(&sh)
}

func handleHttp(sh *ServerHandler) {
	http.HandleFunc("/", sh.HandleRequest)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}