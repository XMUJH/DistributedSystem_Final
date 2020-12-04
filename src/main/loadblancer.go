package main

//
// start the LoadBlancer process
//
// go run loadblancer.go
//

//import "mr"
import "time"
import "fmt"
import (
	//"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)


func LoadBlancerHandler(res http.ResponseWriter, req *http.Request) {
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
	fmt.Println("This is a Load Blancer")
	url, _ := url.Parse("http://localhost:8081")
	proxy := httputil.NewSingleHostReverseProxy(url)

	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	proxy.ServeHTTP(res, req)

	//Output Format
	writeLen, err := res.Write(msg)
	if err != nil || writeLen != len(msg) {
		log.Println(err, "write len:", writeLen)
	}
}

func main() {
	time.Sleep(100 * time.Millisecond)
	handleHttp()
}

func handleHttp() {
	http.HandleFunc("/", LoadBlancerHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}