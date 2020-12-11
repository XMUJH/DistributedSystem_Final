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

func (sh *ServerHandler) HandleLightRequest(res http.ResponseWriter, req *http.Request) {
	msg, err := ioutil.ReadAll(req.Body)
	if err != nil {
		res.Write([]byte(err.Error()))
		return
	}
	if req.URL.RequestURI() == "/favicon.ico" {
        return
    }

	//Server Process Request
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
	ip, err := core.ExternalIP()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Server Running on "+ip.String()+":8081")

	sh := ServerHandler{}
	sh.s = core.InitiationServer(ip.String()+":8081")

	time.Sleep(100 * time.Millisecond)
	handleHttp(&sh, ip.String())
}

func handleHttp(sh *ServerHandler, ip string) {
	http.HandleFunc("/light", sh.HandleLightRequest)
	err := http.ListenAndServe(ip+":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}