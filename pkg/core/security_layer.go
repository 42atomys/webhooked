package core

import (
	"net/http"
)

type SecurityLayer interface {
	// Get the name of the security layer
	// Will be unique across all security layers
	Name() string
	// Is Secure is the interface called before the processing of a request
	// If the request is not secure, the request will be aborted with a
	// http.StatusUnauthorized error. If the request is secure, the function
	// should return true.
	IsSecure(headers http.Header, body []byte) bool
}
