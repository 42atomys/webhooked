package core

type Entry interface {
	// Get the name of the entry
	// Will be unique across all entries
	Name() string
	// Security layer used by this entry.
	// If nil, the entry is not secured.
	Security() SecurityLayer
}
