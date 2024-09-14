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

type worker struct {
	id   string
	bugz *bugzilla.Client
	db   *database.DB
	wg   *sync.WaitGroup
}

func (w *worker) downloadComment(in <-chan int, out chan<- types.Comment) {
	defer w.wg.Done()

	for id := range in {
		comments, err := w.bugz.DownloadBugComments(id)
		if err != nil {
			fmt.Printf("%s: failed to download comments for bug %d: %s\n", w.id, id, err)
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
		return
	}

	fmt.Printf("%s finished\n", w.id)
}

func DownloadBugzillaComments(bugz *bugzilla.Client, db *database.DB) {
	// TODO verify that bugs were downloaded and warn user if 0 bugs were found in the database

	var downloaderWaitGroup sync.WaitGroup
	var saverWaitGroup sync.WaitGroup

	idChan := make(chan int)
	commentChan := make(chan types.Comment, COMMENT_BATCH_SIZE)

	downloaders := make([]worker, 0, DOWNLOAD_WORKER_COUNT)

	for idx := range cap(downloaders) {
		w := worker{
			id:   fmt.Sprintf("commend-downloader-%d", idx),
			bugz: bugz,
			wg:   &downloaderWaitGroup,
		}
		downloaders = append(downloaders, w)
		downloaderWaitGroup.Add(1)
	}

	saver := worker{
		id: fmt.Sprintf("commend-saver"),
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

	fmt.Println("All comments downloaded successfully.")
}

func DownloadBugzillaBugs(bugz *bugzilla.Client, db *database.DB) {
	bugs, err := bugz.DownloadBugzillaBugs()
	util.CheckFatal("failed to download bugzilla bugs", err)

	for _, bug := range bugs {
		err = db.InsertBug(bug)
		util.CheckFatal("failed to insert bug into the database", err)

	}
}
