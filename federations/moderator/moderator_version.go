package moderator

type ModeratorVersion uint64

const (
	MODERATOR_PANDORA ModeratorVersion = iota
)

func (e ModeratorVersion) String() string {
	switch e {
	case MODERATOR_PANDORA:
		return "MODERATOR_PANDORA"
	default:
		return "Unknown Moderator Version"
	}
}
