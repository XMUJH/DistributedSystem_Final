package main

//
// cpu test 
//
// go run cpuTest.go
//

import "time"
import "github.com/shirou/gopsutil/cpu"
import "fmt"

func GetCpuPercent() float64 {
	percent, _:= cpu.Percent(time.Second, false)
	return percent[0]
}

func main() {
	for true {
		time.Sleep(time.Second * 1)

		fmt.Println(GetCpuPercent())

	}
}

