package mr

import (
	"log"
	"sync"
)
import "net"
import "os"
import "net/rpc"
import "net/http"
import "time"

//locks
var mapTaskLock sync.Mutex
var reduceTaskLock sync.Mutex
var mapRwLock sync.Mutex
var reduceRwLock sync.Mutex

type MapTask struct {
	id int
	fileName string
}

type ReduceTask struct {
	id int
}

type MapWaitTask struct {
	task MapTask
	timeStamp int64
}

type ReduceWaitTask struct {
	task ReduceTask
	timeStamp int64
}

type Master struct {
	// Your definitions here.
	mapTasks []MapTask
	reduceTasks []ReduceTask
	mapWaitList map[int]MapWaitTask
	reduceWaitList map[int]ReduceWaitTask
	nReduce int
	nMap int
}

// Your code here -- RPC handlers for the worker to call.

//
// an example RPC handler.
//
func (m *Master) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}


//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	 os.Remove("mr-socket")
	l, e := net.Listen("unix", "mr-socket")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	ret := false

	// Your code here.
	mapFlag := false
	reduceFlag := false

	// judge Map Tasks
	if len(m.mapTasks) == 0 && len(m.mapWaitList) == 0 { 
		// make sure the mapTask is all Done
		mapRwLock.Lock()
		if len(m.mapTasks) == 0 && len(m.mapWaitList) == 0 {
			mapFlag = true
		}
		mapRwLock.Unlock()
	}

	// judge Reduce Tasks
	if len(m.reduceTasks) == 0 && len(m.reduceWaitList) == 0 { 
		// make sure the reduceTask is all Done
		reduceRwLock.Lock()
		if len(m.reduceTasks) == 0 && len(m.reduceWaitList) == 0 {
			reduceFlag = true
		}
		reduceRwLock.Unlock()
	}

	ret = mapFlag && reduceFlag

	return ret
}

//RegisterWorker is an RPC method that is called by workers after they have started
// up to report that they are ready to receive tasks.
func (m *Master) RegisterWorker(args *RegisterWorkerArgs, reply *RegisterWorkerReply) error {
	return nil
}

func (m *Master) monitorFailure() {
	for true {
		//monitor Map Tasks
		mapRwLock.Lock()
		for key := range m.mapWaitList {
			if(time.Now().Unix() - m.mapWaitList[key].timeStamp > 10) {
				task, ok := m.mapWaitList[key]
				if(ok) {
					var temp MapTask
					temp.fileName = task.task.fileName
					temp.id = key
					delete(m.mapWaitList, key)
					m.mapTasks = append(m.mapTasks, temp)
				}
			}
		}
		mapRwLock.Unlock()

		//monitor Reduce Tasks
		reduceRwLock.Lock()
		for key := range m.reduceWaitList {
			if(time.Now().Unix() - m.reduceWaitList[key].timeStamp > 10) {
				_, ok := m.reduceWaitList[key]
				if(ok) {
					var temp ReduceTask
					temp.id = key
					delete(m.reduceWaitList, key)
					m.reduceTasks = append(m.reduceTasks, temp)
				}
			}
		}
		reduceRwLock.Unlock()

		time.Sleep(time.Second * 10)
	}
}

//RequestTask is an RPC method that is called by workers to request a map or reduce task
func (m *Master) RequestTask(args *RequestTaskArgs, reply *RequestTaskReply) error {
	//request Map tasks
	if len(m.mapTasks) > 0 {
		mapTaskLock.Lock()
		if len(m.mapTasks) > 0 {
			mapRwLock.Lock()
			reply.Id = m.mapTasks[0].id
			reply.FileName = m.mapTasks[0].fileName
			reply.NReduce = m.nReduce
			reply.JobType = MapPhase
			reply.IsGetJob = true

			var temp MapWaitTask
			temp.timeStamp = time.Now().Unix()
			temp.task = m.mapTasks[0]

			//mapRwLock.Lock()
			m.mapWaitList[m.mapTasks[0].id] = temp
			m.mapTasks = m.mapTasks[1:]
			mapRwLock.Unlock()
		}
		mapTaskLock.Unlock()
	} else if len(m.mapTasks) == 0 && len(m.mapWaitList) == 0 { 
		flag := false
		// make sure the mapTask is all Done
		mapRwLock.Lock()
		if len(m.mapTasks) == 0 && len(m.mapWaitList) == 0 {
			flag = true
		}
		mapRwLock.Unlock()

		//request Reduce tasks
		if flag == true {
			if len(m.reduceTasks) > 0 {
				reduceTaskLock.Lock()
				if len(m.reduceTasks) > 0 {
					reduceRwLock.Lock()
					reply.Id = m.reduceTasks[0].id
					reply.NMap = m.nMap
					reply.JobType = ReducePhase
					reply.IsGetJob = true
		
					var temp ReduceWaitTask
					temp.timeStamp = time.Now().Unix()
					temp.task = m.reduceTasks[0]
		
					m.reduceWaitList[m.reduceTasks[0].id] = temp
					m.reduceTasks = m.reduceTasks[1:]
					reduceRwLock.Unlock()
				}
				reduceTaskLock.Unlock()
			}
		}
	}
	return nil
}

//ReportTask is an RPC method that is called by workers to report a task's status
//whenever a task is finished or failed
//HINT: when a task is failed, master should reschedule it.
func (m *Master) ReportTask(args *ReportTaskArgs, reply *ReportTaskReply) error {

	// report Map Task
	if(args.JobType == MapPhase) {
		if(args.Id != 0 && args.Status == true) {
			mapRwLock.Lock()
			_, ok := m.mapWaitList[args.Id]
			if(ok) {
				delete(m.mapWaitList, args.Id)
			}
			mapRwLock.Unlock()
		}
		if(args.Id != 0 && args.Status == false) {
			mapRwLock.Lock()
			task, ok := m.mapWaitList[args.Id]
			if(ok) {
				var temp MapTask
				temp.fileName = task.task.fileName
				temp.id = args.Id
				delete(m.mapWaitList, args.Id)
				m.mapTasks = append(m.mapTasks, temp)
			}
			mapRwLock.Unlock()
		}
	}

	//report Reduce Task
	if(args.JobType == ReducePhase) {
		if(args.Id != 0 && args.Status == true) {
			reduceRwLock.Lock()
			_, ok := m.reduceWaitList[args.Id]
			if(ok) {
				delete(m.reduceWaitList, args.Id)
			}
			reduceRwLock.Unlock()
		}
		if(args.Id != 0 && args.Status == false) {
			reduceRwLock.Lock()
			_, ok := m.reduceWaitList[args.Id]
			if(ok) {
				var temp ReduceTask
				temp.id = args.Id
				delete(m.reduceWaitList, args.Id)
				m.reduceTasks = append(m.reduceTasks, temp)
			}
			reduceRwLock.Unlock()
		}
	}

	return nil
}

//
// create a Master.
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	// Your code here.
	m.server()
	//initiate mapTasks
	var mapTasksTemp []MapTask
	var mapWaitListTemp = make(map[int]MapWaitTask)

	for i := 0; i<len(files); i++ {
		var temp MapTask
		temp.fileName = files[i]
		temp.id = i + 1
		mapTasksTemp = append(mapTasksTemp, temp)
	}
	m.mapTasks = mapTasksTemp
	m.nMap = len(files)
	m.mapWaitList = mapWaitListTemp

	//initiate reduceTasks
	var reduceTasksTemp []ReduceTask
	var reduceWaitListTemp = make(map[int]ReduceWaitTask)

	for i := 0; i<nReduce; i++ {
		var temp ReduceTask
		temp.id = i + 1
		reduceTasksTemp = append(reduceTasksTemp, temp)
	}
	m.reduceTasks = reduceTasksTemp
	m.nReduce = nReduce
	m.reduceWaitList = reduceWaitListTemp

	//start a routine to monitor Worker Failure
	go m.monitorFailure()

	return &m
}
