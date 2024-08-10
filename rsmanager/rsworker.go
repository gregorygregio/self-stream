package rsmanager

import (
	"fmt"
	"time"
)

func StartResourceWorker() {
	fmt.Println("Starting Resources worker")

	workerTicker := time.NewTicker(30 * time.Second)

	for range workerTicker.C {
		onWorkerTick()
	}
}

func onWorkerTick() {
	fmt.Println("Tick")
	resourcesToLoad, err := GetResourcesToLoad()
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
	}

	fmt.Printf("Number of resources to Load: %v\n", len(resourcesToLoad))

	for i := 0; i < len(resourcesToLoad); i++ {
		go processResource(&resourcesToLoad[i])
	}
}

func processResource(resource *ResourceInfo) {
	fmt.Printf("Working on resource %v\n", resource.RawFileName)
	err := LoadResource(resource)
	if err != nil {
		fmt.Printf("Error to load resource: %v\n", err.Error())
	}
}
