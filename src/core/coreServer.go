package core

import "fmt"
import "log"
import "net/rpc"
import "hash/fnv"
import "os"
import "io/ioutil"
import "encoding/json"
import "sort"
import "time"


//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

func doMap(mapf func(string, string) []KeyValue, reply RequestTaskReply) {
	//do map job
	intermediate := []KeyValue{}
	file, err := os.Open(reply.FileName)
	if err != nil {
		Report(reply, false)
		log.Fatalf("cannot open %v", reply.FileName)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		Report(reply, false)
		log.Fatalf("cannot read %v", reply.FileName)
	}
	file.Close()
	intermediate = mapf(reply.FileName, string(content))

	//encode json
	var intermediateFiles []*os.File
	for i := 1; i <= reply.NReduce; i++ {
		intermediateFileName := reduceName(reply.Id, i)
		temp, _ := os.OpenFile(intermediateFileName, os.O_CREATE|os.O_WRONLY, 0777)
		intermediateFiles = append(intermediateFiles, temp)
		defer temp.Close()
	}

	for i := 0; i < len(intermediate); i++ {
		index := ihash(intermediate[i].Key) % reply.NReduce
		enc := json.NewEncoder(intermediateFiles[index])
		err := enc.Encode(&intermediate[i])
		if err != nil {
			Report(reply, false)
			log.Fatal(err)
		}
	}
	
	Report(reply, true)
}

func doReduce(reducef func(string, []string) string, reply RequestTaskReply) {

	//decode json
	var intermediateFiles []*os.File
	for i := 1; i <= reply.NMap; i++ {
		intermediateFileName := reduceName(i, reply.Id)
		temp, _ := os.OpenFile(intermediateFileName, os.O_RDONLY, 0777)
		intermediateFiles = append(intermediateFiles, temp)
		defer temp.Close()
	}

	intermediate := []KeyValue{}
	for i := 0; i < len(intermediateFiles); i++ {
		dec := json.NewDecoder(intermediateFiles[i])
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			intermediate = append(intermediate, kv)
		}
	}

	//do reduce job
	sort.Sort(ByKey(intermediate))
	oname := mergeName(reply.Id)
	ofile, _ := os.OpenFile(oname, os.O_CREATE|os.O_WRONLY, 0777)
	defer ofile.Close()

	i := 0
	for i < len(intermediate) {
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		output := reducef(intermediate[i].Key, values)

		// this is the correct format for each line of Reduce output.
		fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)

		i = j
	}

	Report(reply, true)
}

func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	for true {
		Registration()
		reply := RequestTask()

		if reply.IsGetJob && reply.JobType == MapPhase {
			go doMap(mapf, reply)
		}else if reply.IsGetJob && reply.JobType == ReducePhase {
			go doReduce(reducef, reply)
		}

		time.Sleep(time.Second * 2)
	}
	// uncomment to send the Example RPC to the master.
	// CallExample()

}

func Registration() {
	args := RegisterWorkerArgs{}
	reply := RegisterWorkerReply{}

	call("Master.RegisterWorker", &args, &reply)
}

func RequestTask() RequestTaskReply {
	args := RequestTaskArgs{}
	reply := RequestTaskReply{}

	call("Master.RequestTask", &args, &reply)

	return reply
}

func Report(job RequestTaskReply, status bool) {
	args := ReportTaskArgs{}
	reply := ReportTaskReply{}
	args.Id = job.Id;
	args.JobType = job.JobType
	args.Status = status;

	call("Master.ReportTask", &args, &reply)
}

//
// example function to show how to make an RPC call to the master.
//
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	call("Master.Example", &args, &reply)

	// reply.Y should be 100.
	fmt.Printf("reply.Y %v\n", reply.Y)
}

//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	//c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	 c, err := rpc.DialHTTP("unix", "mr-socket")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
