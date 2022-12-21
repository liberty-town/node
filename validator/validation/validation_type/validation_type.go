package validation_type

type ValidatorChallengeType uint64

const (
	VALIDATOR_CHALLENGE_NO_CAPTCHA ValidatorChallengeType = iota
	VALIDATOR_CHALLENGE_HCAPTCHA
	VALIDATOR_CHALLENGE_CUSTOM
)
