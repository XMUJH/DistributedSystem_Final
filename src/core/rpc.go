package mr

//
// RPC definitions.
//

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

type RegisterWorkerArgs struct {

}

type RegisterWorkerReply struct {

}

type RequestTaskArgs struct {

}

type RequestTaskReply struct {
	FileName string
	Id int
	NReduce int
	NMap int
	JobType JobPhase
	IsGetJob bool
}

type ReportTaskArgs struct {
	JobType JobPhase
	Id int
	Status bool

}

type ReportTaskReply struct {

}
// Add your RPC definitions here.

