package bugz

import (
	"fmt"
	. "hugpoint.tech/freebsd/forge/database"
	"sync"
)

type CommentDownloadWorker struct {
	ID        int
	BugIDChan <-chan int64
	Client    *BugzClient
	DB        *DB
}

func (worker *CommentDownloadWorker) downloadComment() {
	for {
		// Receive bugID from the work channel
		bugID, continueWork := <-worker.BugIDChan

		if !continueWork {
			// If no more work, exit the loop
			break
		}

		// Download and insert comments directly
		_, err := worker.Client.DownloadBugComments(bugID)
		if err != nil {
			fmt.Printf("Error downloading comments for bug %d: %v", bugID, err)
			continue
		}
	}
}

func (bc *BugzClient) DownloadAllComments(ids []int64) error {

	// Create channel to send bugIDs for workers to process
	work := make(chan int64)

	// Create a WaitGroup to track the worker completion
	var wg sync.WaitGroup

	maxWorkers := 20
	for workerNum := 0; workerNum < maxWorkers; workerNum++ {
		worker := CommentDownloadWorker{
			BugIDChan: work,
			Client:    bc, // Pass Bugzilla client reference
		}

		// Increment the WaitGroup counter
		wg.Add(1)

		// Launch workers in separate goroutines
		go func(w CommentDownloadWorker) {
			defer wg.Done()     // Ensure WaitGroup is decremented when worker finishes
			w.downloadComment() // Start downloading comments in goroutines
		}(worker)
	}

	// Send each bugID to the workers through the work channel
	for _, bugID := range ids {
		work <- bugID
	}

	// Close the work channel to signal no more work
	close(work)

	// Wait for all workers to finish
	wg.Wait()

	fmt.Println("All comments downloaded successfully.")
	return nil
}

type AttachmentDownloadWorker struct {
	ID     int
	Work   <-chan int
	Client *BugzClient
}

func (worker *AttachmentDownloadWorker) DownloadAttachments() {

}
