package main

import (
	"sync"
	"time"
)

// DummyWorker - Example for provider. Must be exported.
func DummyWorker(server string, waitGroup *sync.WaitGroup) {
	// Send signal about finish processing on end of function body, so thread may be destroyed.
	defer waitGroup.Done()
	// Collect execution time statistics for thread
	timeStart := time.Now()
	// Indicate entering in function
	ApplicationLogger.Output(3, "[DEBUG] Dummy provider ("+server+"): Thread started")

	/*
	 * Specify action via function calls (in this file)
	 */

	// Indicate leaving function, thread dies.
	ApplicationLogger.Output(3, "[DEBUG] Dummy provider ("+server+"): Thread finished in "+time.Since(timeStart).String())
}
