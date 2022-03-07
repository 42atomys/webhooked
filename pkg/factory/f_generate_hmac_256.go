package factory

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
)

type generateHMAC256Factory struct{ Factory }

func (*generateHMAC256Factory) Name() string {
	return "generate_hmac_256"
}

func (*generateHMAC256Factory) DefinedInpus() []*Var {
	return []*Var{
		{false, reflect.TypeOf(&InputConfig{}), "secret", &InputConfig{}},
		{false, reflect.TypeOf(&InputConfig{}), "payload", &InputConfig{}},
	}
}

func (*generateHMAC256Factory) DefinedOutputs() []*Var {
	return []*Var{
		{false, reflect.TypeOf(""), "value", ""},
	}
}

func (c *generateHMAC256Factory) Func() RunFunc {
	return func(factory *Factory, configRaw map[string]interface{}) error {
		payloadVar, ok := factory.Input("payload")
		if !ok {
			return fmt.Errorf("missing input payload")
		}

		secretVar, ok := factory.Input("secret")
		if !ok {
			return fmt.Errorf("missing input secret")
		}

		// Create a new HMAC by defining the hash type and the key (as byte array)
		h := hmac.New(sha256.New, []byte(secretVar.Value.(*InputConfig).First()))

		// Write Data to it
		h.Write([]byte(payloadVar.Value.(*InputConfig).First()))

		// Get result and encode as hexadecimal string
		sha := hex.EncodeToString(h.Sum(nil))
		factory.Output("value", sha)
		return nil
	}
}
