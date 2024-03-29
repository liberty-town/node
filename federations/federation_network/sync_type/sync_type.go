package sync_type

import (
	"errors"
)

type SyncVersion uint64

const (
	SYNC_ACCOUNTS SyncVersion = iota
	SYNC_LISTINGS
	SYNC_ACCOUNTS_SUMMARIES
	SYNC_LISTINGS_SUMMARIES
	SYNC_MESSAGES
	SYNC_REVIEWS
	SYNC_THREADS
	SYNC_COMMENTS
	SYNC_POLLS
)

func (t SyncVersion) GetStringStoreName() (string, error) {
	switch t {
	case SYNC_ACCOUNTS:
		return "accounts", nil
	case SYNC_LISTINGS:
		return "listings", nil
	case SYNC_ACCOUNTS_SUMMARIES:
		return "accounts_summaries", nil
	case SYNC_LISTINGS_SUMMARIES:
		return "listings_summaries", nil
	case SYNC_MESSAGES:
		return "messages", nil
	case SYNC_REVIEWS:
		return "reviews", nil
	case SYNC_THREADS:
		return "threads", nil
	case SYNC_COMMENTS:
		return "comments", nil
	case SYNC_POLLS:
		return "polls", nil
	default:
		return "", errors.New("invalid sync type")
	}
}
