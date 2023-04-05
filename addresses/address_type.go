package addresses

type AddressVersion uint64

const (
	SIMPLE_PUBLIC_KEY AddressVersion = iota
	SIMPLE_HASH
)

func (e AddressVersion) String() string {
	switch e {
	case SIMPLE_PUBLIC_KEY:
		return "SIMPLE_PUBLIC_KEY"
	case SIMPLE_HASH:
		return "SIMPLE_HASH"
	default:
		return "Unknown Address Version"
	}
}
