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
import "crypto/rand"
import "math/big"
import "strconv"
import "crypto/sha1"

type ServerHandler struct {
	s *core.Server
}

func getRandomString() string {
	tag := make([]byte, 100)
	for i:=0;i<100;i++ {
		result, _ := rand.Int(rand.Reader, big.NewInt(int64(26)))
		index, _ := strconv.Atoi(result.String())
		b := index + 65
        tag[i] = byte(b)
	}

	return string(tag)
}

func getHeavyRandomString() string {
	tag := make([]byte, 100)
	for i:=0;i<100;i++ {
	 result, _ := rand.Int(rand.Reader, big.NewInt(int64(26)))
	 index, _ := strconv.Atoi(result.String())
	 b := index + 65
		   tag[i] = byte(b)
	}
   
	return string(tag)
}

func getLightRandomString() string {
	tag := make([]byte, 100)
	for i:=0;i<2;i++ {
	 result, _ := rand.Int(rand.Reader, big.NewInt(int64(26)))
	 index, _ := strconv.Atoi(result.String())
	 b := index + 65
		   tag[i] = byte(b)
	}
   
	return string(tag)
}

func light() {
 
	M := 500
	//start := time.Now()
	for i := 0; i < M; i++ {
	 tagString := getLightRandomString()
	 //fmt.Println(tagString)
	 Sha1Inst := sha1.New()
	 Sha1Inst.Write([]byte(tagString))
	 _ = Sha1Inst.Sum([]byte(""))
	 time.Sleep(1000000  * time.Nanosecond)
	}
	//during := time.Since(start)
	//fmt.Println("Running time: ", during)
	//wg.Done()
}

func heavy() {
 
	M := 1500
	//start := time.Now()
	for i := 0; i < M; i++ {
	 tagString := getHeavyRandomString()
	 //fmt.Println(tagString)
	 Sha1Inst := sha1.New()
	 Sha1Inst.Write([]byte(tagString))
	 _ = Sha1Inst.Sum([]byte(""))
	 time.Sleep(3000000  * time.Nanosecond)
	}
	//during := time.Since(start)
	//fmt.Println("Running time: ", during)
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
	//Hash
	// for i:=0;i<1;i++ {
	// 	tagString := getRandomString()
	// 	//fmt.Println(tagString)
	// 	Sha1Inst := sha1.New()
	// 	Sha1Inst.Write([]byte(tagString))
	// 	_ = Sha1Inst.Sum([]byte(""))
	// }

	//Loop
	// tool :=0
	// for i:= 0; i<10000000; i++ {
	// 	tool = 1
	// }
	//fmt.Println(tool)

	light()

	//fmt.Println("Light Request Handled by Server")

	//Output Format
	//noUse := strconv.Itoa(tool)
	noUse := "This msg is from server (light)"
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

func (sh *ServerHandler) HandleHeavyRequest(res http.ResponseWriter, req *http.Request) {
	msg, err := ioutil.ReadAll(req.Body)
	if err != nil {
		res.Write([]byte(err.Error()))
		return
	}
	if req.URL.RequestURI() == "/favicon.ico" {
        return
    }

	//Server Process Request
	//Hash
	// for i:=0;i<500;i++ {
	// 	tagString := getRandomString()
	// 	//fmt.Println(tagString)
	// 	Sha1Inst := sha1.New()
	// 	Sha1Inst.Write([]byte(tagString))
	// 	_ = Sha1Inst.Sum([]byte(""))
	// }

	//Sleep
	// for i:=0;i<100;i++ {
	// 	time.Sleep(time.Millisecond * 10)
	// }

	//Loop
	// tool :=0
	// for i:= 0; i<1000000000; i++ {
	// 	tool = 1
	// }
	//fmt.Println(tool)

	heavy()
	
	//fmt.Println("Heavy Request Handled by Server")

	//Output Format
	//noUse := strconv.Itoa(tool)
	noUse := "This msg is from server (heavy)"
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
	http.HandleFunc("/heavy", sh.HandleHeavyRequest)
	err := http.ListenAndServe(ip+":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}