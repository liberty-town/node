package chat_message

func SortKeys(sender, receiver string) (string, string, bool) {
	if sender > receiver {
		return receiver, sender, true
	}
	return sender, receiver, false
}
