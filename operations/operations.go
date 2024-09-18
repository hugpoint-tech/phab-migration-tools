package operations

/// This package contains high-level logic of user-facing operations.
/// This package is necessary to decouple other packages from each other
/// and keep main clean.

import (
	"fmt"
	"hugpoint.tech/freebsd/forge/bugzilla"
	types "hugpoint.tech/freebsd/forge/common/bugzilla"
	"hugpoint.tech/freebsd/forge/database"
	"hugpoint.tech/freebsd/forge/util"
	"sync"
)

var DOWNLOAD_WORKER_COUNT = 20

const COMMENT_BATCH_SIZE = 100

const attachmentBatchSize = 100

type worker struct {
	id         string
	bugz       *bugzilla.Client
	db         *database.DB
	wg         *sync.WaitGroup
	errorCount int
}

func (w *worker) downloadComment(in <-chan int, out chan<- types.Comment) {
	defer w.wg.Done()

	for id := range in {
		comments, err := w.bugz.DownloadBugComments(id)
		if err != nil {
			fmt.Printf("%s: failed to download comments for bug %d: %s\n", w.id, id, err)
			w.errorCount++
			fmt.Printf("%s: total errors: %d\n", w.id, w.errorCount) // Show errors while downloading
			continue
		}

		for _, c := range comments {
			out <- c
		}
		fmt.Printf("%s: downloaded comments for bug %d\n", w.id, id)
	}

	fmt.Printf("%s finished\n", w.id)
}

func (w *worker) saveComments(in <-chan types.Comment) {
	defer w.wg.Done()

	buffer := make([]types.Comment, 0, COMMENT_BATCH_SIZE)

	for comment := range in {
		buffer = append(buffer, comment)

		if len(buffer) == COMMENT_BATCH_SIZE {
			err := w.db.InsertComment(buffer...)
			if err != nil {
				fmt.Printf("%s: failed to save comments %s\n", w.id, err)
				w.errorCount++
				fmt.Printf("%s: total save errors: %d\n", w.id, w.errorCount) // Show errors while saving
				continue
			}
			fmt.Printf("%s: saved %d comments\n", w.id, len(buffer))
			buffer = buffer[:0]
		}
	}

	err := w.db.InsertComment(buffer...)
	fmt.Printf("%s: saved %d comments", w.id, len(buffer))
	if err != nil {
		fmt.Printf("%s: failed to save comments %s\n", w.id, err)
		w.errorCount++
		fmt.Printf("%s: total save errors: %d\n", w.id, w.errorCount)
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
	commentChan := make(chan types.Comment, COMMENT_BATCH_SIZE)

	downloaders := make([]worker, 0, DOWNLOAD_WORKER_COUNT)

	for idx := range cap(downloaders) {
		w := worker{
			id:   fmt.Sprintf("comment-downloader-%d", idx),
			bugz: bugz,
			wg:   &downloaderWaitGroup,
		}
		downloaders = append(downloaders, w)
		downloaderWaitGroup.Add(1)
	}

	saver := worker{
		id: fmt.Sprintf("comment-saver"),
		db: db,
		wg: &saverWaitGroup,
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
		fmt.Printf("%s: finished with total errors: %d\n", d.id, d.errorCount)
		totalErrors += d.errorCount
	}
	fmt.Printf("%s: finished with total errors: %d\n", saver.id, saver.errorCount)
	totalErrors += saver.errorCount
	fmt.Printf("Comments downloaded and saved with %d errors\n", totalErrors)

}

func DownloadBugzillaBugs(bugz *bugzilla.Client, db *database.DB) {
	bugs, err := bugz.DownloadBugzillaBugs()
	util.CheckFatal("failed to download bugzilla bugs", err)

	for _, bug := range bugs {
		err = db.InsertBug(bug)
		util.CheckFatal("failed to insert bug into the database", err)

	}
}

func (w *worker) downloadAttachment(in <-chan int, out chan<- types.Attachment) {
	defer w.wg.Done() // Mark worker as done when the function exits

	for id := range in { // Loop over the bug IDs received from the in channel
		attachments, err := w.bugz.DownloadBugAttachments(id) // Download attachments for the bug
		if err != nil {
			fmt.Printf("%s: failed to download attachments for bug %d: %s\n", w.id, id, err)
			w.errorCount++
			fmt.Printf("%s: total errors: %d\n", w.id, w.errorCount) // Show errors while downloading
			continue
		}

		// Send each downloaded attachment to the out channel
		for _, a := range attachments {
			out <- a
		}
		fmt.Printf("%s: downloaded attachments for bug %d\n", w.id, id)
	}

	fmt.Printf("%s finished\n", w.id) // Log that the worker is done
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
				fmt.Printf("%s: failed to save attachments %s\n", w.id, err)
				w.errorCount++
				fmt.Printf("%s: total save errors: %d\n", w.id, w.errorCount) // Show errors while saving
				continue
			}
			fmt.Printf("%s: saved %d attachments\n", w.id, len(buffer))
			buffer = buffer[:0] // Reset buffer after insertion
		}
	}

	// Insert any remaining attachments in the buffer
	err := w.db.InsertAttachment(buffer...)
	fmt.Printf("%s: saved %d attachments", w.id, len(buffer))
	if err != nil {
		fmt.Printf("%s: failed to save attachments %s\n", w.id, err)
		w.errorCount++
		fmt.Printf("%s: total save errors: %d\n", w.id, w.errorCount) // Show errors while saving
		return
	}

	fmt.Printf("%s finished\n", w.id) // Log that the worker is done
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
	downloaders := make([]worker, 0, DOWNLOAD_WORKER_COUNT)
	for idx := range cap(downloaders) {
		w := worker{
			id:   fmt.Sprintf("attachment-downloader-%d", idx),
			bugz: bugz,                 // Bugzilla client for downloading attachments
			wg:   &downloaderWaitGroup, // Use WaitGroup to track completion
		}
		downloaders = append(downloaders, w)
		downloaderWaitGroup.Add(1) // Add a worker to the WaitGroup
	}

	// Create and launch the attachment saver worker
	saver := worker{
		id: "attachment-saver", // Unique identifier for the saver
		db: db,                 // Database connection for saving attachments
		wg: &saverWaitGroup,    // Use WaitGroup to track completion
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
		fmt.Printf("%s: finished with total errors: %d\n", d.id, d.errorCount)
		totalErrors += d.errorCount
	}
	fmt.Printf("%s: finished with total errors: %d\n", saver.id, saver.errorCount)
	totalErrors += saver.errorCount
	fmt.Printf("Attachments downloaded and saved with %d errors\n", totalErrors)

}
