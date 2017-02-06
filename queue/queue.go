package queue

// Package process handles the processing of Datas as they change their status

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

type Queue struct {
	Verbose                bool
	running                bool
	processors             []Processor
	waitGroup              *sync.WaitGroup
	chanFinished           chan int
	chanFinishedRegistered bool // true, if GetChanFinished has been called
}

//------------------------------------------------------------------
// ~ CONSTRUCTORS
//------------------------------------------------------------------

func NewQueue() *Queue {
	return &Queue{
		chanFinished: make(chan int),
	}
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

func (q *Queue) getChanFinished() chan int {
	q.chanFinishedRegistered = true
	return q.chanFinished
}

// AddProcessor Add a processor to the queue. Do not add multiple processors for the same task!
func (q *Queue) AddProcessor(processor Processor) {
	q.processors = append(q.processors, processor)
	if q.Verbose {
		log.Println("Added Processor to queue. New length: ", len(q.processors))
	}
	for _, p := range q.processors {
		if q.Verbose {
			fmt.Println("\t", p.GetId())
		}
	}
}

func (q *Queue) GetProcessors() []Processor {
	return q.processors
}

func (q *Queue) IsRunning() bool {
	return q.running
}

// Start continuously tries to run all available processors until all assigned jobs have been cpmpleted or
// Stop() has been called.
// As long as there is only one processor per kind (Status, Payment, ...), there should be no race conditions
func (q *Queue) Start() error {
	if q.Verbose {
		log.Println("Queue: Schedule Start")
	}
	if q.IsRunning() {
		return errors.New("Did not start Queue. It's already running!")
	}
	q.running = true
	q.waitGroup = &sync.WaitGroup{}
	for _, proc := range q.processors {
		q.waitGroup.Add(1)
		go schedule(proc, q.waitGroup)
		fmt.Println("New go routine 8")
	}
	q.waitGroup.Wait()
	q.running = false
	// for _, proc := range q.processors {
	// 	proc.ResetStop()
	// }
	if q.Verbose {
		fmt.Println("")
		fmt.Println("*****------------------------------------****")
	}
	for _, proc := range q.processors {
		proc.Report()
	}
	for _, proc := range q.processors {
		proc.Reset()
	}
	if q.Verbose {
		fmt.Println("*****------------------------------------****")
	}
	if q.chanFinishedRegistered { // if true, we suppose that some is listening to the channel. Because otherwise we would block here forever!
		q.chanFinished <- 1
		q.chanFinishedRegistered = false
	}
	return nil
}

func (q *Queue) Stop() {
	if q.Verbose {
		log.Println("Queue: Schedule Stop")
	}
	chanQueueFinished := q.getChanFinished()
	for _, proc := range q.processors {
		proc.Stop()
	}
	<-chanQueueFinished

}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

func schedule(proc Processor, waitGroup *sync.WaitGroup) {
	chanStart := make(chan int)
	chanStop := make(chan int) // this is called by ScheduleStop()

	go func() {
		fmt.Println("New go routine 7")
		for {
			select {
			case <-proc.GetChanExit():
				//fmt.Println("exiting")
				if proc.GetCountProcessed() < proc.GetJobsAssigned() && !proc.GetStop() {
					time.Sleep(100 * time.Millisecond) // wait a moment and try again
					chanStart <- 1
				} else {
					chanStop <- 1
				}
			}
		}
	}()

	go func() {
		fmt.Println("New go routine 6")
		for {
			select {
			case <-chanStart:

				//log.Println("Started Processor:", proc.GetId())

				runProcessor(proc)
			}
		}
	}()

	go func() {
		fmt.Println("New go routine 5")
		for {
			select {
			// Initially start processor or stop processing completely
			case run := <-proc.GetChanRun():
				if run {
					chanStart <- 1
				} else {
					chanStop <- 1
				}
			}
		}
	}()
	proc.GetChanRun() <- true
	<-chanStop
	waitGroup.Done()
	//	log.Println("Exiting schedule:", proc.GetId())

}

func runProcessor(processor Processor) error {
	processor.SetStartTimeProcessing(time.Now().UnixNano())
	chanDone := make(chan int)
	chanDone2 := make(chan int)
	chanReady := make(chan interface{})
	chanCheckRunning := make(chan int)
	chanGoCheckRunning := make(chan int)

	iter, err := processor.Find(processor.GetQuery(), processor.GetPersistor())
	if err != nil {
		log.Println(err)
		processor.GetChanExit() <- 1 // this will be received in schedule() and stop the processor
		return err
	}

	// Process availably data
	go func() {
		//fmt.Println("New go routine 1")
		for {
			select {
			case data := <-chanReady:
				// fmt.Println("--- DATA")
				// fmt.Println(spew.Sdump(data))
				f := func(data interface{}) {
					err := processor.Process(data)
					if err != nil {
						log.Println(err)
					}
					processor.IncCountProcessed()
					processor.DecRunningJobs()
					processor.GetWaitGroup().Done()
					//	fmt.Println("Return 1")
					return
				}

				go f(data)
				fmt.Println("New go routine 2")
				chanCheckRunning <- 1
			case <-chanDone2:
				//	fmt.Println("Return 1")
				return

			}
		}
		//fmt.Println("Exiting 1")
	}()

	// Wait for chanDone to end current processing
	go func() {
		//fmt.Println("New go routine 3")
		for {
			select {
			case <-chanDone:
				// Wait for jobs to finsish and exit
				//log.Println("** chanDone", processor.GetId())
				processor.GetWaitGroup().Wait()
				processor.SetWaitGroupFinished(true)
				processor.GetChanExit() <- 1
				processor.SetEndTimeProcessing(time.Now().UnixNano())
				chanDone2 <- 1
				//		fmt.Println("Return 3")
				return

			case <-chanGoCheckRunning:
				//log.Println("** goCheckRunning", processor.GetId())
				chanCheckRunning <- 1
			}
		}

		//	fmt.Println("Exiting 3")
	}()
	go func() {
		//	fmt.Println("New go routine 4")
		run := true
		//Loop:
		for run {
			select {
			case <-chanCheckRunning:
				if processor.GetStop() {
					//			fmt.Println("--- chanDone return 1")
					chanDone <- 1
					run = false
					//			fmt.Println("Return 4")
					return //break Loop
				}
				//log.Println("** chanCheckRunning", processor.GetId())
				if processor.GetJobsStarted() >= processor.GetJobsAssigned() { // We are done
					//	log.Println("** goChanDone :: JobsAssignedProcessed", processor.GetId())
					chanDone <- 1
					//			fmt.Println("--- chanDone return 2")
					return // break Loop
				} else if processor.GetRunningJobs() >= processor.GetMaxConcurrency() { // Wait for better times

					//	log.Println("** wait", processor.GetId())
					chanGoCheckRunning <- 1
				} else {
					if processor.GetJobsStarted() < processor.GetJobsAssigned() {
						data, err := iter()
						if err != nil {
							log.Println("Error: Could not get data", err)
							chanDone <- 1
							//			fmt.Println("--- chanDone return 3")
							return //break
						}
						if data != nil {
							processor.GetWaitGroup().Add(1)
							processor.IncJobsStarted()
							processor.IncRunningJobs() // We count here also, because we cannot access current number of jobs through waitgroup
							chanReady <- data
						} else {
							//	log.Println("data is nil", err)
							chanDone <- 1
							//	fmt.Println("--- chanDone return 4")
							return // break Loop
						}
					}
				}

			}
		}
		//fmt.Println("Exiting 4")
	}()
	chanCheckRunning <- 1 // Start loop
	return nil
}
