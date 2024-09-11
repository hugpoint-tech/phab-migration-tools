package bugz

import (
	. "hugpoint.tech/freebsd/forge/database"
	"log"
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

		// Get comments from the Bugzilla API using the bugID
		comments, err := worker.Client.DownloadBugzillaComments(bugID)
		if err != nil {
			log.Printf("Error downloading comments for bug %d: %v", bugID, err)
			continue
		}

		// Insert each comment into the database
		for _, comment := range comments {
			worker.DB.InsertComment(comment)
		}
	}
}

func (bc *BugzClient) downloadAllComments() {
	// Create a channel to send bugIDs for workers to process
	work := make(chan int64)

	// Create a slice to store worker objects
	var workersPool []CommentDownloadWorker

	// Create a pool of workers
	maxWorkers := 10
	for workerNum := 0; workerNum < maxWorkers; workerNum++ {
		worker := CommentDownloadWorker{
			BugIDChan: work,
			Client:    bc,    // Pass Bugzilla client reference
			DB:        bc.DB, // Pass database reference
		}
		// Add the worker to the pool
		workersPool = append(workersPool, worker)
	}

	// Launch workers in separate goroutines
	for _, worker := range workersPool {
		go worker.downloadComment() // Start downloading comments in goroutines
	}

	// Fetch all bug IDs that need comment downloads from the database
	bugIDs, _ := bc.DB.GetAllBugIDs()

	// Send each bugID to the workers through the work channel
	for _, bugID := range bugIDs {
		work <- bugID
	}

	// Close the work channel to signal no more work
	close(work)
}

type AttachmentDownloadWorker struct {
	ID     int
	Work   <-chan int
	Client *BugzClient
}

func (worker *AttachmentDownloadWorker) DownloadAttachments() {

}
