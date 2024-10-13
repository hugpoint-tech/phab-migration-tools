package operations

/// This package contains high-level logic of user-facing operations.
/// This package is necessary to decouple other packages from each other
/// and keep main clean.

import (
	_ "context"
	"fmt"
	"hugpoint.tech/freebsd/forge/bugzilla"
	types "hugpoint.tech/freebsd/forge/common/bugzilla"
	"hugpoint.tech/freebsd/forge/database"
	"hugpoint.tech/freebsd/forge/util"
	"sync"
	"time"
)

var (
	downloadWorkerCount = 20
	saverWorkerCount    = 10
	commentBatchSize    = 100
	bugBatchSize        = 1000
	pageSizeLimit       = 1000
	attachmentBatchSize = 100
)

type worker struct {
	id         int
	bugz       *bugzilla.Client
	db         *database.DB
	wg         *sync.WaitGroup
	errorCount int
	workerType string // Worker type (e.g., "comment-downloader")
	bugChan    chan types.Bug
}

func (w *worker) DownloadBugsWorker(limit int, bugChan chan types.Bug) error {
	offset := 0 // Start from the beginning

	for {
		// Start measuring download time for the current worker
		startDownload := time.Now()

		// Call the Client's DownloadBugs method with the current offset and limit
		bugs, err := w.bugz.DownloadBugs(offset, limit)
		if err != nil {
			return fmt.Errorf("worker %s failed to download bugs: %w", w.Id(), err)
		}

		// Measure download time for the current batch
		downloadDuration := time.Since(startDownload)
		fmt.Printf("Worker %s: Downloaded %d bugs in %v\n", w.Id(), len(bugs), downloadDuration)

		// Start measuring insert time (sending bugs to channel)
		startInsert := time.Now()

		// Send each bug to the channel for saving
		for _, bug := range bugs {
			bugChan <- bug // Send each bug to the channel
		}

		// Measure insert time for the current batch
		insertDuration := time.Since(startInsert)
		fmt.Printf("Worker %s: Inserted %d bugs into channel in %v\n", w.Id(), len(bugs), insertDuration)

		// If the number of bugs downloaded is less than the limit, exit the loop
		if len(bugs) < limit {
			break
		}

		// Increment the offset for the next batch of bugs
		offset += limit
	}

	fmt.Printf("Worker %s: Finished downloading all bugs\n", w.Id())
	return nil
}

func (w *worker) saveBugs(bugChan chan types.Bug) {
	for bug := range bugChan { // Loop over bugs sent to the channel
		err := w.db.InsertBug(bug) // Save each bug to the database
		if err != nil {
			fmt.Printf("Saver %d: failed to save bug ID %d: %s\n", w.id, bug.ID, err)
			w.errorCount++ // Track errors
		}
	}
	fmt.Printf("Saver %s: Finished saving all bugs\n", w.Id())
}

func DownloadBugzillaBugs(bugz *bugzilla.Client, db *database.DB) {
	// Initialize wait groups for downloaders and saver
	var downloaderWaitGroup sync.WaitGroup
	var saverWaitGroup sync.WaitGroup
	var totalErrors int

	bugChan := make(chan types.Bug, bugBatchSize) // Channel for bugs to be saved

	// Initialize downloader workers
	downloaders := make([]worker, downloadWorkerCount)
	for idx := 0; idx < downloadWorkerCount; idx++ {
		w := worker{
			id:         idx,
			bugz:       bugz,
			wg:         &downloaderWaitGroup,
			workerType: "bug-downloader",
		}
		downloaders[idx] = w
		downloaderWaitGroup.Add(1)

		// Start each worker in a separate goroutine
		go func(w worker) {
			defer downloaderWaitGroup.Done()
			err := w.DownloadBugsWorker(pageSizeLimit, bugChan)
			if err != nil {
				fmt.Printf("Worker %s encountered an error: %s\n", w.Id(), err)
			}
		}(w)
	}

	// Initialize saver workers
	savers := make([]worker, saverWorkerCount)
	for idx := 0; idx < saverWorkerCount; idx++ {
		w := worker{
			id:         idx,
			workerType: "bug-saver",
			db:         db,
			wg:         &saverWaitGroup,
		}
		savers[idx] = w
		saverWaitGroup.Add(1)

		// Start each saver worker in a separate goroutine
		go func(w worker) {
			defer saverWaitGroup.Done()
			w.saveBugs(bugChan) // Save bugs from the channel
		}(w)
	}

	downloaderWaitGroup.Wait() // Wait for all downloaders to finish
	close(bugChan)             // Close the bug channel to signal the saver
	saverWaitGroup.Wait()      // Wait for the saver to finish

	// Sum up errors from all workers
	for _, d := range downloaders {
		fmt.Printf("%s: finished with total errors: %d\n", d.Id(), d.errorCount)
		totalErrors += d.errorCount
	}

	// Sum up errors from all workers
	for _, s := range savers {
		fmt.Printf("%s: finished with total errors: %d\n", s.Id(), s.errorCount)
		totalErrors += s.errorCount
	}

	if totalErrors == 0 {
		fmt.Println("All bugs downloaded and saved successfully with no errors.")
	} else {
		fmt.Printf("There were errors during the download process. Total errors: %d\n", totalErrors)
	}
}

func (w *worker) downloadComment(in <-chan int, out chan<- types.Comment) {
	defer w.wg.Done()

	for id := range in {
		comments, err := w.bugz.DownloadBugComments(id)
		if err != nil {
			fmt.Printf("%d: failed to download comments for bug %d: %s\n", w.id, id, err)
			w.errorCount++
			continue
		}

		for _, c := range comments {
			out <- c
		}
		fmt.Printf("%d: downloaded comments for bug %d\n", w.id, id)
	}

	fmt.Printf("%d finished\n", w.id)
}

// Id method to return the formatted ID
func (w *worker) Id() string {
	return fmt.Sprintf("%s-%d", w.workerType, w.id)
}

func (w *worker) saveComments(in <-chan types.Comment) {
	defer w.wg.Done()

	buffer := make([]types.Comment, 0, commentBatchSize)

	for comment := range in {
		buffer = append(buffer, comment)

		if len(buffer) == commentBatchSize {
			err := w.db.InsertComment(buffer...)
			if err != nil {
				fmt.Printf("%d: failed to save comments %s\n", w.id, err)
				w.errorCount++
				continue
			}
			fmt.Printf("%d: saved %d comments\n", w.id, len(buffer))
			buffer = buffer[:0]
		}
	}

	err := w.db.InsertComment(buffer...)
	fmt.Printf("%d: saved %d comments", w.id, len(buffer))
	if err != nil {
		fmt.Printf("%d: failed to save comments %s\n", w.id, err)
		w.errorCount++
		return
	}

	fmt.Printf("%s finished\n", w.id)
}

func DownloadBugzillaComments(bugz *bugzilla.Client, db *database.DB) {
	// TODO verify that bugs were downloaded and warn user if 0 bugs were found in the database

	var downloaderWaitGroup sync.WaitGroup
	var saverWaitGroup sync.WaitGroup
	var totalErrors int

	idChan := make(chan int)
	commentChan := make(chan types.Comment, commentBatchSize)

	downloaders := make([]worker, 0, downloadWorkerCount)

	for idx := range cap(downloaders) {
		w := worker{
			id:         idx,
			bugz:       bugz,
			wg:         &downloaderWaitGroup,
			workerType: "comment-downloader",
		}
		fmt.Println(w.Id()) // This will print "comment-downloader-0", "comment-downloader-1", etc.
		downloaders = append(downloaders, w)
		downloaderWaitGroup.Add(1)
	}

	saver := worker{
		workerType: "comment-saver",
		db:         db,
		wg:         &saverWaitGroup,
	}

	go saver.saveComments(commentChan)

	for _, d := range downloaders {
		go d.downloadComment(idChan, commentChan)
	}

	err := db.ForEachBug(func(b types.Bug) error {
		idChan <- b.ID
		return nil
	})
	util.CheckFatal("failed to read bugs form the database", err)

	close(idChan)
	downloaderWaitGroup.Wait()
	close(commentChan)
	saverWaitGroup.Wait()

	// Sum up errors from all workers
	for _, d := range downloaders {
		fmt.Printf("%s: finished with total errors: %d\n", d.Id(), d.errorCount)
		totalErrors += d.errorCount
	}
	fmt.Printf("%s: finished with total errors: %d\n", saver.Id(), saver.errorCount)
	totalErrors += saver.errorCount
	fmt.Printf("Comments downloaded and saved with %d errors\n", totalErrors)

}

func (w *worker) downloadAttachment(in <-chan int, out chan<- types.Attachment) {
	defer w.wg.Done() // Mark worker as done when the function exits

	for id := range in { // Loop over the bug IDs received from the in channel
		attachments, err := w.bugz.DownloadBugAttachments(id) // Download attachments for the bug
		if err != nil {
			fmt.Printf("%s: failed to download attachments for bug %d: %s\n", w.id, id, err)
			w.errorCount++
			continue
		}

		// Send each downloaded attachment to the out channel
		for _, a := range attachments {
			out <- a
		}
		fmt.Printf("%d: downloaded attachments for bug %d\n", w.id, id)
	}

	fmt.Printf("%d finished\n", w.id) // Log that the worker is done
}

func (w *worker) saveAttachments(in <-chan types.Attachment) {
	defer w.wg.Done() // Mark worker as done when the function exits

	// Create a buffer to batch attachments for database insertion
	buffer := make([]types.Attachment, 0, attachmentBatchSize)

	for attachment := range in { // Loop over the attachments received from the in channel
		buffer = append(buffer, attachment)

		// If the buffer reaches the batch size, insert attachments into the database
		if len(buffer) == attachmentBatchSize {
			err := w.db.InsertAttachment(buffer...)
			if err != nil {
				fmt.Printf("%d: failed to save attachments %s\n", w.id, err)
				w.errorCount++
				continue
			}
			fmt.Printf("%d: saved %d attachments\n", w.id, len(buffer))
			buffer = buffer[:0] // Reset buffer after insertion
		}
	}

	// Insert any remaining attachments in the buffer
	err := w.db.InsertAttachment(buffer...)
	fmt.Printf("%d: saved %d attachments", w.id, len(buffer))
	if err != nil {
		fmt.Printf("%d: failed to save attachments %s\n", w.id, err)
		w.errorCount++
		return
	}

	fmt.Printf("%d finished\n", w.id) // Log that the worker is done
}

func DownloadBugzillaAttachments(bugz *bugzilla.Client, db *database.DB) {
	// Initialize WaitGroups for downloaders and the saver
	var downloaderWaitGroup sync.WaitGroup
	var saverWaitGroup sync.WaitGroup
	var totalErrors int

	// Channels for passing bug IDs and attachments between workers
	idChan := make(chan int)                                           // Channel to pass bug IDs to downloaders
	attachmentChan := make(chan types.Attachment, attachmentBatchSize) // Buffered channel to pass attachments to the saver

	// Create and launch attachment download workers
	downloaders := make([]worker, 0, downloadWorkerCount)
	for idx := range cap(downloaders) {
		w := worker{
			id:         idx,
			bugz:       bugz,                 // Bugzilla client for downloading attachments
			wg:         &downloaderWaitGroup, // Use WaitGroup to track completion
			workerType: "attachment-downloader",
		}
		fmt.Println(w.Id())
		downloaders = append(downloaders, w)
		downloaderWaitGroup.Add(1) // Add a worker to the WaitGroup
	}

	// Create and launch the attachment saver worker
	saver := worker{
		workerType: "attachment-saver", // Unique identifier for the saver
		db:         db,                 // Database connection for saving attachments
		wg:         &saverWaitGroup,    // Use WaitGroup to track completion
	}
	go saver.saveAttachments(attachmentChan) // Start saving attachments in a goroutine

	// Start each downloader worker in a separate goroutine
	for _, d := range downloaders {
		go d.downloadAttachment(idChan, attachmentChan)
	}

	// Fetch bug IDs from the database and send them to the downloaders
	err := db.ForEachBug(func(b types.Bug) error {
		idChan <- b.ID // Send bug ID to the idChan
		return nil     // Return nil to continue iteration
	})
	util.CheckFatal("failed to read bugs from the database", err)

	close(idChan)              // Close idChan to signal no more bug IDs will be sent
	downloaderWaitGroup.Wait() // Wait for all downloaders to finish
	close(attachmentChan)      // Close attachmentChan after downloaders are done
	saverWaitGroup.Wait()      // Wait for the saver to finish saving all attachments

	// Sum up errors from all workers
	for _, d := range downloaders {
		fmt.Printf("%s: finished with total errors: %d\n", d.Id(), d.errorCount)
		totalErrors += d.errorCount
	}
	fmt.Printf("%s: finished with total errors: %d\n", saver.Id(), saver.errorCount)
	totalErrors += saver.errorCount

	fmt.Printf("Attachments downloaded and saved with %d errors\n", totalErrors)
}
