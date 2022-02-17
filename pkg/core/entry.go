package core

type Entry interface {
	// Get the name of the entry
	// Will be unique across all entries
	Name() string
	// Security layer used by this entry.
	// If nil, the entry is not secured.
	Security() SecurityLayer
	// DomainsMatch needs to be able to match the domain of the request
	// when the function return true, the request will be identified as
	// a request for this entry.
	DomainMatches(host string) bool
}
