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
	running    bool
	processors []Processor
}

//------------------------------------------------------------------
// ~ CONSTRUCTORS
//------------------------------------------------------------------

func NewQueue() *Queue {
	return &Queue{}
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// AddProcessor Add a processor to the queue. Do not add multiple processors for the same task!
func (q *Queue) AddProcessor(processor Processor) {
	q.processors = append(q.processors, processor)
	log.Println("Added Processor to queue. New length: ", len(q.processors))
	for _, p := range q.processors {
		fmt.Println("\t", p.GetId())
	}
}

func (q *Queue) GetProcessors() []Processor {
	return q.processors
}

func (q *Queue) IsRunning() bool {
	return q.running
}

// ScheduleStart continuously tries to run all available processors after interval until ScheduleStop has been called.
// As long as there is only one processor per kind (Status, Payment, ...), there should be no race conditions
func (q *Queue) Start() error {
	log.Println("Queue: Schedule Start")
	if q.IsRunning() {
		return errors.New("Did not start Queue. It's already running!")
	}
	q.running = true
	waitGroup := &sync.WaitGroup{}
	for _, proc := range q.processors {
		waitGroup.Add(1)
		go schedule(proc, waitGroup)
	}
	waitGroup.Wait()
	q.running = false
	fmt.Println("")
	fmt.Println("*****------------------------------------****")
	for _, proc := range q.processors {
		proc.Report()
	}
	fmt.Println("*****------------------------------------****")
	return nil
}

func (q *Queue) Stop() {
	log.Println("Queue: Schedule Stop")
	for _, proc := range q.processors {
		proc.Stop()
	}
}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

func schedule(proc Processor, waitGroup *sync.WaitGroup) {
	chanStart := make(chan int)
	chanStop := make(chan int) // this is called by ScheduleStop()

	go func() {
		for {
			select {
			case <-proc.GetChanExit():
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
		for {
			select {
			case <-chanStart:
				log.Println("Started Processor:", proc.GetId())
				runProcessor(proc)
			}
		}
	}()

	go func() {
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
	log.Println("Exiting schedule:", proc.GetId())

}

func runProcessor(processor Processor) error {
	processor.SetStartTimeProcessing(time.Now().UnixNano())
	chanDone := make(chan int)
	chanReady := make(chan interface{})
	chanCheckRunning := make(chan int)
	chanGoCheckRunning := make(chan int)

	iter, err := processor.Find(processor.GetQuery(), processor.GetPersistor())
	if err != nil {
		log.Println(err)
		processor.GetChanExit() <- 1 // this will be received in schedule() and stop the processor
		return err
	}
	go func() {
		for {
			select {
			case data := <-chanReady:
				//log.Println("** chanReady", processor.GetId())
				f := func(data interface{}) {
					err := processor.Process(data)
					if err != nil {
						log.Println(err)
					}
					processor.IncCountProcessed()
					processor.DecRunningJobs()
					processor.GetWaitGroup().Done()
				}

				go f(data)
				chanCheckRunning <- 1

			}
		}
	}()

	go func() {
		for {
			select {
			case <-chanDone:
				// Wait for jobs to finsish and exit
				//log.Println("** chanDone", processor.GetId())
				processor.GetWaitGroup().Wait()
				processor.SetWaitGroupFinished(true)
				processor.GetChanExit() <- 1
				processor.SetEndTimeProcessing(time.Now().UnixNano())

			case <-chanGoCheckRunning:
				//log.Println("** goCheckRunning", processor.GetId())
				chanCheckRunning <- 1
			}
		}

	}()
	go func() {
		run := true
	Loop:
		for run {
			select {
			case <-chanCheckRunning:
				if processor.GetStop() {
					chanDone <- 1
					run = false
					break Loop
				}
				//log.Println("** chanCheckRunning", processor.GetId())
				if processor.GetJobsStarted() >= processor.GetJobsAssigned() { // We are done
					//	log.Println("** goChanDone :: JobsAssignedProcessed", processor.GetId())
					chanDone <- 1
					break Loop
				} else if processor.GetRunningJobs() >= processor.GetMaxConcurrency() { // Wait for better times

					//	log.Println("** wait", processor.GetId())
					chanGoCheckRunning <- 1
				} else {
					if processor.GetJobsStarted() < processor.GetJobsAssigned() {
						data, err := iter()

						if err != nil {
							log.Println("Error: Could not get data", err)
							chanDone <- 1
							break
						}
						if data != nil {
							processor.GetWaitGroup().Add(1)
							processor.IncJobsStarted()
							processor.IncRunningJobs() // We count here also, because we cannot access current number of jobs through waitgroup
							chanReady <- data
						} else {
							log.Println("data is nil", err)
							chanDone <- 1
							break Loop
						}
					}
				}

			}
		}
	}()
	chanCheckRunning <- 1 // Start loop
	return nil
}
