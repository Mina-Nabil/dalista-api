package server

import (
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var registerValidatorsOnce sync.Once

// RegisterCustomValidators registers app-specific struct tags on Gin's
// validator. Call once during process startup (and from handler tests'
// TestMain). No custom tags are registered yet — feature phases add theirs here
// (e.g. a phone-number validator for the OTP sign-in flow in Phase 1).
func RegisterCustomValidators() {
	registerValidatorsOnce.Do(func() {
		if _, ok := binding.Validator.Engine().(*validator.Validate); !ok {
			return
		}
		// Future: engine.RegisterValidation("dalista_phone", ...)
	})
}
