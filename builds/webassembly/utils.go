package main

import (
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/federations/federation"
	"liberty-town/node/validator/validation"
	"syscall/js"
)

func federationValidate(f *federation.Federation, getMessage func() []byte, cb js.Value) (*validation.Validation, error) {
	return f.SignValidation(getMessage, func(data []byte) []byte {
		promise := cb.Invoke(string(data))
		solution, errs := webassembly_utils.Await(promise)
		if solution == nil || len(solution) != 1 || solution[0].IsNull() || len(errs) > 0 {
			return nil
		}
		return []byte(solution[0].String())
	})
}
