package unique

import (
	"fmt"
	"log"
	"sync"
	"testing"
)

func TestCreateUniqueIdsSimple(t *testing.T) {
	for i := 0; i < 100; i++ {
		newId := GetNewID()
		fmt.Println(newId)
	}
}
func TestCreateUniqueIdsConcurrent(t *testing.T) {
	mutex := &sync.Mutex{}
	generatedIds := map[string]bool{}
	generate := func(id string, wg *sync.WaitGroup) {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			newId := GetNewID()
			fmt.Println("Generator: "+id, newId)
			mutex.Lock()
			_, ok := generatedIds[newId]
			if ok {
				// This id is not unique!
				t.Fatal("Generated duplicate for id", newId)
			}
			generatedIds[newId] = true
			mutex.Unlock()
		}
	}

	var wg sync.WaitGroup
	wg.Add(4)
	go generate("1", &wg)
	go generate("2", &wg)
	go generate("3", &wg)
	go generate("4", &wg)
	wg.Wait()
	log.Println("All go routines finished")

}
