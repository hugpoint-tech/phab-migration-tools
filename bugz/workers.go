package bugz

type CommentDownloadWorker struct {
	ID        int
	BugIDChan <-chan int
	Client    *BugzClient
}

func (worker *CommentDownloadWorker) DownloadComments() {

}

type AttachmentDownloadWorker struct {
	ID     int
	Work   <-chan int
	Client *BugzClient
}

func (worker *AttachmentDownloadWorker) DownloadAttachments() {

}
