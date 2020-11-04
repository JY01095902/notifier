package notifier

type Attachment struct {
	FileName    string
	Content     []byte
	ContentType string
}

type Notifier interface {
	Notify(receivers []string, subject, content string, attachments ...Attachment) error
}
