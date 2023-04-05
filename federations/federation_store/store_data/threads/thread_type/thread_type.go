package thread_type

type ThreadType byte

const (
	THREAD_TEXT ThreadType = iota
	THREAD_LINK
	THREAD_IMAGE
)
