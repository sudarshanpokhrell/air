package validator

import "regexp"

var (
	EmailRX = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	UUIDRX  = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	SlugRX  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,62}$`)
	HashRX  = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`) // base64url sha256, no padding
)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func In(value string, list ...string) bool {
	for _, item := range list {
		if value == item {
			return true
		}
	}
	return false
}
