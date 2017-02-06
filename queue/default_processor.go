package queue

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/foomo/shop/persistence"
	"gopkg.in/mgo.v2/bson"
)

//------------------------------------------------------------------
// ~ Interfaces
//------------------------------------------------------------------
type Processor interface {
	GetId() string
	GetPersistor() *persistence.Persistor
	GetQuery() *bson.M
	SetQuery(*bson.M)
	Process(interface{}) error
	Find(*bson.M, *persistence.Persistor) (func() (interface{}, error), error)
	GetMutex() *sync.Mutex
	GetRunningJobs() int
	IncRunningJobs()
	DecRunningJobs()
	GetMaxConcurrency() int
	SetMaxConcurrency(n int)
	GetChanRun() chan bool // this one is used by Scheduler to start and stop the processoer
	GetChanExit() chan int // this one is observed Scheduler and is used by Processor to indicdate that it exited by itself
	SetJobsAssigned(n int)
	GetJobsAssigned() int
	IncCountProcessed()
	GetCountProcessed() int
	GetWaitGroup() *sync.WaitGroup
	SetWaitGroupFinished(bool)
	GetWaitGroupFinished() bool
	GetMaxUsedConcurrency() int
	SetMaxUsedConcurrency(int)
	GetJobsStarted() int
	IncJobsStarted()
	SetStartTimeProcessing(int64)
	SetEndTimeProcessing(int64)
	GetProcessingTime() time.Duration
	GetTimePerJob() time.Duration
	Stop()
	//	ResetStop()
	GetStop() bool
	Report()
	Reset()
}

//------------------------------------------------------------------
// ~ Public Types
//------------------------------------------------------------------
// DefaultProcessor implements Processor. Customize this processor
// by setting a specific Persistor, ProcessingFunc and DataWrapper
type DefaultProcessor struct {
	waitGroup           *sync.WaitGroup
	Id                  string
	query               *bson.M
	mutex               *sync.Mutex
	RunningJobs         int
	CountProcessed      int
	chanRun             chan bool
	chanExit            chan int
	Persistor           *persistence.Persistor
	ProcessingFunc      func(interface{}) error // implements the actual processing
	GetDataWrapper      func() interface{}      // returns an truct to unmarshal db result into.
	isRunning           bool
	maxConcurrency      int
	maxUsedConcurrency  int
	jobsAssigned        int // Stop processor after jobsAssigned. This is mainly for testing.
	waitGroupFinished   bool
	jobsStarted         int
	stop                bool
	startTimeProcessing int64
	endTimeProcessing   int64
	Verbose             bool
}

//------------------------------------------------------------------
// ~ Public Methods
//------------------------------------------------------------------
func NewDefaultProcessor(id string) *DefaultProcessor {

	pr := &DefaultProcessor{
		Verbose:        true,
		Id:             id,
		waitGroup:      &sync.WaitGroup{},
		query:          &bson.M{},
		mutex:          &sync.Mutex{},
		chanRun:        make(chan bool),
		chanExit:       make(chan int),
		maxConcurrency: 16,
		ProcessingFunc: func(interface{}) error {

			log.Println("Nothing happening here... Specify a ProcessingFunc!")

			return nil
		},
	}
	return pr
}

func (proc *DefaultProcessor) GetPersistor() *persistence.Persistor {
	return proc.Persistor
}

func (proc *DefaultProcessor) GetChanRun() chan bool {
	return proc.chanRun
}
func (proc *DefaultProcessor) GetChanExit() chan int {
	return proc.chanExit
}

func (proc *DefaultProcessor) GetMutex() *sync.Mutex {
	return proc.mutex
}
func (proc *DefaultProcessor) GetRunningJobs() int {
	proc.mutex.Lock()
	n := proc.RunningJobs
	proc.mutex.Unlock()
	return n
}
func (proc *DefaultProcessor) IncRunningJobs() {
	proc.mutex.Lock()
	proc.RunningJobs = proc.RunningJobs + 1
	proc.SetMaxUsedConcurrency(proc.RunningJobs)
	proc.mutex.Unlock()
}
func (proc *DefaultProcessor) DecRunningJobs() {
	proc.mutex.Lock()
	proc.RunningJobs = proc.RunningJobs - 1
	proc.mutex.Unlock()
}
func (proc *DefaultProcessor) GetId() string {
	return proc.Id
}
func (proc *DefaultProcessor) GetQuery() *bson.M {
	return proc.query
}

func (proc *DefaultProcessor) SetQuery(query *bson.M) {
	proc.query = query
}

func (pr *DefaultProcessor) Process(data interface{}) error {
	return pr.ProcessingFunc(data)
}

func (proc *DefaultProcessor) GetMaxConcurrency() int {

	return proc.maxConcurrency
}
func (proc *DefaultProcessor) SetMaxConcurrency(n int) {
	proc.maxConcurrency = n
}
func (proc *DefaultProcessor) SetJobsAssigned(n int) {
	proc.jobsAssigned = n
}

func (proc *DefaultProcessor) GetJobsStarted() int {
	return proc.jobsStarted
}
func (proc *DefaultProcessor) IncJobsStarted() {
	proc.mutex.Lock()
	proc.jobsStarted = proc.jobsStarted + 1
	proc.mutex.Unlock()
}

func (proc *DefaultProcessor) GetJobsAssigned() int {
	return proc.jobsAssigned
}
func (proc *DefaultProcessor) GetCountProcessed() int {
	return proc.CountProcessed
}

func (proc *DefaultProcessor) GetWaitGroup() *sync.WaitGroup {
	return proc.waitGroup
}

func (proc *DefaultProcessor) SetWaitGroupFinished(finished bool) {
	proc.waitGroupFinished = finished
}
func (proc *DefaultProcessor) GetWaitGroupFinished() bool {
	return proc.waitGroupFinished
}
func (proc *DefaultProcessor) IncCountProcessed() {
	proc.mutex.Lock()
	proc.CountProcessed = proc.CountProcessed + 1
	proc.mutex.Unlock()
}

func (proc *DefaultProcessor) GetMaxUsedConcurrency() int {
	return proc.maxUsedConcurrency
}
func (proc *DefaultProcessor) SetMaxUsedConcurrency(current int) {
	if current > proc.maxUsedConcurrency {
		proc.maxUsedConcurrency = current
	}
}

func (proc *DefaultProcessor) SetStartTimeProcessing(t int64) {
	proc.startTimeProcessing = t
}
func (proc *DefaultProcessor) SetEndTimeProcessing(t int64) {
	proc.endTimeProcessing = t

}

// Stop Wait for all current Jobs to finish and stop processor
func (proc *DefaultProcessor) Stop() {
	proc.stop = true
}
func (proc *DefaultProcessor) GetStop() bool {
	return proc.stop
}

// func (proc *DefaultProcessor) ResetStop() {
// 	proc.stop = false
// }

func (proc *DefaultProcessor) GetProcessingTime() time.Duration {
	return time.Duration(proc.endTimeProcessing - proc.startTimeProcessing)
}
func (proc *DefaultProcessor) GetTimePerJob() time.Duration {
	return time.Duration(int64(float64(proc.GetProcessingTime()) / float64(proc.GetCountProcessed())))
}

func (proc *DefaultProcessor) Reset() {
	proc.SetWaitGroupFinished(false)
	proc.chanRun = make(chan bool)
	proc.chanExit = make(chan int)
	proc.isRunning = false
	proc.stop = false
}

// Find returns an iterator for all entries found matching on query.
func (proc *DefaultProcessor) Find(query *bson.M, p *persistence.Persistor) (iter func() (data interface{}, err error), err error) {
	if proc.Verbose {
		log.Println("Default Processor Find")
	}
	_, err = p.GetCollection().Find(query).Count()
	if err != nil {
		log.Println(err)
	}
	q := p.GetCollection().Find(query).Sort("_id")

	count, err := q.Count()
	if proc.Verbose {
		log.Println("Found", count, "items in database ", "("+proc.GetId()+")")
	}
	if err != nil {
		return
	}
	mgoiter := q.Iter()
	iter = func() (interface{}, error) {
		data := proc.GetDataWrapper()
		if mgoiter.Next(data) {
			return data, nil
		}
		return nil, nil
	}
	return
}

func (proc *DefaultProcessor) Report() {
	fmt.Println("")
	fmt.Println("----- STATISTICS", proc.GetId())
	fmt.Println("Total processing time:", proc.GetProcessingTime())
	fmt.Println("Processed Jobs:", proc.GetCountProcessed(), "Time per Job:", proc.GetTimePerJob())
	fmt.Println("Maximum allowed concurrency:", proc.GetMaxConcurrency())
	fmt.Println("Maximum used concurrency:", proc.GetMaxUsedConcurrency())
	fmt.Println("Jobs started:", proc.GetJobsStarted())
	fmt.Println("Jobs running:", proc.GetRunningJobs())
	fmt.Println("WaitgroupFinished:", proc.GetWaitGroupFinished())
	fmt.Println("")
}
